package dbgp

import (
	"encoding/xml"
	"strings"

	"golang.org/x/net/html/charset"
)

// ProtocolInit represents the initial protocol message sent by Xdebug
type ProtocolInit struct {
	XMLName    xml.Name `xml:"init"`
	AppID      string   `xml:"appid,attr"`
	IDEKey     string   `xml:"idekey,attr"`
	Session    string   `xml:"session,attr"`
	Thread     string   `xml:"thread,attr"`
	Parent     string   `xml:"parent,attr"`
	Language   string   `xml:"language,attr"`
	ProtocolV  string   `xml:"protocol_version,attr"`
	FileURI    string   `xml:"fileuri,attr"`
	EngineID   string   `xml:"engine>name"`
	EngineVer  string   `xml:"engine>version"`
}

// ProtocolResponse represents a response message from Xdebug
type ProtocolResponse struct {
	XMLName       xml.Name             `xml:"response"`
	Command       string               `xml:"command,attr"`
	Status        string               `xml:"status,attr"`
	Reason        string               `xml:"reason,attr"`
	TransactionID string               `xml:"transaction_id,attr"`
	ID            string               `xml:"id,attr"`
	Error         *ProtocolError       `xml:"error"`
	Breakpoints   []ProtocolBreakpoint `xml:"breakpoint"`
	Properties    []ProtocolProperty   `xml:"property"`
	Contexts      []ProtocolContext    `xml:"context"`
	Message       []ProtocolMessage    `xml:"message"`
	Stack         []ProtocolStack      `xml:"stack"`
	Filename      string               `xml:"filename,attr"`
	Lineno        string               `xml:"lineno,attr"`
	Source        string               `xml:",chardata"`
	// feature_get response fields
	FeatureName string `xml:"feature_name,attr"`
	Supported   string `xml:"supported,attr"`
}

// ProtocolError represents an error in a response
type ProtocolError struct {
	Code    string `xml:"code,attr"`
	Message string `xml:"message"`
}

// ProtocolProperty represents a variable/property in a response
type ProtocolProperty struct {
	XMLName      xml.Name            `xml:"property"`
	Name         string              `xml:"name,attr"`
	FullName     string              `xml:"fullname,attr"`
	Type         string              `xml:"type,attr"`
	ClassType    string              `xml:"classname,attr"`
	Facet        string              `xml:"facet,attr"`
	Size         string              `xml:"size,attr"`
	Page         string              `xml:"page,attr"`
	PageSize     string              `xml:"pagesize,attr"`
	Address      string              `xml:"address,attr"`
	Key          string              `xml:"key,attr"`
	NumChildren  string              `xml:"numchildren,attr"`
	Encoding     string              `xml:"encoding,attr"`
	Value        string              `xml:",chardata"`
	Children     []ProtocolProperty  `xml:"property"`
}

// ProtocolBreakpoint represents a breakpoint in a response
type ProtocolBreakpoint struct {
	XMLName       xml.Name `xml:"breakpoint"`
	ID            string   `xml:"id,attr"`
	Type          string   `xml:"type,attr"`
	State         string   `xml:"state,attr"`
	Filename      string   `xml:"filename,attr"`
	Lineno        string   `xml:"lineno,attr"`
	Function      string   `xml:"function,attr"`
	Exception     string   `xml:"exception,attr"`
	HitValue      string   `xml:"hit_value,attr"`
	HitCondition  string   `xml:"hit_condition,attr"`
	HitCount      string   `xml:"hit_count,attr"`
	Expression    string   `xml:"expression"`
}

// ProtocolContext represents a context (variable scope) in a response
type ProtocolContext struct {
	XMLName xml.Name `xml:"context"`
	Name    string   `xml:"name,attr"`
	ID      string   `xml:"id,attr"`
}

// ProtocolMessage represents a message element in a response
type ProtocolMessage struct {
	XMLName  xml.Name `xml:"message"`
	Filename string   `xml:"filename,attr"`
	Lineno   string   `xml:"lineno,attr"`
}

// ProtocolStack represents a stack frame in a stack trace response
type ProtocolStack struct {
	XMLName  xml.Name `xml:"stack"`
	Where    string   `xml:"where,attr"`
	Level    string   `xml:"level,attr"`
	Type     string   `xml:"type,attr"`
	Filename string   `xml:"filename,attr"`
	Lineno   string   `xml:"lineno,attr"`
}

// CreateProtocolFromXML parses XML data and returns appropriate protocol structure
func CreateProtocolFromXML(xmlData string) (interface{}, error) {
	xmlData = strings.TrimSpace(xmlData)

	// Try to parse as init message first
	if strings.HasPrefix(xmlData, "<?xml") && strings.Contains(xmlData, "<init ") {
		var init ProtocolInit
		decoder := xml.NewDecoder(strings.NewReader(xmlData))
		decoder.CharsetReader = charset.NewReaderLabel
		err := decoder.Decode(&init)
		if err != nil {
			return nil, err
		}
		return &init, nil
	}

	// Try to parse as response message
	if strings.Contains(xmlData, "<response ") {
		var response ProtocolResponse
		decoder := xml.NewDecoder(strings.NewReader(xmlData))
		decoder.CharsetReader = charset.NewReaderLabel
		err := decoder.Decode(&response)
		if err != nil {
			return nil, err
		}
		return &response, nil
	}

	// If neither format matches, try response as fallback
	var response ProtocolResponse
	decoder := xml.NewDecoder(strings.NewReader(xmlData))
	decoder.CharsetReader = charset.NewReaderLabel
	err := decoder.Decode(&response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

// HasError checks if the response contains an error
func (r *ProtocolResponse) HasError() bool {
	return r.Error != nil && r.Error.Code != ""
}

// GetErrorMessage returns formatted error message
func (r *ProtocolResponse) GetErrorMessage() string {
	if !r.HasError() {
		return ""
	}
	if r.Error.Message != "" {
		return r.Error.Message
	}
	return "Error code: " + r.Error.Code
}
