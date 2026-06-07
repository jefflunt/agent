# Coding Conventions

This document outlines the standard coding style, structure, and idioms enforced in the Agent Router CLI.

## Standard Style & Go Idioms

### 1. Formatting
- All Go files must be formatted with `gofmt` (or `goimports`).
- Avoid unnecessary blank lines or overly long lines.

### 2. Error Handling
- Never ignore returned errors. Handle them explicitly or return them wrapped.
- Wrap low-level errors using `%w` to preserve context.
- Define explicit sentinel errors or custom error types for domain-level failure modes (e.g., `ErrEmptyTarget` or `ErrConfigMissing`).

### 3. Dependency Injection & Mockability
- Avoid global mutable states or direct calls to non-deterministic OS functions (like `os.ReadFile`, `os.LookupEnv`, `exec.CommandContext`) inside business packages.
- Instead, wrap OS or process operations behind clean, minimal interfaces (e.g., `OS` in `pkg/config/config.go` or `Command` in `pkg/runner/runner.go`).
- Accept interfaces or function-pointer options in struct initializers so unit tests can swap them with mock implementations.

---

## Code Structure Patterns

### Variable & Struct Naming
- Keep local variables short, precise, and idiomatic (e.g., `rnr` for Runner, `adp` for Adapter, `cfg` for Config, `err` for error).
- Prefer camelCase for local variables and PascalCase for exported identifiers.

### Package Responsibilities
- **`cmd/agent/`**: Execution environment setup, STDIN/STDOUT piping, flag evaluation. Zero business logic.
- **`pkg/config/`**: Locate, read, and unmarshal YAML configuration into standard configuration models.
- **`pkg/adapter/`**: Low-level parsing and validation of target spec formats.
- **`pkg/runner/`**: Runner definitions, registry lookup, and CLI subprocess orchestration.
