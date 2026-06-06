package main

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/jefflunt/agent/pkg/config"
	"github.com/jefflunt/agent/pkg/runner"
)

type mockRunner struct{}

func (m *mockRunner) Run(ctx context.Context, model string, prompt string) (string, error) {
	return "mock response", nil
}

func init() {
	runner.Register("copilot", &mockRunner{})
}

// errorReader is an io.Reader that always returns an error.
type errorReader struct{}

func (errorReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("read error")
}

func TestCLI_Success(t *testing.T) {
	stdin := strings.NewReader("user prompt here")
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	cli := &CLI{
		Stdin:  stdin,
		Stdout: stdout,
		Stderr: stderr,
		Args:   []string{"agent", "test-adapter"},
		LoadConfig: func() (*config.Config, error) {
			return &config.Config{
				Adapters: map[string]string{
					"test-adapter": "copilot:openai/gpt-4o",
				},
			}, nil
		},
	}

	exitCode := cli.Run()
	if exitCode != 0 {
		t.Errorf("expected exit code 0, got %d. Stderr: %q", exitCode, stderr.String())
	}
}

func TestCLI_Success_Verbose(t *testing.T) {
	stdin := strings.NewReader("  user prompt with spaces   ")
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	cli := &CLI{
		Stdin:  stdin,
		Stdout: stdout,
		Stderr: stderr,
		Args:   []string{"agent", "--verbose", "test-adapter"},
		LoadConfig: func() (*config.Config, error) {
			return &config.Config{
				Adapters: map[string]string{
					"test-adapter": "copilot:openai/gpt-4o",
				},
			}, nil
		},
	}

	exitCode := cli.Run()
	if exitCode != 0 {
		t.Errorf("expected exit code 0, got %d. Stderr: %q", exitCode, stderr.String())
	}

	stderrStr := stderr.String()
	if !strings.Contains(stderrStr, "verbose: successfully read prompt") {
		t.Errorf("expected verbose log in stderr, got: %q", stderrStr)
	}
	if !strings.Contains(stderrStr, "verbose: resolved adapter") {
		t.Errorf("expected verbose resolved log in stderr, got: %q", stderrStr)
	}
}

func TestCLI_Success_ShortVerbose(t *testing.T) {
	stdin := strings.NewReader("prompt")
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	cli := &CLI{
		Stdin:  stdin,
		Stdout: stdout,
		Stderr: stderr,
		Args:   []string{"agent", "test-adapter", "-v"},
		LoadConfig: func() (*config.Config, error) {
			return &config.Config{
				Adapters: map[string]string{
					"test-adapter": "copilot:openai/gpt-4o",
				},
			}, nil
		},
	}

	exitCode := cli.Run()
	if exitCode != 0 {
		t.Errorf("expected exit code 0, got %d. Stderr: %q", exitCode, stderr.String())
	}

	stderrStr := stderr.String()
	if !strings.Contains(stderrStr, "verbose:") {
		t.Errorf("expected verbose log in stderr, got: %q", stderrStr)
	}
}

func TestCLI_MissingAdapter(t *testing.T) {
	stdin := strings.NewReader("prompt")
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	cli := &CLI{
		Stdin:  stdin,
		Stdout: stdout,
		Stderr: stderr,
		Args:   []string{"agent"},
		LoadConfig: func() (*config.Config, error) {
			return &config.Config{}, nil
		},
	}

	exitCode := cli.Run()
	if exitCode != 1 {
		t.Errorf("expected exit code 1, got %d", exitCode)
	}

	stderrStr := stderr.String()
	if !strings.Contains(stderrStr, "Error: adapter name is required") {
		t.Errorf("expected required adapter error, got: %q", stderrStr)
	}
	if !strings.Contains(stderrStr, "Usage:") {
		t.Errorf("expected usage instructions, got: %q", stderrStr)
	}
}

func TestCLI_TooManyArguments(t *testing.T) {
	stdin := strings.NewReader("prompt")
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	cli := &CLI{
		Stdin:  stdin,
		Stdout: stdout,
		Stderr: stderr,
		Args:   []string{"agent", "adapter1", "adapter2"},
		LoadConfig: func() (*config.Config, error) {
			return &config.Config{}, nil
		},
	}

	exitCode := cli.Run()
	if exitCode != 1 {
		t.Errorf("expected exit code 1, got %d", exitCode)
	}

	stderrStr := stderr.String()
	if !strings.Contains(stderrStr, "Error: too many arguments") {
		t.Errorf("expected too many arguments error, got: %q", stderrStr)
	}
}

func TestCLI_UnknownFlag(t *testing.T) {
	stdin := strings.NewReader("prompt")
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	cli := &CLI{
		Stdin:  stdin,
		Stdout: stdout,
		Stderr: stderr,
		Args:   []string{"agent", "--unknown", "test-adapter"},
		LoadConfig: func() (*config.Config, error) {
			return &config.Config{}, nil
		},
	}

	exitCode := cli.Run()
	if exitCode != 1 {
		t.Errorf("expected exit code 1, got %d", exitCode)
	}

	stderrStr := stderr.String()
	if !strings.Contains(stderrStr, "Error: unknown flag") {
		t.Errorf("expected unknown flag error, got: %q", stderrStr)
	}
}

func TestCLI_EmptyStdinPrompt(t *testing.T) {
	stdin := strings.NewReader("   \n   \t")
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	cli := &CLI{
		Stdin:  stdin,
		Stdout: stdout,
		Stderr: stderr,
		Args:   []string{"agent", "test-adapter"},
		LoadConfig: func() (*config.Config, error) {
			return &config.Config{}, nil
		},
	}

	exitCode := cli.Run()
	if exitCode != 1 {
		t.Errorf("expected exit code 1, got %d", exitCode)
	}

	stderrStr := stderr.String()
	if !strings.Contains(stderrStr, "Error: prompt cannot be empty") {
		t.Errorf("expected prompt empty error, got: %q", stderrStr)
	}
}

func TestCLI_StdinReadError(t *testing.T) {
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	cli := &CLI{
		Stdin:  errorReader{},
		Stdout: stdout,
		Stderr: stderr,
		Args:   []string{"agent", "test-adapter"},
		LoadConfig: func() (*config.Config, error) {
			return &config.Config{}, nil
		},
	}

	exitCode := cli.Run()
	if exitCode != 1 {
		t.Errorf("expected exit code 1, got %d", exitCode)
	}

	stderrStr := stderr.String()
	if !strings.Contains(stderrStr, "Error: failed to read from STDIN") {
		t.Errorf("expected failed read from STDIN error, got: %q", stderrStr)
	}
}

func TestCLI_LoadConfigError(t *testing.T) {
	stdin := strings.NewReader("prompt")
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	cli := &CLI{
		Stdin:  stdin,
		Stdout: stdout,
		Stderr: stderr,
		Args:   []string{"agent", "test-adapter"},
		LoadConfig: func() (*config.Config, error) {
			return nil, errors.New("custom config error")
		},
	}

	exitCode := cli.Run()
	if exitCode != 1 {
		t.Errorf("expected exit code 1, got %d", exitCode)
	}

	stderrStr := stderr.String()
	if !strings.Contains(stderrStr, "Error: failed to load configuration: custom config error") {
		t.Errorf("expected failed configuration load error, got: %q", stderrStr)
	}
}

func TestCLI_AdapterNotFound(t *testing.T) {
	stdin := strings.NewReader("prompt")
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	cli := &CLI{
		Stdin:  stdin,
		Stdout: stdout,
		Stderr: stderr,
		Args:   []string{"agent", "nonexistent-adapter"},
		LoadConfig: func() (*config.Config, error) {
			return &config.Config{
				Adapters: map[string]string{
					"test-adapter": "copilot:openai/gpt-4o",
				},
			}, nil
		},
	}

	exitCode := cli.Run()
	if exitCode != 1 {
		t.Errorf("expected exit code 1, got %d", exitCode)
	}

	stderrStr := stderr.String()
	if !strings.Contains(stderrStr, `Error: adapter "nonexistent-adapter" not found in configuration`) {
		t.Errorf("expected missing adapter error, got: %q", stderrStr)
	}
}

func TestCLI_AdapterParseError(t *testing.T) {
	stdin := strings.NewReader("prompt")
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	cli := &CLI{
		Stdin:  stdin,
		Stdout: stdout,
		Stderr: stderr,
		Args:   []string{"agent", "bad-adapter"},
		LoadConfig: func() (*config.Config, error) {
			return &config.Config{
				Adapters: map[string]string{
					"bad-adapter": "invalid_format",
				},
			}, nil
		},
	}

	exitCode := cli.Run()
	if exitCode != 1 {
		t.Errorf("expected exit code 1, got %d", exitCode)
	}

	stderrStr := stderr.String()
	if !strings.Contains(stderrStr, "Error: failed to parse adapter specification for") {
		t.Errorf("expected adapter parse error, got: %q", stderrStr)
	}
}
