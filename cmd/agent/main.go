package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/jefflunt/agent/pkg/adapter"
	"github.com/jefflunt/agent/pkg/config"
	"github.com/jefflunt/agent/pkg/runner"
)

// CLI encapsulates the execution context of the agent CLI.
// This allows the entire execution logic to be fully tested by providing
// mocked or controlled streams, arguments, and configuration loaders.
type CLI struct {
	Stdin      io.Reader
	Stdout     io.Writer
	Stderr     io.Writer
	Args       []string
	LoadConfig func() (*config.Config, error)
}

// Run executes the CLI logic and returns an integer exit code (0 for success, non-zero for error).
func (c *CLI) Run() int {
	var verbose bool
	var positional []string

	// Skip the program name (Args[0]) if arguments are provided
	args := c.Args
	if len(args) > 0 {
		args = args[1:]
	}

	for _, arg := range args {
		if arg == "--verbose" || arg == "-v" {
			verbose = true
		} else if strings.HasPrefix(arg, "-") {
			fmt.Fprintf(c.Stderr, "Error: unknown flag %q\n", arg)
			c.printUsage()
			return 1
		} else {
			positional = append(positional, arg)
		}
	}

	if len(positional) == 0 {
		fmt.Fprintln(c.Stderr, "Error: adapter name is required")
		c.printUsage()
		return 1
	}
	if len(positional) > 1 {
		fmt.Fprintf(c.Stderr, "Error: too many arguments; expected exactly one adapter name, got %d\n", len(positional))
		c.printUsage()
		return 1
	}

	adapterName := positional[0]

	// Read from STDIN
	promptBytes, err := io.ReadAll(c.Stdin)
	if err != nil {
		fmt.Fprintf(c.Stderr, "Error: failed to read from STDIN: %v\n", err)
		return 1
	}

	prompt := strings.TrimSpace(string(promptBytes))
	if prompt == "" {
		fmt.Fprintln(c.Stderr, "Error: prompt cannot be empty")
		return 1
	}

	if verbose {
		fmt.Fprintf(c.Stderr, "verbose: successfully read prompt of length %d bytes\n", len(promptBytes))
	}

	// Load configuration
	loadConfig := c.LoadConfig
	if loadConfig == nil {
		loadConfig = config.Load
	}

	cfg, err := loadConfig()
	if err != nil {
		fmt.Fprintf(c.Stderr, "Error: failed to load configuration: %v\n", err)
		return 1
	}

	// Resolve the adapter target string
	spec, ok := cfg.Adapters[adapterName]
	if !ok {
		fmt.Fprintf(c.Stderr, "Error: adapter %q not found in configuration\n", adapterName)
		return 1
	}

	// Parse the adapter target string
	adp, err := adapter.Parse(spec)
	if err != nil {
		fmt.Fprintf(c.Stderr, "Error: failed to parse adapter specification for %q: %v\n", adapterName, err)
		return 1
	}

	if verbose {
		fmt.Fprintf(c.Stderr, "verbose: resolved adapter %q to CLI %q, Provider %q, Model %q\n",
			adapterName, adp.CLIName, adp.Provider, adp.Model)
	}

	// For now, since downstream runner implementation is in a placeholder stage, we call Placeholder() to satisfy imports.
	_ = runner.Placeholder()

	return 0
}

func (c *CLI) printUsage() {
	fmt.Fprintln(c.Stderr, "\nUsage: agent [options] <adapter-name>")
	fmt.Fprintln(c.Stderr, "Options:")
	fmt.Fprintln(c.Stderr, "  --verbose, -v   Enable verbose logging to STDERR")
}

func main() {
	cli := &CLI{
		Stdin:  os.Stdin,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
		Args:   os.Args,
	}
	os.Exit(cli.Run())
}
