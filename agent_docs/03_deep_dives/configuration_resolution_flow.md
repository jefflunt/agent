# Deep Dive: Configuration Resolution Flow

## Overview
This document walks through how the Agent Router CLI (`agent`) locates, reads, parses, and resolves its adapter configurations.

---

## Purpose
The configuration system translates friendly, user-defined adapter names (like `fast` or `primary`) to exact, low-level provider and model specifications. To support multiple developer machines and automated unit testing, this flow is dynamic, highly tolerant, and fully mockable.

---

## The Step-by-Step Resolution Flow

```
                      +-----------------------------+
                      |   1. Locate Config File     |
                      +-----------------------------+
                                     |
                Is AGENT_CONFIG_PATH env var set?
                /                           \
             [Yes]                         [No]
              /                               \
             v                                 v
    Use env var path                    Get UserHomeDir
                                               |
                                               v
                                   Append ".agent/config.yml"
                                     |
                                     v
                      +-----------------------------+
                      |     2. Read File Data       |
                      +-----------------------------+
                                     |
                             Does file exist?
                             /              \
                          [No]             [Yes]
                           /                  \
                          v                    v
              Return ErrConfigMissing     Read file bytes
                                               |
                                               v
                      +-----------------------------+
                      |      3. Parse YAML Data     |
                      +-----------------------------+
                                     |
                          Try to Unmarshal Nested
                        (under "adapters:" key)
                                     |
                            Successful parsing?
                            /               \
                         [No]              [Yes]
                          /                   \
                         v                     v
                 Try Flat Unmarshal       Return Config Model
                         |
                Successful parsing?
                /               \
             [No]              [Yes]
              /                   \
             v                     v
    Return parsing error     Return Config Model
```

### 1. Locating the Configuration File
The location is resolved inside `ResolvePath()` checking:
1. **`AGENT_CONFIG_PATH`**: If this environment variable is set and non-empty, the CLI uses its exact value.
2. **Default Fallback**: If missing, the engine resolves the current user's home directory using `UserHomeDir()` (e.g. `/Users/username` on macOS or `/home/username` on Linux) and appends `.agent/config.yml`.

### 2. Reading the File Data
- The file is read from the local disk using the `OS.ReadFile` interface.
- If the file is absent, a custom `ErrConfigMissing` error is returned. This struct provides a helpful prompt guiding the user on how to populate their configurations.

### 3. Parsing and Unmarshaling (Flexible Formats)
Our YAML parser in `ParseConfig()` is designed to be highly tolerant by supporting two distinct YAML layouts:

#### Layout A: Nested Map (Recommended)
This layout wraps adapter mappings under an explicit `adapters` root:
```yaml
adapters:
  test_copilot: "copilot:anthropic/claude-haiku-4.5"
  test_gemini: "gemini:google/gemini-3.5-flash"
```

#### Layout B: Flat Map
This layout defines the mappings directly at the document root level:
```yaml
test_copilot: "copilot:anthropic/claude-haiku-4.5"
test_gemini: "gemini:google/gemini-3.5-flash"
```

The parser first attempts to unmarshal the input bytes into the Nested structure. If no adapters are resolved, it automatically falls back to unmarshaling into a Flat map. If both unmarshal processes fail, a detailed malformed YAML error is thrown.
