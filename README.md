# Agent Router CLI (`agent`)

An AI-powered centralized routing facade and pipeline execution tool written in Go. The `agent` CLI acts as a unified gateway for multiple under-the-hood LLM CLI engines, standardizing input streams, output streams, configuration handling, and subprocess management.

With `agent`, you can map friendly user-defined adapters (such as `primary`, `fast`, or `test_claude`) to specific AI drivers and models, executing prompts seamlessly via UNIX pipes and redirects.

---

## **Features**

* 🚰 **Standardized I/O**: Reads prompts directly from `STDIN` and writes clean, sanitized responses to `STDOUT`.
* 🗺️ **Centralized Routing**: One simple configuration file (`~/.agent/config.yml`) routes requests to different underlying AI assistants.
* 📦 **Deep CLI Integrations**: Unified subprocess drivers for multiple popular LLM developer tools:
  * **GitHub Copilot CLI (`copilot`)**: Runs queries with automatic ANSI, progress-spinner, and wrapper-block stripping.
  * **Opencode (`opencode`)**: Runs JSON-streamed queries, aggregating tokens in real-time.
  * **Claude Code (`claude`)**: Direct native invocation with tool permissions sandboxed for safe, read-only headless prompts.
  * **Gemini CLI (`gemini`)**: Integrated headless plan-mode execution with automatic `GEMINI_API_KEY` credential fallback.
  * **Google Antigravity CLI (`agy`)**: Unified print-mode integration.

---

## **Installation**

To build and compile the `agent` binary, ensure you have Go (1.26+) installed, then execute:

```bash
# 1. Clone the repository and navigate to the directory
cd agent

# 2. Build the optimized agent CLI binary
go build -o agent cmd/agent/main.go

# 3. Move the binary into your active PATH for global access (optional)
sudo mv agent /usr/local/bin/
```

---

## **Configuration**

The `agent` CLI locates its configuration file by checking:
1. The **`AGENT_CONFIG_PATH`** environment variable (if set).
2. Falling back to the default file location at **`~/.agent/config.yml`**.

### **Mapping Format**
Adapters are defined inside the config using the following standard specification format:
```yaml
<adapter-name>: "<driver-cli>:<provider>/<model>"
```

### **Example Configuration (`~/.agent/config.yml`)**
Create or update your `~/.agent/config.yml` with your favorite drivers:

```yaml
adapters:
  # Using the Opencode Driver
  test_opencode: "opencode:google/gemini-3.5-flash"

  # Using the Copilot Driver
  test_copilot: "copilot:anthropic/claude-haiku-4.5"

  # Using the Claude Code Driver
  test_claude: "claude:anthropic/claude-sonnet-4-6"

  # Using the Gemini CLI Driver (Automatically injects GEMINI_API_KEY from GEMINI_KEY if missing)
  test_gemini: "gemini:google/gemini-3.5-flash"

  # Using the Google Antigravity CLI Driver
  test_antigravity: "agy:google/gemini-3.5-flash"
```

---

## **Usage**

### **1. Basic Prompts**
Pipe any prompt directly into `agent` along with your target adapter name:

```bash
echo "What is 2+2?" | agent test_copilot
# Output: 4
```

### **2. Verbose Mode**
Enable verbose logging on `STDERR` to inspect under-the-hood adapter resolutions, loaded paths, and timings using the `--verbose` or `-v` flags:

```bash
echo "Explain gravity in one sentence." | agent --verbose test_antigravity
```

**Stdout/Stderr Output:**
```text
verbose: successfully read prompt of length 32 bytes
verbose: resolved adapter "test_antigravity" to CLI "agy", Provider "google", Model "gemini-3.5-flash"
Gravity is the universal force of attraction acting between all matter.
```

---

## **Developer Quick Reference**

### **Project Architecture**
```text
├── cmd/
│   └── agent/
│       ├── main.go         # CLI Entrypoint & CLI execution context
│       └── main_test.go    # CLI Integration and flag tests
├── pkg/
│   ├── adapter/            # Splitter/Validator of "<cli>:<provider>/<model>" format
│   ├── config/             # YAML reader and AGENT_CONFIG_PATH parser
│   └── runner/             # Unified interface and Subprocess CLI Drivers:
│       ├── runner.go       # Core Registry & Command Factories
│       ├── copilot_test.go # Copilot sanitization & execution tests
│       ├── claude_test.go  # Claude unit tests
│       ├── gemini_test.go  # Gemini key fallback & unit tests
│       └── agy_test.go     # Antigravity unit tests
```

### **Running Tests**
Run the comprehensive test suite to verify code formatting, parsing, mock streaming, and driver routing:

```bash
go test -v ./...
```
