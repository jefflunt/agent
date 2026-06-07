# Domain Concepts

This document outlines the core concepts and terminologies used across the `agent` codebase.

## Core Concepts

### 1. Adapter Name
A friendly, user-defined string alias (e.g., `primary`, `fast`, `test_claude`, `experimental`) specified in the YAML configuration. Instead of invoking low-level commands with lengthy flags, users pipe prompts directly to these high-level adapter names:
```bash
echo "Hello" | agent fast
```

### 2. Specification String (Spec)
The format specifying the target CLI driver, provider, and model. It follows the pattern:
```text
<cliName>:<provider>/<model>
```
*Example*: `copilot:anthropic/claude-haiku-4.5` maps:
- `cliName` = `copilot`
- `provider` = `anthropic`
- `model` = `claude-haiku-4.5`

### 3. Driver (CLI Driver)
The under-the-hood CLI executable tool that interfaces with the AI providers. Supported drivers are:
- `opencode`
- `copilot` (GitHub Copilot CLI)
- `claude` (Claude Code)
- `gemini` (Gemini CLI)
- `agy` (Google Antigravity CLI)

### 4. Runner
The Go implementation responsible for orchestrating the execution of a specific Driver. Each Runner wraps the logic of spawning a subprocess, supplying parameters (such as `--model` and the query prompt), capturing `stdout`/`stderr`, and applying tailored output sanitization or stream unmarshaling.

### 5. Prompt
The text prompt received on `STDIN` representing the instruction or task. The CLI strips trailing spaces and empty newlines before handing the prompt over to the Runner.

### 6. Subprocess Command Factory
An abstraction (`Command` and `RealCommand`) that wraps the Go standard `os/exec` command creation. This allows testing of CLI drivers without running real binary processes, verifying instead the exact arguments generated and standard streams.
