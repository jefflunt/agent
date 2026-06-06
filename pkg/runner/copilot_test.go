package runner

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"
)

func TestCopilotRunner_Success(t *testing.T) {
	stdoutData := "<response>\n```markdown\n---Hello from Copilot!---\n```\n</response>"
	var capturedCtx context.Context
	var capturedExecutable string
	var capturedArgs []string

	r := &CopilotRunner{
		Executable: "custom-copilot",
		CommandFactory: func(ctx context.Context, name string, args ...string) Command {
			capturedCtx = ctx
			capturedExecutable = name
			capturedArgs = args
			return &MockCommand{
				StdoutPipeFunc: func() (io.ReadCloser, error) {
					return io.NopCloser(strings.NewReader(stdoutData)), nil
				},
				StderrPipeFunc: func() (io.ReadCloser, error) {
					return io.NopCloser(strings.NewReader("some debug stderr data")), nil
				},
			}
		},
	}

	ctx := context.Background()
	res, err := r.Run(ctx, "copilot-gpt-4", "test-prompt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res != "Hello from Copilot!" {
		t.Errorf("expected sanitized response %q, got %q", "Hello from Copilot!", res)
	}

	if capturedCtx != ctx {
		t.Error("context was not correctly passed to CommandFactory")
	}

	if capturedExecutable != "custom-copilot" {
		t.Errorf("expected executable %q, got %q", "custom-copilot", capturedExecutable)
	}

	expectedArgs := []string{"-s", "-p", "test-prompt", "--excluded-tools=*", "--model", "copilot-gpt-4"}
	if len(capturedArgs) != len(expectedArgs) {
		t.Fatalf("expected %d args, got %d: %v", len(expectedArgs), len(capturedArgs), capturedArgs)
	}
	for i, arg := range capturedArgs {
		if arg != expectedArgs[i] {
			t.Errorf("expected arg at index %d to be %q, got %q", i, expectedArgs[i], arg)
		}
	}
}

func TestCopilotRunner_DefaultExecutable(t *testing.T) {
	var capturedExecutable string

	r := &CopilotRunner{
		CommandFactory: func(ctx context.Context, name string, args ...string) Command {
			capturedExecutable = name
			return &MockCommand{}
		},
	}

	_, _ = r.Run(context.Background(), "model", "prompt")
	if capturedExecutable != "copilot" {
		t.Errorf("expected executable to default to %q, got %q", "copilot", capturedExecutable)
	}
}

func TestCopilotRunner_StdoutPipeError(t *testing.T) {
	expectedErr := errors.New("stdout pipe error")
	r := &CopilotRunner{
		CommandFactory: func(ctx context.Context, name string, args ...string) Command {
			return &MockCommand{
				StdoutPipeFunc: func() (io.ReadCloser, error) {
					return nil, expectedErr
				},
			}
		},
	}

	_, err := r.Run(context.Background(), "model", "prompt")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "failed to get stdout pipe") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestCopilotRunner_StderrPipeError(t *testing.T) {
	expectedErr := errors.New("stderr pipe error")
	r := &CopilotRunner{
		CommandFactory: func(ctx context.Context, name string, args ...string) Command {
			return &MockCommand{
				StderrPipeFunc: func() (io.ReadCloser, error) {
					return nil, expectedErr
				},
			}
		},
	}

	_, err := r.Run(context.Background(), "model", "prompt")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "failed to get stderr pipe") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestCopilotRunner_StartError(t *testing.T) {
	expectedErr := errors.New("start failed")
	r := &CopilotRunner{
		CommandFactory: func(ctx context.Context, name string, args ...string) Command {
			return &MockCommand{
				StartFunc: func() error {
					return expectedErr
				},
			}
		},
	}

	_, err := r.Run(context.Background(), "model", "prompt")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "failed to start subprocess") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestCopilotRunner_SubprocessExitErrorWithStderr(t *testing.T) {
	expectedWaitErr := errors.New("exit status 1")
	r := &CopilotRunner{
		CommandFactory: func(ctx context.Context, name string, args ...string) Command {
			return &MockCommand{
				StderrPipeFunc: func() (io.ReadCloser, error) {
					return io.NopCloser(strings.NewReader("copilot binary failed to authenticate\n")), nil
				},
				WaitFunc: func() error {
					return expectedWaitErr
				},
			}
		},
	}

	_, err := r.Run(context.Background(), "model", "prompt")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "subprocess exited with error") {
		t.Errorf("unexpected error message: %v", err)
	}
	if !strings.Contains(err.Error(), "copilot binary failed to authenticate") {
		t.Errorf("expected stderr message in the error, but was: %v", err)
	}
}

func TestCopilotRunner_SubprocessExitErrorEmptyStderr(t *testing.T) {
	expectedWaitErr := errors.New("exit status 1")
	r := &CopilotRunner{
		CommandFactory: func(ctx context.Context, name string, args ...string) Command {
			return &MockCommand{
				WaitFunc: func() error {
					return expectedWaitErr
				},
			}
		},
	}

	_, err := r.Run(context.Background(), "model", "prompt")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "subprocess exited with error") {
		t.Errorf("unexpected error message: %v", err)
	}
	if strings.Contains(err.Error(), "stderr:") {
		t.Errorf("did not expect stderr part in error since stderr was empty, but got: %v", err)
	}
}

func TestSanitizeCopilotOutput(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple text",
			input:    "Hello, World!",
			expected: "Hello, World!",
		},
		{
			name:     "ansi escape sequences removal",
			input:    "\x1b[31mRed\x1b[0m and \x1b[1mbold\x1b[0m text.",
			expected: "Red and bold text.",
		},
		{
			name:     "other control characters removal",
			input:    "Control\x01\x02chars\x1fshould\x7fbe\u009fremoved.",
			expected: "Controlcharsshouldberemoved.",
		},
		{
			name:     "carriage return line overrides",
			input:    "Line 1\nHello\rWorld\rNew Text",
			expected: "Line 1\nNew Text",
		},
		{
			name:     "carriage return line overrides with longer first word",
			input:    "Line 1\nLongestWord\rShort",
			expected: "Line 1\nShortstWord",
		},
		{
			name: "filtering progress markers and spinner lines",
			input: `⠋ working hard...
progress: 15%
thinking...
analyzing...
working...
fetching...
loading...
running tool...
[1/15] step 1
progress: 99%
running some_tool...
actual response text is here
⠙ spinner line to ignore`,
			expected: "actual response text is here",
		},
		{
			name:     "XML response wrapper",
			input:    "<response>\nResponse content\n</response>",
			expected: "Response content",
		},
		{
			name:     "XML result wrapper",
			input:    "<result>Result content</result>",
			expected: "Result content",
		},
		{
			name:     "XML output wrapper",
			input:    "<output>Output content</output>",
			expected: "Output content",
		},
		{
			name:     "XML text wrapper",
			input:    "<text>Text content</text>",
			expected: "Text content",
		},
		{
			name:     "XML wrapper case insensitivity",
			input:    "<RESPONSE>\nCase insensitive response\n</response>",
			expected: "Case insensitive response",
		},
		{
			name:     "markdown code block multi line",
			input:    "```markdown\n# This is markdown\nMore content here\n```",
			expected: "# This is markdown\nMore content here",
		},
		{
			name:     "markdown code block json",
			input:    "```json\n{\"foo\": \"bar\"}\n```",
			expected: "{\"foo\": \"bar\"}",
		},
		{
			name:     "markdown code block single line",
			input:    "```simple text```",
			expected: "simple text",
		},
		{
			name:     "dashed banner wrapper",
			input:    "---And some text---",
			expected: "And some text",
		},
		{
			name: "all wrappers nested together",
			input: "<response>\n" + "```" + "markdown\n---Nested cleanup success---\n" + "```" + "\n</response>",
			expected: "Nested cleanup success",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sanitizeCopilotOutput(tt.input)
			if got != tt.expected {
				t.Errorf("sanitizeCopilotOutput() = %q; expected %q", got, tt.expected)
			}
		})
	}
}
