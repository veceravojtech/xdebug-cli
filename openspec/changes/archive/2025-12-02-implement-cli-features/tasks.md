## 1. Project Scaffold

- [x] 1.1 Initialize Go module (`go mod init`)
- [x] 1.2 Create `cmd/source-cli/main.go` entry point
- [x] 1.3 Create `internal/cli/root.go` with root command setup
- [x] 1.4 Add Cobra and Viper dependencies
- [x] 1.5 Verify project builds with `go build ./...`

## 2. Preview Command

- [x] 2.1 Create `internal/cli/preview.go` with preview subcommand
- [x] 2.2 Implement animated progress indicator (spinner/loading animation)
- [x] 2.3 Add duration parsing (e.g., `10s`, `5m`)
- [x] 2.4 Create `internal/cli/preview_test.go` with unit tests
- [x] 2.5 Verify `source-cli preview source 10s` works correctly

## 3. Install Command

- [x] 3.1 Create `internal/cli/install.go` with install subcommand
- [x] 3.2 Implement build with timestamp embedding via ldflags
- [x] 3.3 Implement copy to `~/.local/bin/` with directory creation
- [x] 3.4 Create `internal/cli/install_test.go` with unit tests
- [x] 3.5 Verify `source-cli install` installs binary correctly

## 4. Distribution

- [x] 4.1 Create `install.sh` script for easy installation
- [x] 4.2 Create `.source-cli.yaml.example` configuration template
- [x] 4.3 Update `.gitignore` for Go artifacts and config files

## 5. Documentation

- [x] 5.1 Update CLAUDE.md with final command documentation
- [x] 5.2 Create README.md with usage instructions
