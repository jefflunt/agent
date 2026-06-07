# Onboarding & Setup

This document helps developers set up their local environment to run, test, and contribute to the Agent Router CLI.

## Prerequisites

To build and compile the project, ensure you have:
- **Go**: Version 1.26 or newer.
- **Git**: Installed for repository and dependency management.

---

## Getting Started

### 1. Build the Binary
Clone or navigate to the project directory and build the binary:
```bash
go build -o agent cmd/agent/main.go
```

### 2. Run the Test Suite
Ensure the codebase is working correctly by executing all Go tests:
```bash
go test -v ./...
```

---

## Configuration Setup

Before executing the CLI, you must set up an adapter configuration file.

### Default Path
The CLI checks for configuration at:
1. `AGENT_CONFIG_PATH` environment variable if specified.
2. `~/.agent/config.yml` (default path).

### 1. Create a Configuration Directory and File
Create the default directory and write a basic configuration:
```bash
mkdir -p ~/.agent
cat <<EOF > ~/.agent/config.yml
adapters:
  test_copilot: "copilot:anthropic/claude-haiku-4.5"
  test_gemini: "gemini:google/gemini-3.5-flash"
EOF
```

### 2. Define Environment Variables (If Required by Drivers)
Some drivers might require specific API keys or CLI installations:
- **Gemini**: Needs `GEMINI_API_KEY` (or `GEMINI_KEY` which is automatically backed up/injected by the runner).
- **Copilot**: Requires GitHub Copilot CLI authenticated session (`gh copilot`).
- **Claude**: Requires `claude` (Claude Code) installed.
- **Opencode**: Requires `opencode` installed.

---

## Verifying Local Execution

Verify that your build works correctly using a pipe input:

```bash
echo "Hello, world!" | ./agent test_copilot
```

### Using Verbose Mode
Use `--verbose` or `-v` flags to print adapter mapping logic, exact parsed models, and CLI runner routing to `STDERR`:

```bash
echo "Hello, world!" | ./agent -v test_copilot
```
