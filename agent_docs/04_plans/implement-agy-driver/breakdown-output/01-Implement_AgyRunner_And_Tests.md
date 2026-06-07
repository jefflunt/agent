# Task 1: Implement `AgyRunner` and Unit Tests in `pkg/runner`

## Objective
Implement the `AgyRunner` struct conforming to the `Runner` interface, in `pkg/runner/runner.go` and register it via `init()`. Implement comprehensive unit tests in `pkg/runner/agy_test.go`.

## Logic to Implement
1. Define the `AgyRunner` struct in `pkg/runner/runner.go`:
   ```go
   type AgyRunner struct {
       Executable     string
       CommandFactory func(ctx context.Context, name string, args ...string) Command
   }
   ```
2. Implement its `Run` method:
   - Accept `ctx context.Context`, `model string`, and `prompt string`.
   - Run the subprocess with command: `agy --print "<prompt>" --model <model>`.
   - Asynchronously read from stdout and stderr to prevent buffer blockages.
   - Return stdout on success, and wrap any exit/wait errors with stderr details.
3. Register `"agy"` to mapping registry in `init()`.

## Tests to Write
Create `pkg/runner/agy_test.go` and add the following tests:
1. `TestAgyRunner_Success`: Verifies subprocess executes with correct flags and returns expected clean text.
2. `TestAgyRunner_DefaultExecutable`: Asserts it defaults to `"agy"`.
3. `TestAgyRunner_SubprocessErrors`: Covers start and exit/wait failures.
