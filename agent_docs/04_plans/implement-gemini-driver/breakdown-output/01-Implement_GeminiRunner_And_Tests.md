# Task 1: Implement `GeminiRunner` and Unit Tests in `pkg/runner`

## Objective
Implement the `GeminiRunner` struct conforming to the `Runner` interface, in `pkg/runner/runner.go` and register it via `init()`. Implement comprehensive unit tests in `pkg/runner/gemini_test.go`.

## Logic to Implement
1. Define the `GeminiRunner` struct in `pkg/runner/runner.go`:
   ```go
   type GeminiRunner struct {
       Executable     string
       CommandFactory func(ctx context.Context, name string, args ...string) Command
   }
   ```
2. Implement its `Run` method:
   - Accept `ctx context.Context`, `model string`, and `prompt string`.
   - Strip any provider prefix from the model string (e.g., `google/gemini-3.5-flash` becomes `gemini-3.5-flash`).
   - Run the subprocess with command: `gemini -p "<prompt>" --approval-mode=plan -m <model>`.
   - Asynchronously read from stdout and stderr to prevent buffer blockages.
   - Env Var Fallback: Since the `gemini` CLI requires `GEMINI_API_KEY`, if `GEMINI_API_KEY` is not set in the active environment, but `GEMINI_KEY` is set, we must inject `GEMINI_API_KEY` into the subprocess execution environment using `GEMINI_KEY`'s value.
   - Return stdout on success, and wrap any exit/wait errors with stderr details.
3. Register `"gemini"` to mapping registry in `init()`.

## Tests to Write
Create `pkg/runner/gemini_test.go` and add the following tests:
1. `TestGeminiRunner_Success`: Verifies subprocess executes with correct flags and returns expected clean text.
2. `TestGeminiRunner_StripsProviderPrefix`: Asserts that `google/gemini-3.5-flash` gets stripped down to `gemini-3.5-flash`.
3. `TestGeminiRunner_DefaultExecutable`: Asserts it defaults to `"gemini"`.
4. `TestGeminiRunner_SubprocessErrors`: Covers start and exit/wait failures.
