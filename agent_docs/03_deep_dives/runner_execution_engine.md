# Deep Dive: Runner Execution Engine

## Overview
The Runner Execution Engine provides a unified interface and registry for invoking, streaming, and sanitizing outputs from various external AI command-line drivers.

---

## Purpose
The primary role of the execution engine is to encapsulate the platform-specific complexities of spawning third-party AI CLI subprocesses. Each third-party CLI has unique inputs, outputs, progress reporting, ANSI escape sequences, API key expectations, and error conventions. The engine wraps all of these in a clean, consistent Go interface.

---

## Components

### 1. The Runner Interface (`pkg/runner/runner.go`)
Defines the execution contract:
```go
type Runner interface {
    Run(ctx context.Context, model string, prompt string) (string, error)
}
```

### 2. Thread-Safe Registry
Maintains a registry of available driver runners mapping the CLI name (e.g. `"copilot"`, `"opencode"`) to its corresponding `Runner` instance. Protected by a read-write mutex (`sync.RWMutex`).
- `Register(name string, r Runner)`
- `Get(name string) (Runner, bool)`

### 3. Mockable Command Factory (`Command`)
To support unit testing without spawning actual system binaries, commands are wrapped in a `Command` interface:
- `RealCommand`: Executes standard `exec.CommandContext`.
- `MockCommand`: Stubs execution flow, stdout/stderr channels, and exit codes.

---

## Individual Driver Implementations

### A. Opencode Driver (`OpencodeRunner`)
- **Invocation**: `opencode run <prompt> --model <model> --format json`
- **Streaming**: Captures standard output in real-time. Parses each newline-delimited stream event as an `OpencodeEvent` JSON struct:
  ```json
  {"type":"text","part":{"text":"fragment"}}
  ```
- **Aggregation**: Extracts text fragments and concatenates them to assemble the complete prompt answer.

### B. Copilot Driver (`CopilotRunner`)
- **Invocation**: `copilot -s -p "<prompt>" --excluded-tools=* --model <model>`
- **Output Sanitization**: The GitHub Copilot CLI outputs rich terminal layouts (carriage returns, spinners, ANSI color codes, and XML block wraps). The Copilot runner applies a five-step sanitization process:
  1. **Carriage Return Processing**: Processes `\r` occurrences to overwrite text in-memory exactly as a terminal emulator would.
  2. **ANSI Code Stripping**: Removes terminal color codes using regular expressions (`\x1b\[[0-9;?]*[a-zA-Z]`).
  3. **Control Character Filtering**: Drops all non-printable ASCII control characters.
  4. **Spinner & Progress Stripping**: Discards lines containing terminal spinners (`⠋`, `⠙`, etc.) or progress cues (like `thinking...` or `running tool...`).
  5. **Framing Wrapper Stripping**: Removes wrapper envelopes such as XML tags (`<response>`, `<result>`) or Markdown code blocks (```).

### C. Claude Driver (`ClaudeRunner`)
- **Invocation**: Native headless invocation with tool permissions sandboxed for safe, read-only headless prompts.

### D. Gemini Driver (`GeminiRunner`)
- **API Key Fallback**: If the environment variable `GEMINI_API_KEY` is missing, the Gemini runner attempts to read `GEMINI_KEY` from the system environment and injects it as `GEMINI_API_KEY` for the spawned process.

### E. Antigravity Driver (`AgyRunner`)
- **Invocation**: Integrates with the Google Antigravity CLI print-mode format.

---

## Subprocess Execution and Safety Flow

To avoid deadlock when executing subprocesses in Go:
1. Pipes are opened for both `StdoutPipe()` and `StderrPipe()`.
2. The command is started using `Start()`.
3. Standard error and standard output streams are read concurrently using Go-routines and standard buffers (or `bufio.Scanner`).
4. `cmd.Wait()` is invoked only after stream reading finishes to capture the subprocess exit code.
5. If the exit code is non-zero, the contents of standard error are compiled and returned inside a detailed error payload.
