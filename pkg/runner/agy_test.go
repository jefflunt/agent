package runner

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"
)

func TestAgyRunner_Success(t *testing.T) {
	stdoutData := "Hello from Agy!"
	var capturedCtx context.Context
	var capturedExecutable string
	var capturedArgs []string

	r := &AgyRunner{
		Executable: "custom-agy",
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
	res, err := r.Run(ctx, "google/gemini-3.5-flash", "test-prompt", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res != "Hello from Agy!" {
		t.Errorf("expected response %q, got %q", "Hello from Agy!", res)
	}

	if capturedCtx != ctx {
		t.Error("context was not correctly passed to CommandFactory")
	}

	if capturedExecutable != "custom-agy" {
		t.Errorf("expected executable %q, got %q", "custom-agy", capturedExecutable)
	}

	expectedArgs := []string{"--print", "test-prompt", "--model", "google/gemini-3.5-flash"}
	if len(capturedArgs) != len(expectedArgs) {
		t.Fatalf("expected %d args, got %d: %v", len(expectedArgs), len(capturedArgs), capturedArgs)
	}
	for i, arg := range capturedArgs {
		if arg != expectedArgs[i] {
			t.Errorf("expected arg at index %d to be %q, got %q", i, expectedArgs[i], arg)
		}
	}
}

func TestAgyRunner_DefaultExecutable(t *testing.T) {
	var capturedExecutable string

	r := &AgyRunner{
		CommandFactory: func(ctx context.Context, name string, args ...string) Command {
			capturedExecutable = name
			return &MockCommand{}
		},
	}

	_, _ = r.Run(context.Background(), "model", "prompt", nil)
	if capturedExecutable != "agy" {
		t.Errorf("expected executable to default to %q, got %q", "agy", capturedExecutable)
	}
}

func TestAgyRunner_SubprocessExitErrorWithStderr(t *testing.T) {
	expectedWaitErr := errors.New("exit status 1")
	r := &AgyRunner{
		CommandFactory: func(ctx context.Context, name string, args ...string) Command {
			return &MockCommand{
				StderrPipeFunc: func() (io.ReadCloser, error) {
					return io.NopCloser(strings.NewReader("agy command error message\n")), nil
				},
				WaitFunc: func() error {
					return expectedWaitErr
				},
			}
		},
	}

	_, err := r.Run(context.Background(), "model", "prompt", nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "subprocess exited with error") {
		t.Errorf("unexpected error message: %v", err)
	}
	if !strings.Contains(err.Error(), "agy command error message") {
		t.Errorf("expected stderr message in the error, but was: %v", err)
	}
}

func TestAgyRunner_CustomFlags(t *testing.T) {
	var capturedArgs []string
	r := &AgyRunner{
		CommandFactory: func(ctx context.Context, name string, args ...string) Command {
			capturedArgs = args
			return &MockCommand{
				StdoutPipeFunc: func() (io.ReadCloser, error) {
					return io.NopCloser(strings.NewReader("response")), nil
				},
			}
		},
	}

	customFlags := []string{"--some-flag", "--another-flag"}
	_, err := r.Run(context.Background(), "google/gemini-3.5-flash", "prompt", customFlags)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify that custom flags are appended.
	// Expected args: "--print", "prompt", "--model", "google/gemini-3.5-flash", "--some-flag", "--another-flag"
	expectedArgs := []string{"--print", "prompt", "--model", "google/gemini-3.5-flash", "--some-flag", "--another-flag"}
	if len(capturedArgs) != len(expectedArgs) {
		t.Fatalf("expected args to have length %d, got %v", len(expectedArgs), capturedArgs)
	}
	for i, arg := range capturedArgs {
		if arg != expectedArgs[i] {
			t.Errorf("expected arg at index %d to be %q, got %q", i, expectedArgs[i], arg)
		}
	}
}
