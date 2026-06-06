# Implement a Go-based CLI entrypoint in main.go that parses a positional argument (the target adapter name) and options (such as --verbose), reads the user prompt from STDIN, and prints appropriate errors to STDERR, serving as the input handler for the agent router.

This task involves implementing the main Go entrypoint (main.go) for the centralized 'agent' CLI utility. The primary goal is to parse positional command-line arguments to extract the target adapter profile name, handle option flags such as '--verbose' (which controls whether internal driver timing or logs are written to STDERR), and consume piped stream data from STDIN as the prompt buffer.

The input handler must validate that a valid adapter name is provided as a positional argument. If the adapter name is missing, it should write an informative usage error to STDERR and exit with a non-zero code. Similarly, it must read the full prompt from STDIN; if STDIN is empty or cannot be read, it must handle the condition cleanly.

Once the input is successfully captured, the entrypoint will serve as the gateway to downstream components (config parsing, driver routing, and runner execution). During this implementation stage, the core focus is ensuring robust argument parsing and STDIN buffer aggregation inside main.go.
