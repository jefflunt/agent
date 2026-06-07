# Testing Patterns

This document describes the testing architecture and guidelines for writing and executing tests in the Agent Router CLI.

## Core Testing Philosophy

- **Zero External Dependencies**: We rely solely on the Go standard library `testing` package. Do not introduce heavy assertion libraries.
- **Hermetic Testing**: All tests must run locally, in parallel, and without internet access or depending on physical configuration files on the host machine.
- **High Coverage of Failure Paths**: Validate error handling, invalid inputs, and corrupt outputs or stream chunk boundaries.

---

## Testing Techniques & Mocking Standards

### 1. Mocking OS and File System (`pkg/config`)
We use interface-based mocks for file reading and environment variable lookups. 

**Standard**:
Define a mock structure that mirrors the package's `OS` interface, allowing functional overrides of its behaviors:

```go
type MockOS struct {
	ReadFileFunc    func(name string) ([]byte, error)
	UserHomeDirFunc func() (string, error)
	LookupEnvFunc   func(key string) (string, bool)
}

func (m MockOS) ReadFile(name string) ([]byte, error) {
	if m.ReadFileFunc != nil {
		return m.ReadFileFunc(name)
	}
	return nil, os.ErrNotExist
}
```

### 2. Mocking Subprocesses and CLI Streams (`pkg/runner`)
Subprocess drivers invoke command line tools. Spawning real commands under unit tests is fragile. Instead, we use a `Command` factory interface that returns a mock-capable abstraction.

**Standard**:
Test runners using `MockCommand` to stub execution flow, stdout, stderr streams, and return codes:

```go
type MockCommand struct {
	StartFunc      func() error
	WaitFunc       func() error
	StdoutPipeFunc func() (io.ReadCloser, error)
	StderrPipeFunc func() (io.ReadCloser, error)
}
```

Inside your unit test, assign a custom `CommandFactory` that verifies that the arguments passed to the CLI driver matches your expectation, and return your stubbed `MockCommand` containing the mock output streams.

### 3. Testing CLI Main entrypoint (`cmd/agent`)
To fully test flag evaluation, exit codes, and output streaming without using real command-line contexts, our entrypoint is encapsulated in a testable `CLI` struct:

```go
type CLI struct {
	Stdin      io.Reader
	Stdout     io.Writer
	Stderr     io.Writer
	Args       []string
	LoadConfig func() (*config.Config, error)
}
```

In `cmd/agent/main_test.go`, instantiate the `CLI` struct with mocked STDIN bytes, custom CLI args, and a mock config loader, then evaluate the integer exit code returned by `CLI.Run()`.

---

## Executing Tests

To run the comprehensive test suite with verbose reporting:
```bash
go test -v ./...
```
