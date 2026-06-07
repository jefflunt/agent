package runner

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"sync"
)

// Runner dictates the contract for all backend driver runners.
type Runner interface {
	Run(ctx context.Context, model string, prompt string, flags []string) (string, error)
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
func (r *OpencodeRunner) Run(ctx context.Context, model string, prompt string, flags []string) (string, error) {
	executable := r.Executable
	if executable == "" {
		executable = "opencode"
	}

	factory := r.CommandFactory
	if factory == nil {
		factory = NewRealCommand
	}

	var args []string
	if len(flags) > 0 {
		args = append([]string{"run", prompt, "--model", model}, flags...)
	} else {
		args = []string{"run", prompt, "--model", model, "--format", "json"}
	}
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

// CopilotRunner implements the Runner interface for the copilot driver.
type CopilotRunner struct {
	Executable     string
	CommandFactory func(ctx context.Context, name string, args ...string) Command
}

// Run executes the external copilot subprocess with the correct arguments,
// manages context-aware subprocess execution, and sanitizes the output stream.
func (r *CopilotRunner) Run(ctx context.Context, model string, prompt string, flags []string) (string, error) {
	executable := r.Executable
	if executable == "" {
		executable = "copilot"
	}

	factory := r.CommandFactory
	if factory == nil {
		factory = NewRealCommand
	}

	// Copilot expects only the bare model name. If a fully-qualified model
	// (e.g., "provider/model") is passed, we strip the provider prefix.
	if parts := strings.Split(model, "/"); len(parts) == 2 {
		model = parts[1]
	}

	var args []string
	if len(flags) > 0 {
		args = append([]string{"-p", prompt, "--model", model}, flags...)
	} else {
		// copilot -s -p "<prompt>" --excluded-tools=* --model <model>
		args = []string{"-s", "-p", prompt, "--excluded-tools=*", "--model", model}
	}
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

	var stdoutBuf strings.Builder
	wg.Add(1)
	go func() {
		defer wg.Done()
		_, _ = io.Copy(&stdoutBuf, stdout)
	}()

	wg.Wait()

	waitErr := cmd.Wait()

	if waitErr != nil {
		stderrStr := strings.TrimSpace(stderrBuf.String())
		if stderrStr != "" {
			return "", fmt.Errorf("subprocess exited with error: %v, stderr: %s", waitErr, stderrStr)
		}
		return "", fmt.Errorf("subprocess exited with error: %v", waitErr)
	}

	cleaned := sanitizeCopilotOutput(stdoutBuf.String())
	return cleaned, nil
}

func sanitizeCopilotOutput(raw string) string {
	// 1. Process carriage returns to handle lines overwritten in terminal.
	raw = processCarriageReturns(raw)

	// 2. Remove ANSI escape sequences
	ansiRegex := regexp.MustCompile(`\x1b\[[0-9;?]*[a-zA-Z]`)
	cleaned := ansiRegex.ReplaceAllString(raw, "")

	// 3. Filter out other unwanted control characters (ASCII < 32), keeping \n, \t, \r
	var sb strings.Builder
	for _, r := range cleaned {
		if r == '\n' || r == '\t' || r == '\r' {
			sb.WriteRune(r)
			continue
		}
		if r < 32 || (r >= 127 && r < 160) {
			continue // skip other control chars
		}
		sb.WriteRune(r)
	}
	cleaned = sb.String()

	// 4. Filter out standard progress indicators/markers and lines containing spinner characters
	lines := strings.Split(cleaned, "\n")
	var filteredLines []string

	spinnerChars := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	progressPrefixes := []string{
		"progress:",
		"thinking...",
		"analyzing...",
		"working...",
		"fetching...",
		"loading...",
		"running tool...",
	}

	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		lowerLine := strings.ToLower(trimmedLine)

		isProgress := false
		for _, prefix := range progressPrefixes {
			if strings.HasPrefix(lowerLine, prefix) {
				isProgress = true
				break
			}
		}
		if isProgress {
			continue
		}

		hasSpinner := false
		for _, char := range spinnerChars {
			if strings.Contains(trimmedLine, char) {
				hasSpinner = true
				break
			}
		}
		if hasSpinner {
			continue
		}

		if matched, _ := regexp.MatchString(`^(?i)\[\d+/\d+\]|progress:\s*\d+%|running\s+\w+\.\.\.`, trimmedLine); matched {
			continue
		}

		filteredLines = append(filteredLines, line)
	}

	cleaned = strings.Join(filteredLines, "\n")

	// 5. Handle framing wrappers.
	cleaned = strings.TrimSpace(cleaned)

	// A: Check XML-style wrapper e.g. <response>...</response>, <result>...</result>
	xmlWrappers := [][]string{
		{"<response>", "</response>"},
		{"<result>", "</result>"},
		{"<output>", "</output>"},
		{"<text>", "</text>"},
	}
	for _, w := range xmlWrappers {
		if strings.HasPrefix(strings.ToLower(cleaned), w[0]) && strings.HasSuffix(strings.ToLower(cleaned), w[1]) {
			cleaned = cleaned[len(w[0]) : len(cleaned)-len(w[1])]
			cleaned = strings.TrimSpace(cleaned)
			break
		}
	}

	// B: Check markdown code block wrapper (e.g., ```markdown ... ``` or ```json ... ``` or just ``` ... ```)
	if strings.HasPrefix(cleaned, "```") && strings.HasSuffix(cleaned, "```") {
		idx := strings.Index(cleaned, "\n")
		if idx != -1 {
			cleaned = cleaned[idx+1 : len(cleaned)-3]
			cleaned = strings.TrimSpace(cleaned)
		} else {
			cleaned = cleaned[3 : len(cleaned)-3]
			cleaned = strings.TrimSpace(cleaned)
		}
	}

	// C: Check "---" banner wrapper at the beginning and end
	if strings.HasPrefix(cleaned, "---") && strings.HasSuffix(cleaned, "---") {
		cleaned = cleaned[3 : len(cleaned)-3]
		cleaned = strings.TrimSpace(cleaned)
	}

	return cleaned
}

func processCarriageReturns(s string) string {
	var lines []string
	currentLine := []rune{}
	pos := 0

	for _, r := range s {
		if r == '\n' {
			lines = append(lines, string(currentLine))
			currentLine = []rune{}
			pos = 0
		} else if r == '\r' {
			pos = 0
		} else {
			if pos < len(currentLine) {
				currentLine[pos] = r
			} else {
				currentLine = append(currentLine, r)
			}
			pos++
		}
	}
	lines = append(lines, string(currentLine))
	return strings.Join(lines, "\n")
}

func init() {
	Register("opencode", &OpencodeRunner{})
	Register("copilot", &CopilotRunner{})
	Register("claude", &ClaudeRunner{})
	Register("gemini", &GeminiRunner{})
	Register("agy", &AgyRunner{})
}

// ClaudeRunner implements the Runner interface for the claude (Claude Code) driver.
type ClaudeRunner struct {
	Executable     string
	CommandFactory func(ctx context.Context, name string, args ...string) Command
}

// Run executes the under-the-hood claude CLI subprocess and captures its output.
func (r *ClaudeRunner) Run(ctx context.Context, model string, prompt string, flags []string) (string, error) {
	executable := r.Executable
	if executable == "" {
		executable = "claude"
	}

	factory := r.CommandFactory
	if factory == nil {
		factory = NewRealCommand
	}

	// Claude expects only the bare model name. If a fully-qualified model
	// (e.g., "provider/model") is passed, we strip the provider prefix.
	if parts := strings.Split(model, "/"); len(parts) == 2 {
		model = parts[1]
	}

	var args []string
	if len(flags) > 0 {
		args = append([]string{"-p", prompt, "--model", model}, flags...)
	} else {
		// claude -p "<prompt>" --tools="" --model <model>
		args = []string{"-p", prompt, "--tools=\"\"", "--model", model}
	}
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

	var stdoutBuf strings.Builder
	wg.Add(1)
	go func() {
		defer wg.Done()
		_, _ = io.Copy(&stdoutBuf, stdout)
	}()

	wg.Wait()

	waitErr := cmd.Wait()

	if waitErr != nil {
		stderrStr := strings.TrimSpace(stderrBuf.String())
		if stderrStr != "" {
			return "", fmt.Errorf("subprocess exited with error: %v, stderr: %s", waitErr, stderrStr)
		}
		return "", fmt.Errorf("subprocess exited with error: %v", waitErr)
	}

	return stdoutBuf.String(), nil
}

// GeminiRunner implements the Runner interface for the gemini driver.
type GeminiRunner struct {
	Executable     string
	CommandFactory func(ctx context.Context, name string, args ...string) Command
}

// Run executes the under-the-hood gemini CLI subprocess and captures its output.
func (r *GeminiRunner) Run(ctx context.Context, model string, prompt string, flags []string) (string, error) {
	executable := r.Executable
	if executable == "" {
		executable = "gemini"
	}

	factory := r.CommandFactory
	if factory == nil {
		factory = NewRealCommand
	}

	// Gemini expects only the bare model name. If a fully-qualified model
	// (e.g., "provider/model") is passed, we strip the provider prefix.
	if parts := strings.Split(model, "/"); len(parts) == 2 {
		model = parts[1]
	}

	// Propagate GEMINI_API_KEY if GEMINI_KEY is present and GEMINI_API_KEY is empty
	if os.Getenv("GEMINI_API_KEY") == "" {
		if val := os.Getenv("GEMINI_KEY"); val != "" {
			os.Setenv("GEMINI_API_KEY", val)
			defer os.Unsetenv("GEMINI_API_KEY")
		}
	}

	var args []string
	if len(flags) > 0 {
		args = append([]string{"-p", prompt, "-m", model}, flags...)
	} else {
		// gemini -p "<prompt>" --approval-mode=plan -m <model>
		args = []string{"-p", prompt, "--approval-mode=plan", "-m", model}
	}
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

	var stdoutBuf strings.Builder
	wg.Add(1)
	go func() {
		defer wg.Done()
		_, _ = io.Copy(&stdoutBuf, stdout)
	}()

	wg.Wait()

	waitErr := cmd.Wait()

	if waitErr != nil {
		stderrStr := strings.TrimSpace(stderrBuf.String())
		if stderrStr != "" {
			return "", fmt.Errorf("subprocess exited with error: %v, stderr: %s", waitErr, stderrStr)
		}
		return "", fmt.Errorf("subprocess exited with error: %v", waitErr)
	}

	return stdoutBuf.String(), nil
}

// AgyRunner implements the Runner interface for the agy (Google Antigravity CLI) driver.
type AgyRunner struct {
	Executable     string
	CommandFactory func(ctx context.Context, name string, args ...string) Command
}

// Run executes the under-the-hood agy CLI subprocess and captures its output.
func (r *AgyRunner) Run(ctx context.Context, model string, prompt string, flags []string) (string, error) {
	executable := r.Executable
	if executable == "" {
		executable = "agy"
	}

	factory := r.CommandFactory
	if factory == nil {
		factory = NewRealCommand
	}

	// agy --print "<prompt>" --model <model>
	args := []string{"--print", prompt, "--model", model}
	if len(flags) > 0 {
		args = append(args, flags...)
	}
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

	var stdoutBuf strings.Builder
	wg.Add(1)
	go func() {
		defer wg.Done()
		_, _ = io.Copy(&stdoutBuf, stdout)
	}()

	wg.Wait()

	waitErr := cmd.Wait()

	if waitErr != nil {
		stderrStr := strings.TrimSpace(stderrBuf.String())
		if stderrStr != "" {
			return "", fmt.Errorf("subprocess exited with error: %v, stderr: %s", waitErr, stderrStr)
		}
		return "", fmt.Errorf("subprocess exited with error: %v", waitErr)
	}

	return stdoutBuf.String(), nil
}
