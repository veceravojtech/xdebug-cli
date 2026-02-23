---
project_name: 'xdebug-cli'
user_name: 'Vojta'
date: '2026-02-23'
sections_completed: ['technology_stack', 'language_rules', 'architecture_rules', 'testing_rules', 'code_quality', 'workflow_rules', 'critical_rules']
status: 'complete'
rule_count: 47
optimized_for_llm: true
---

# Project Context for AI Agents

_This file contains critical rules and patterns that AI agents must follow when implementing code in this project. Focus on unobvious details that agents might otherwise miss._

---

## Technology Stack & Versions

- **Go**: 1.25.4
- **CLI Framework**: github.com/spf13/cobra v1.10.1
- **Networking**: golang.org/x/net v0.47.0
- **Protocol**: DBGp (XML over TCP, Xdebug debug protocol)
- **IPC**: Unix domain sockets with JSON protocol
- **Build**: Shell script (install.sh) with ldflags version injection
- **Minimal dependency policy**: Only 2 direct dependencies. Do NOT add new dependencies without explicit approval.

## Critical Implementation Rules

### Go Language Rules

- **Error wrapping**: Always `fmt.Errorf("context: %w", err)`. Never return bare errors without context.
- **Early returns**: Return on error immediately. No nested else blocks after error checks.
- **Silent failures**: Only ignore errors (`_`) for non-critical cleanup operations (temp file removal, deadline clearing).
- **Constructors**: Always `New<Type>(dependencies)`. Never initialize structs directly from outside the package.
- **Receivers**: Single-letter abbreviation of type name: `s` for Server, `c` for Client, `v` for View, `e` for Executor, `d` for Daemon.
- **Imports**: Three groups (stdlib | internal | third-party), alphabetical within each. No dot imports, no aliases unless collision.
- **Mutexes**: `sync.RWMutex` for read-heavy state, `sync.Mutex` for write operations. Always per-struct, never package-level.
- **Type assertions**: Always check with `ok` pattern: `val, ok := result.(*Type)`. Return error if assertion fails.
- **Constants**: `PascalCase` for exported, `camelCase` for unexported. Group related constants with `const()` blocks.

### Architecture Rules

- **Layer boundaries**: CLI -> Daemon -> IPC -> DBGp -> View. Never skip layers or import across non-adjacent packages.
- **View adapter pattern**: View defines interfaces in `view/types.go`. DBGp implements them in `dbgp/view_adapters.go`. View NEVER imports dbgp directly.
- **Daemon model**: Two-process fork. Parent waits, child becomes daemon with TCP + Unix socket listeners. Always register in session registry.
- **File paths**: PID at `/tmp/xdebug-cli-daemon-{port}.pid`, status at `/tmp/xdebug-cli-daemon-{port}.status`, socket at `/tmp/xdebug-cli-session-{port}.sock`, registry at `~/.xdebug-cli/sessions.json`.
- **Command dispatch**: Single switch in `executor.go` with all aliases in same case. Every handler returns `ipc.CommandResult`. Never add command handling outside the executor.
- **IPC retry**: Exponential backoff `100ms * 2^attempt`. Use `ConnectWithRetry()`, never raw connect with sleep loops.
- **Graceful shutdown**: Use `context.WithCancel` and select on `ctx.Done()`. Never use `os.Exit()` in library code.
- **Global CLI args**: `CLIArgs cfg.CLIParameter` is the single source for flag values. Cobra flags bind to this struct.

### Testing Rules

- **Three test tiers**: `_test.go` (unit), `_integration_test.go` (integration), `_lifecycle_test.go` (lifecycle). Name files accordingly.
- **Same-package tests**: Tests live in the same package as source (white-box). Never use `_test` package suffix.
- **Table-driven tests**: Use `[]struct{ name string; ... }` with `t.Run(tt.name, ...)` for any test with 2+ cases.
- **No assertion libraries**: Use only `t.Errorf`, `t.Fatalf`, `t.Fatal`. No testify, no gomega. Match the minimal dependency policy.
- **Hand-written mocks**: Implement interfaces manually. No mock frameworks. Use factory functions like `createMock<Type>()`.
- **Test isolation**: `t.TempDir()` for filesystem, `defer os.Unsetenv()` for env vars, `defer os.Remove()` for temp files. Tests must not leak state.
- **Test naming**: `Test<Type>(t)`, `Test<Type>_<Method>(t)`, or `Test<Type>_<Scenario>(t)`. Integration tests: `Test<Type>Integration_<Scenario>`.

### Code Quality & Style Rules

- **Minimal main.go**: Entry point only wires root command and calls `Execute()`. No business logic in cmd/.
- **One concept per file**: Each file owns one primary type or responsibility. Split when a file grows beyond its concept.
- **File naming**: `snake_case.go`. Match filename to primary type/concept (e.g., `session.go` for `Session`).
- **No stuttering**: Package name is part of the identifier. Use `daemon.New()` or `daemon.Daemon`, not `daemon.NewDaemon` or `daemon.DaemonService`.
- **File structure order**: Type definition -> Constructor -> Exported methods -> Unexported helpers.
- **Comments**: Only for non-obvious behavior or protocol specifics. No redundant doc comments restating the function name.
- **No unused code**: No commented-out code blocks. No `_` prefixed unused variables kept "for later".
- **Version sync**: `internal/cfg/config.go` Version constant and `install.sh` VERSION must always match.

### Development Workflow Rules

- **Build command**: `go build -o xdebug-cli ./cmd/xdebug-cli`. Test with `go test ./...`.
- **Install**: `./install.sh` builds with ldflags and installs to `~/.local/bin/xdebug-cli`.
- **All code internal**: Everything lives under `internal/`. No `pkg/` directory. New packages go under `internal/`.
- **Commit style**: Imperative mood, concise. E.g., "Add breakpoint validation", "Fix EOF error diagnostics".
- **CLAUDE.md as source of truth**: Keep CLAUDE.md updated when adding commands, flags, or changing behavior. Follow the spec-to-documentation mapping table.
- **Spec-driven development**: Check `_bmad-output/planning-artifacts/` for specs before implementing features. Specs define the contract.

### Critical Don't-Miss Rules

- **No new dependencies**: The project has a minimal dependency policy (2 direct deps). Never add new Go modules without explicit approval.
- **No package-level state**: No `init()` functions, no package-level `var` for mutable state. All state lives in structs.
- **View-DBGp boundary**: View package must NEVER import `internal/dbgp`. Use the interface adapter pattern. This is the core decoupling mechanism.
- **No os.Exit() in libraries**: Only `main.go` may call `os.Exit()`. Library code returns errors up the call chain.
- **DBGp message framing**: Messages are `length\x00xmldata\x00`. Both null bytes are required. Respect `MaxMessageSize` (100MB).
- **Absolute breakpoint paths**: Xdebug requires absolute file paths for breakpoints. Validate and reject relative paths early.
- **Daemon auto-kill**: `daemon start` always kills existing daemon on the same port first. Never leave stale daemons.
- **PID validation**: Always check if PID is alive before trusting PID files. Stale PIDs in `/tmp/` are common after crashes.
- **Mutex discipline**: Session state accessed from both IPC handlers and TCP handlers concurrently. Always lock before read/write.
- **Exit codes matter**: 0 = success, 1 = error, 124 = breakpoint timeout. Scripts depend on these codes.

---

## Usage Guidelines

**For AI Agents:**

- Read this file before implementing any code
- Follow ALL rules exactly as documented
- When in doubt, prefer the more restrictive option
- Update this file if new patterns emerge

**For Humans:**

- Keep this file lean and focused on agent needs
- Update when technology stack changes
- Review quarterly for outdated rules
- Remove rules that become obvious over time

Last Updated: 2026-02-23
