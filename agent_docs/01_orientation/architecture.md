# Architecture Overview

This document describes the high-level architecture of the Agent Router CLI (`agent`).

```
                +---------------------------------+
                |             STDIN               |
                +---------------------------------+
                                 |
                                 v
                +---------------------------------+
                |         cmd/agent/main.go       | <--- Reads Args & Flags
                +---------------------------------+
                    /                         \
                   /                           \ Loads config
   Parses target  /                             \
   specification /                               v
                v                       +------------------+
        +---------------+               |    pkg/config    | (AGENT_CONFIG_PATH or
        |  pkg/adapter  |               +------------------+  ~/.agent/config.yml)
        +---------------+
                \                               /
                 \                             /
                  v                           v
                +---------------------------------+
                |        Runner Resolution        | (Match adp.CLIName)
                +---------------------------------+
                                 |
                                 v
                        +------------------+
                        |    pkg/runner    | (Registry of CLI Subprocess Drivers)
                        +------------------+
                        /   |    |    |    \
                       /    |    |    |     \
                      v     v    v    v      v
                     agy claude copilot gemini opencode
                                 |
                                 v
                +---------------------------------+
                |             STDOUT              | (Sanitized/formatted reply)
                +---------------------------------+
```

## System Components

### 1. CLI Entrypoint (`cmd/agent/main.go`)
The CLI acts as the coordinator. It:
- Reads the input prompt from `STDIN`.
- Resolves configuration via `pkg/config`.
- Selects the correct adapter based on the positional CLI argument.
- Delegates target parsing to `pkg/adapter`.
- Resolves the matching subprocess runner in `pkg/runner`.
- Runs the driver, streams stdout, and outputs the final result to `STDOUT`.

### 2. Configuration Loader (`pkg/config/config.go`)
- Finds and parses the YAML configuration file.
- Checks `AGENT_CONFIG_PATH` first, falling back to `~/.agent/config.yml`.
- Supports flat YAML mappings and nested mappings (under the `adapters` key).
- Abstracted with an `OS` interface for seamless file system and env variable unit testing.

### 3. Adapter Spec Parser (`pkg/adapter/adapter.go`)
- Parses target spec strings in the format `cliName:provider/model` (e.g. `copilot:anthropic/claude-haiku-4.5`).
- Validates delimiters and raises detailed errors for empty parts or missing components.

### 4. Runner Subprocess Engine (`pkg/runner/`)
- Dictates a unified `Runner` interface: `Run(ctx context.Context, model string, prompt string) (string, error)`.
- Implements individual subprocess runners that invoke underlying CLI tools:
  - **`agy`**: Direct execution of Google Antigravity CLI.
  - **`claude`**: Safe, sandboxed invocation of Claude Code CLI.
  - **`copilot`**: Handles complex terminal ANSI cleanup, carriage return scrubbing, progress indicators/spinners stripping, and XML wrapper removals.
  - **`gemini`**: Integrated headless plan execution with API key environment fallbacks.
  - **`opencode`**: Parses real-time streamed JSON lines, aggregating text tokens.
- Employs a `Command` factory wrapper (`RealCommand` vs mocked test commands) for reliable unit testing of CLI invocation mechanics.
