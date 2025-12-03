# Tasks: Implement Xdebug CLI

## Phase 1: Foundation (cfg package)

- [x] 1.1 Create `internal/cfg/config.go` with Version and CLIParameter struct
- [x] 1.2 Add tests for cfg package

## Phase 2: DBGp Protocol Layer

### 2.1 Server
- [x] 2.1.1 Create `internal/dbgp/server.go` with Server struct
- [x] 2.1.2 Implement NewServer(address, port)
- [x] 2.1.3 Implement Listen() to bind TCP socket
- [x] 2.1.4 Implement Accept(handler) loop
- [x] 2.1.5 Add server tests

### 2.2 Connection
- [x] 2.2.1 Create `internal/dbgp/connection.go` with Connection struct
- [x] 2.2.2 Implement NewConnection(net.Conn)
- [x] 2.2.3 Implement ReadMessage() with DBGp framing (size\0xml\0)
- [x] 2.2.4 Implement SendMessage(string) with null terminator
- [x] 2.2.5 Implement GetResponse() -> *ProtocolResponse
- [x] 2.2.6 Implement Close()
- [x] 2.2.7 Add connection tests

### 2.3 Protocol
- [x] 2.3.1 Create `internal/dbgp/protocol.go` with type definitions
- [x] 2.3.2 Define ProtocolInit struct with XML tags
- [x] 2.3.3 Define ProtocolResponse struct with XML tags
- [x] 2.3.4 Define ProtocolProperty, ProtocolBreakpoint, etc.
- [x] 2.3.5 Implement CreateProtocolFromXML(string) parser
- [x] 2.3.6 Implement HasError() method on ProtocolResponse
- [x] 2.3.7 Add protocol parsing tests

### 2.4 Session
- [x] 2.4.1 Create `internal/dbgp/session.go` with Session struct
- [x] 2.4.2 Define SessionStateType enum
- [x] 2.4.3 Implement NewSession()
- [x] 2.4.4 Implement NextTransactionID()
- [x] 2.4.5 Implement AddCommand() and GetLastCommand()
- [x] 2.4.6 Implement SetTargetFiles(rootFile)
- [x] 2.4.7 Add session tests

### 2.5 Client
- [x] 2.5.1 Create `internal/dbgp/client.go` with Client struct
- [x] 2.5.2 Implement NewClient(conn)
- [x] 2.5.3 Implement Init() - read init protocol, set session state
- [x] 2.5.4 Implement Run() - send run command
- [x] 2.5.5 Implement Step() - send step_into command
- [x] 2.5.6 Implement Next() - send step_over command
- [x] 2.5.7 Implement Finish() - send stop command
- [x] 2.5.8 Implement SetBreakpoint(file, line, condition)
- [x] 2.5.9 Implement SetBreakpointToCall(funcName)
- [x] 2.5.10 Implement SetExceptionBreakpoint()
- [x] 2.5.11 Implement GetBreakpointList()
- [x] 2.5.12 Implement GetProperty(name)
- [x] 2.5.13 Implement GetContext(contextID)
- [x] 2.5.14 Implement GetContextNames()
- [x] 2.5.15 Add client tests

## Phase 3: View Layer

### 3.1 Core View
- [x] 3.1.1 Create `internal/view/view.go` with View struct
- [x] 3.1.2 Implement NewView() with stdin reader
- [x] 3.1.3 Implement Print(s), PrintLn(s), PrintErrorLn(s)
- [x] 3.1.4 Implement PrintInputPrefix() - "(xdbg) " prompt
- [x] 3.1.5 Implement GetInputLine()
- [x] 3.1.6 Implement PrintApplicationInformation(version, host, port)

### 3.2 Source Display
- [x] 3.2.1 Create `internal/view/source.go` with SourceFileCache
- [x] 3.2.2 Implement NewSourceFileCache()
- [x] 3.2.3 Implement cacheFile(path)
- [x] 3.2.4 Implement getLines(path, begin, length)
- [x] 3.2.5 Implement PrintSourceLn(fileURI, line, length) on View
- [x] 3.2.6 Implement PrintSourceChangeLn(filename)

### 3.3 Help Messages
- [x] 3.3.1 Create `internal/view/help.go`
- [x] 3.3.2 Implement ShowHelpMessage() - main command list
- [x] 3.3.3 Implement ShowInfoHelpMessage()
- [x] 3.3.4 Implement ShowStepHelpMessage()
- [x] 3.3.5 Implement ShowBreakpointHelpMessage()
- [x] 3.3.6 Implement ShowPrintHelpMessage()
- [x] 3.3.7 Implement ShowContextHelpMessage()

### 3.4 Display Formatting
- [x] 3.4.1 Create `internal/view/display.go`
- [x] 3.4.2 Implement ShowInfoBreakpoints([]ProtocolBreakpoint)
- [x] 3.4.3 Implement PrintPropertyListWithDetails(scope, []Property)
- [x] 3.4.4 Implement printProperty() recursive helper

## Phase 4: CLI Commands

### 4.1 Root Command Updates
- [x] 4.1.1 Update `internal/cli/root.go` - add --host/-l, --port/-p flags
- [x] 4.1.2 Add global CLIArgs variable
- [x] 4.1.3 Update version command to use cfg.Version

### 4.2 Listen Command
- [x] 4.2.1 Create `internal/cli/listen.go` with listenCmd
- [x] 4.2.2 Implement runListeningCmd() - start server
- [x] 4.2.3 Implement listenAccept(conn) - connection handler
- [x] 4.2.4 Implement REPL loop with command dispatch
- [x] 4.2.5 Implement all REPL commands (run, step, next, break, print, context, list, info, finish, help, quit)
- [x] 4.2.6 Implement updateState() - handle protocol responses

### 4.3 Connection Command
- [x] 4.3.1 Create `internal/cli/connection.go` with connectionCmd
- [x] 4.3.2 Implement connection (no args) - show status
- [x] 4.3.3 Implement connection isAlive - check and exit code
- [x] 4.3.4 Implement connection kill - terminate session
- [x] 4.3.5 Add global activeSession state management

## Phase 5: Cleanup and Migration

- [x] 5.1 Remove `internal/progress/` directory
- [x] 5.2 Remove `internal/cli/preview.go` and `preview_test.go`
- [x] 5.3 Update go.mod if new dependencies needed (golang.org/x/net/html/charset)
- [x] 5.4 Update README.md with new command documentation
- [x] 5.5 Update CLAUDE.md with new commands

## Phase 6: Verification

- [x] 6.1 Run `go build -o xdebug-cli ./cmd/xdebug-cli`
- [x] 6.2 Run `go test ./...` - all tests pass
- [x] 6.3 Test `xdebug-cli listen -p 9003` starts server
- [x] 6.4 Test with real PHP + Xdebug connection
- [x] 6.5 Run `./install.sh` to install updated binary
- [x] 6.6 Verify `xdebug-cli version` shows correct version
