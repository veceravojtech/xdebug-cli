#!/bin/bash
# Installation script for xdebug-cli
set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

INSTALL_DIR="${HOME}/.local/bin"
BINARY_NAME="xdebug-cli"
VERSION="1.0.0"
BUILD_TIME=$(date -u '+%Y-%m-%d_%H:%M:%S')

echo -e "${GREEN}=== Xdebug CLI Installation ===${NC}"

# Check Go
if ! command -v go &> /dev/null; then
    echo -e "${RED}Error: Go is not installed.${NC}"
    exit 1
fi

GO_VERSION=$(go version | awk '{print $3}')
echo -e "Found Go: ${GREEN}${GO_VERSION}${NC}"

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

echo -e "${YELLOW}Downloading dependencies...${NC}"
go mod download

echo -e "${YELLOW}Building ${BINARY_NAME} v${VERSION}...${NC}"
LDFLAGS="-X github.com/console/xdebug-cli/internal/cli.Version=${VERSION} -X github.com/console/xdebug-cli/internal/cli.BuildTime=${BUILD_TIME}"
go build -ldflags "${LDFLAGS}" -o "${BINARY_NAME}" ./cmd/xdebug-cli

if [ ! -f "${BINARY_NAME}" ]; then
    echo -e "${RED}Error: Build failed${NC}"
    exit 1
fi

if [ ! -d "$INSTALL_DIR" ]; then
    mkdir -p "$INSTALL_DIR"
fi

mv "${BINARY_NAME}" "${INSTALL_DIR}/"
chmod +x "${INSTALL_DIR}/${BINARY_NAME}"

echo -e "${GREEN}Installation complete!${NC}"

if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
    echo -e "${YELLOW}Warning: ${INSTALL_DIR} is not in your PATH${NC}"
    echo 'Add: export PATH="$HOME/.local/bin:$PATH"'
fi

echo -e "${GREEN}Run 'xdebug-cli version' to verify${NC}"
