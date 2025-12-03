# Tasks: Rename CLI to xdebug-cli

## 1. Core Module Rename
- [x] 1.1 Update `go.mod` module name from `github.com/console/source-cli` to `github.com/console/xdebug-cli`
- [x] 1.2 Rename directory `cmd/source-cli/` to `cmd/xdebug-cli/`
- [x] 1.3 Update import in `cmd/xdebug-cli/main.go`

## 2. Internal Package Updates
- [x] 2.1 Update `internal/cli/root.go` - Use field and Long description
- [x] 2.2 Update `internal/cli/preview.go` - import path and example text
- [x] 2.3 Update `internal/cli/install.go` - descriptions and binary name in output
- [x] 2.4 Update `internal/cli/install_test.go` - expected binary path

## 3. Build and Installation
- [x] 3.1 Update `install.sh` - BINARY_NAME, LDFLAGS paths, build command, and success message
- [x] 3.2 Rename `.source-cli.yaml.example` to `.xdebug-cli.yaml.example`
- [x] 3.3 Update content inside the renamed config example file
- [x] 3.4 Update `.gitignore` - binary name and config file pattern

## 4. Documentation
- [x] 4.1 Update `README.md` - all references to source-cli
- [x] 4.2 Update `CLAUDE.md` - all references to source-cli

## 5. Verification
- [x] 5.1 Run `go mod tidy` to verify module resolution
- [x] 5.2 Run `go build -o xdebug-cli ./cmd/xdebug-cli` to verify build
- [x] 5.3 Run `go test ./...` to verify all tests pass
- [x] 5.4 Verify `xdebug-cli version` works correctly
