# Refactor the `Runner` interface in `pkg/runner/runner.go` to accept custom CLI flags (`flags []string`). Update all five implementations (`OpencodeRunner`, `CopilotRunner`, `ClaudeRunner`, `GeminiRunner`, and `AgyRunner`) to implement this updated signature and conditionally bypass their hardcoded default behavioral flags (e.g., `--format json` for Opencode, `-s` and `--excluded-tools=*` for Copilot, `--tools=""` for Claude, and `--approval-mode=plan` for Gemini) when custom flags are provided, while ensuring structural parameters are preserved.

The objective of this task is to update the signature of the `Runner` interface and its five concrete implementations in `pkg/runner/runner.go` to accept an additional argument: `flags []string`. When custom flags are provided, each runner must bypass its default optional behavioral flags while preserving structural arguments.

Following the Driver Integration Matrix:
- **`opencode`**: Keeps `run <prompt> --model <model>`. Bypasses `--format json` if custom flags are provided.
- **`copilot`**: Keeps `-p <prompt> --model <model>`. Bypasses `-s` and `--excluded-tools=*` if custom flags are provided.
- **`claude`**: Keeps `-p <prompt> --model <model>`. Bypasses `--tools=""` if custom flags are provided.
- **`gemini`**: Keeps `-p <prompt> -m <model>`. Bypasses `--approval-mode=plan` if custom flags are provided.
- **`agy`**: Keeps `--print <prompt> --model <model>`. Custom flags are always appended as there are no default behavioral flags to bypass.

All implementations must construct the updated subprocess command arguments accordingly.
