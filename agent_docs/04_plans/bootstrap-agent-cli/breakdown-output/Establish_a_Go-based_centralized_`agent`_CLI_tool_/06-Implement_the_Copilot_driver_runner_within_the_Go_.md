# Implement the Copilot driver runner within the Go agent CLI tool to execute the `copilot` subprocess using the correct arguments, manage context-aware subprocess execution, and sanitize the output stream to return clean text responses.

This task focuses on implementing the `CopilotRunner` struct and its methods in the `pkg/runner` package. The runner will implement the common `Runner` interface: `Run(ctx context.Context, model string, prompt string) (string, error)`. It is responsible for orchestrating the execution of the external `copilot` binary.

The implementation must invoke the `copilot` command with the following options: `copilot -s -p "<prompt>" --excluded-tools=* --model <model>`. It must properly handle subprocess state, process contexts, timeout cancellations, and exit codes. Standard error output from the subprocess should be handled gracefully or routed to stderr under a verbose flag.

In addition, the runner must parse and sanitize the raw subprocess output. Since external CLI processes may emit control characters, progress markers, or framing wrappers, a dedicated stream parser or cleanup routine is required to isolate and return the clean response text back to the main router.
