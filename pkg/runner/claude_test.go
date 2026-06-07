package runner

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"
)

func TestClaudeRunner_Success(t *testing.T) {
	stdoutData := "Hello from Claude!"
	var capturedCtx context.Context
	var capturedExecutable string
	var capturedArgs []string

	r := &ClaudeRunner{
		Executable: "custom-claude",
		CommandFactory: func(ctx context.Context, name string, args ...string) Command {
			capturedCtx = ctx
			capturedExecutable = name
			capturedArgs = args
			return &MockCommand{
				StdoutPipeFunc: func() (io.ReadCloser, error) {
					return io.NopCloser(strings.NewReader(stdoutData)), nil
				},
				StderrPipeFunc: func() (io.ReadCloser, error) {
					return io.NopCloser(strings.NewReader("")), nil
				},
			}
		},
	}

	ctx := context.Background()
	res, err := r.Run(ctx, "claude-sonnet-4-5", "test-prompt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res != "Hello from Claude!" {
		t.Errorf("expected response %q, got %q", "Hello from Claude!", res)
	}

	if capturedCtx != ctx {
		t.Error("context was not correctly passed to CommandFactory")
	}

	if capturedExecutable != "custom-claude" {
		t.Errorf("expected executable %q, got %q", "custom-claude", capturedExecutable)
	}

	expectedArgs := []string{"-p", "test-prompt", "--tools=\"\"", "--model", "claude-sonnet-4-5"}
	if len(capturedArgs) != len(expectedArgs) {
		t.Fatalf("expected %d args, got %d: %v", len(expectedArgs), len(capturedArgs), capturedArgs)
	}
	for i, arg := range capturedArgs {
		if arg != expectedArgs[i] {
			t.Errorf("expected arg at index %d to be %q, got %q", i, expectedArgs[i], arg)
		}
	}
}

func TestClaudeRunner_DefaultExecutable(t *testing.T) {
	var capturedExecutable string

	r := &ClaudeRunner{
		CommandFactory: func(ctx context.Context, name string, args ...string) Command {
			capturedExecutable = name
			return &MockCommand{}
		},
	}

	_, _ = r.Run(context.Background(), "model", "prompt")
	if capturedExecutable != "claude" {
		t.Errorf("expected executable to default to %q, got %q", "claude", capturedExecutable)
	}
}

func TestClaudeRunner_StripsProviderPrefix(t *testing.T) {
	var capturedArgs []string

	r := &ClaudeRunner{
		CommandFactory: func(ctx context.Context, name string, args ...string) Command {
			capturedArgs = args
			return &MockCommand{
				StdoutPipeFunc: func() (io.ReadCloser, error) {
					return io.NopCloser(strings.NewReader("response")), nil
				},
			}
		},
	}

	ctx := context.Background()
	_, err := r.Run(ctx, "anthropic/claude-sonnet-4-5", "prompt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedModelArg := "claude-sonnet-4-5"
	foundModel := false
	for i, arg := range capturedArgs {
		if arg == "--model" && i+1 < len(capturedArgs) {
			if capturedArgs[i+1] == expectedModelArg {
				foundModel = true
			}
		}
	}
	if !foundModel {
		t.Errorf("expected to find --model argument with value %q, but captured args were: %v", expectedModelArg, capturedArgs)
	}
}

func TestClaudeRunner_SubprocessExitErrorWithStderr(t *testing.T) {
	expectedWaitErr := errors.New("exit status 1")
	r := &ClaudeRunner{
		CommandFactory: func(ctx context.Context, name string, args ...string) Command {
			return &MockCommand{
				StderrPipeFunc: func() (io.ReadCloser, error) {
					return io.NopCloser(strings.NewReader("claude command error message\n")), nil
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
	if !strings.Contains(err.Error(), "claude command error message") {
		t.Errorf("expected stderr message in the error, but was: %v", err)
	}
}
