# Task 1: Implement `ClaudeRunner` and Unit Tests in `pkg/runner`

## Objective
Implement the `ClaudeRunner` struct, conforming to the `Runner` interface, in `pkg/runner/runner.go` and register it via `init()`. Implement comprehensive unit tests in `pkg/runner/claude_test.go`.

## Logic to Implement
1. Define the `ClaudeRunner` struct in `pkg/runner/runner.go`:
   ```go
   type ClaudeRunner struct {
       Executable     string
       CommandFactory func(ctx context.Context, name string, args ...string) Command
   }
   ```
2. Implement its `Run` method:
   - Accept `ctx context.Context`, `model string`, and `prompt string`.
   - Strip any provider prefix from the model string (e.g., `anthropic/claude-sonnet-4-5` becomes `claude-sonnet-4-5`).
   - Run the subprocess with command: `claude -p "<prompt>" --tools="" --model <model>`.
   - Asynchronously read from stdout and stderr to prevent buffer blockages.
   - Return stdout on success, and wrap any exit/wait errors with stderr details.
3. Register `"claude"` to mapping registry in `init()`.

## Tests to Write
Create `pkg/runner/claude_test.go` and add the following tests:
1. `TestClaudeRunner_Success`: Verifies subprocess executes with correct flags and returns expected clean text.
2. `TestClaudeRunner_StripsProviderPrefix`: Asserts that `anthropic/claude-sonnet-4-5` gets stripped down to `claude-sonnet-4-5`.
3. `TestClaudeRunner_DefaultExecutable`: Asserts it defaults to `"claude"`.
4. `TestClaudeRunner_SubprocessErrors`: Covers start and exit/wait failures.
