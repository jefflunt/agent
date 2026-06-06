package runner

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"sync"
)

// Runner dictates the contract for all backend driver runners.
type Runner interface {
	Run(ctx context.Context, model string, prompt string) (string, error)
}

// Registry or routing mechanism.
var (
	registry   = make(map[string]Runner)
	registryMu sync.RWMutex
)

// Register registers a Runner implementation.
func Register(name string, r Runner) {
	registryMu.Lock()
	defer registryMu.Unlock()
	registry[name] = r
}

// Get retrieves a Runner implementation by name.
func Get(name string) (Runner, bool) {
	registryMu.RLock()
	defer registryMu.RUnlock()
	r, ok := registry[name]
	return r, ok
}

// Command represents an executable command interface for easy mocking.
type Command interface {
	Start() error
	Wait() error
	StdoutPipe() (io.ReadCloser, error)
	StderrPipe() (io.ReadCloser, error)
}

// RealCommand wraps standard os/exec Cmd.
type RealCommand struct {
	cmd *exec.Cmd
}

// Start starts the command.
func (c *RealCommand) Start() error {
	return c.cmd.Start()
}

// Wait waits for the command to exit.
func (c *RealCommand) Wait() error {
	return c.cmd.Wait()
}

// StdoutPipe returns a pipe that will be connected to the command's standard output when the command starts.
func (c *RealCommand) StdoutPipe() (io.ReadCloser, error) {
	return c.cmd.StdoutPipe()
}

// StderrPipe returns a pipe that will be connected to the command's standard error when the command starts.
func (c *RealCommand) StderrPipe() (io.ReadCloser, error) {
	return c.cmd.StderrPipe()
}

// NewRealCommand is the default CommandFactory.
func NewRealCommand(ctx context.Context, name string, args ...string) Command {
	return &RealCommand{cmd: exec.CommandContext(ctx, name, args...)}
}

// OpencodeRunner implements the Runner interface for the opencode driver.
type OpencodeRunner struct {
	Executable     string
	CommandFactory func(ctx context.Context, name string, args ...string) Command
}

// OpencodeEvent represents a stream fragment from the opencode run CLI.
type OpencodeEvent struct {
	Type string `json:"type"`
	Part *struct {
		Text string `json:"text"`
	} `json:"part"`
}

// Run executes the under-the-hood opencode CLI subprocess and parses its output stream.
func (r *OpencodeRunner) Run(ctx context.Context, model string, prompt string) (string, error) {
	executable := r.Executable
	if executable == "" {
		executable = "opencode"
	}

	factory := r.CommandFactory
	if factory == nil {
		factory = NewRealCommand
	}

	args := []string{"run", prompt, "--model", model, "--format", "json"}
	cmd := factory(ctx, executable, args...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", fmt.Errorf("failed to get stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return "", fmt.Errorf("failed to get stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("failed to start subprocess %s: %w", executable, err)
	}

	// We must read standard error to avoid blocking and capture errors.
	var stderrBuf strings.Builder
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		_, _ = io.Copy(&stderrBuf, stderr)
	}()

	var aggregated strings.Builder
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		var ev OpencodeEvent
		if err := json.Unmarshal(line, &ev); err == nil {
			if ev.Type == "text" && ev.Part != nil {
				aggregated.WriteString(ev.Part.Text)
			}
		}
	}

	scanErr := scanner.Err()

	// Wait for stderr copying to complete.
	wg.Wait()

	waitErr := cmd.Wait()

	if waitErr != nil {
		stderrStr := strings.TrimSpace(stderrBuf.String())
		if stderrStr != "" {
			return "", fmt.Errorf("subprocess exited with error: %v, stderr: %s", waitErr, stderrStr)
		}
		return "", fmt.Errorf("subprocess exited with error: %v", waitErr)
	}

	if scanErr != nil {
		return "", fmt.Errorf("error scanning subprocess stdout: %w", scanErr)
	}

	return aggregated.String(), nil
}

// Placeholder is a temporary function to ensure the package is correctly compiled and imported.
func Placeholder() string {
	fmt.Println("pkg/runner placeholder called")
	return "runner"
}

func init() {
	Register("opencode", &OpencodeRunner{})
}
