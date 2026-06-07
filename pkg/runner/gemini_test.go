package runner

import (
	"context"
	"errors"
	"io"
	"os"
	"strings"
	"testing"
)

func TestGeminiRunner_Success(t *testing.T) {
	stdoutData := "Hello from Gemini!"
	var capturedCtx context.Context
	var capturedExecutable string
	var capturedArgs []string

	r := &GeminiRunner{
		Executable: "custom-gemini",
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
	res, err := r.Run(ctx, "gemini-3.5-flash", "test-prompt", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res != "Hello from Gemini!" {
		t.Errorf("expected response %q, got %q", "Hello from Gemini!", res)
	}

	if capturedCtx != ctx {
		t.Error("context was not correctly passed to CommandFactory")
	}

	if capturedExecutable != "custom-gemini" {
		t.Errorf("expected executable %q, got %q", "custom-gemini", capturedExecutable)
	}

	expectedArgs := []string{"-p", "test-prompt", "--approval-mode=plan", "-m", "gemini-3.5-flash"}
	if len(capturedArgs) != len(expectedArgs) {
		t.Fatalf("expected %d args, got %d: %v", len(expectedArgs), len(capturedArgs), capturedArgs)
	}
	for i, arg := range capturedArgs {
		if arg != expectedArgs[i] {
			t.Errorf("expected arg at index %d to be %q, got %q", i, expectedArgs[i], arg)
		}
	}
}

func TestGeminiRunner_DefaultExecutable(t *testing.T) {
	var capturedExecutable string

	r := &GeminiRunner{
		CommandFactory: func(ctx context.Context, name string, args ...string) Command {
			capturedExecutable = name
			return &MockCommand{}
		},
	}

	_, _ = r.Run(context.Background(), "model", "prompt", nil)
	if capturedExecutable != "gemini" {
		t.Errorf("expected executable to default to %q, got %q", "gemini", capturedExecutable)
	}
}

func TestGeminiRunner_StripsProviderPrefix(t *testing.T) {
	var capturedArgs []string

	r := &GeminiRunner{
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
	_, err := r.Run(ctx, "google/gemini-3.5-flash", "prompt", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedModelArg := "gemini-3.5-flash"
	foundModel := false
	for i, arg := range capturedArgs {
		if arg == "-m" && i+1 < len(capturedArgs) {
			if capturedArgs[i+1] == expectedModelArg {
				foundModel = true
			}
		}
	}
	if !foundModel {
		t.Errorf("expected to find -m argument with value %q, but captured args were: %v", expectedModelArg, capturedArgs)
	}
}

func TestGeminiRunner_EnvVarFallback(t *testing.T) {
	// Preserve existing environment variables
	origAPIKey, apiSet := os.LookupEnv("GEMINI_API_KEY")
	origGeminiKey, keySet := os.LookupEnv("GEMINI_KEY")

	// Ensure both are cleared initially for testing
	os.Unsetenv("GEMINI_API_KEY")
	os.Setenv("GEMINI_KEY", "fallback-key-value-123")

	defer func() {
		// Restore original state
		if apiSet {
			os.Setenv("GEMINI_API_KEY", origAPIKey)
		} else {
			os.Unsetenv("GEMINI_API_KEY")
		}
		if keySet {
			os.Setenv("GEMINI_KEY", origGeminiKey)
		} else {
			os.Unsetenv("GEMINI_KEY")
		}
	}()

	var capturedEnvVal string
	r := &GeminiRunner{
		CommandFactory: func(ctx context.Context, name string, args ...string) Command {
			capturedEnvVal = os.Getenv("GEMINI_API_KEY")
			return &MockCommand{
				StdoutPipeFunc: func() (io.ReadCloser, error) {
					return io.NopCloser(strings.NewReader("response")), nil
				},
			}
		},
	}

	_, err := r.Run(context.Background(), "model", "prompt", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedEnvVal != "fallback-key-value-123" {
		t.Errorf("expected GEMINI_API_KEY environment variable to be propagated as %q, but got %q", "fallback-key-value-123", capturedEnvVal)
	}
}

func TestGeminiRunner_SubprocessExitErrorWithStderr(t *testing.T) {
	expectedWaitErr := errors.New("exit status 1")
	r := &GeminiRunner{
		CommandFactory: func(ctx context.Context, name string, args ...string) Command {
			return &MockCommand{
				StderrPipeFunc: func() (io.ReadCloser, error) {
					return io.NopCloser(strings.NewReader("gemini command error message\n")), nil
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
	if !strings.Contains(err.Error(), "gemini command error message") {
		t.Errorf("expected stderr message in the error, but was: %v", err)
	}
}

func TestGeminiRunner_CustomFlags(t *testing.T) {
	var capturedArgs []string
	r := &GeminiRunner{
		CommandFactory: func(ctx context.Context, name string, args ...string) Command {
			capturedArgs = args
			return &MockCommand{
				StdoutPipeFunc: func() (io.ReadCloser, error) {
					return io.NopCloser(strings.NewReader("response")), nil
				},
			}
		},
	}

	customFlags := []string{"--approval-mode=manual", "--some-flag"}
	_, err := r.Run(context.Background(), "gemini-3.5-flash", "prompt", customFlags)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify that behavioral flags ("--approval-mode=plan") are bypassed, but structural flags are kept.
	// Expected args: "-p", "prompt", "-m", "gemini-3.5-flash", "--approval-mode=manual", "--some-flag"
	expectedArgs := []string{"-p", "prompt", "-m", "gemini-3.5-flash", "--approval-mode=manual", "--some-flag"}
	if len(capturedArgs) != len(expectedArgs) {
		t.Fatalf("expected args to have length %d, got %v", len(expectedArgs), capturedArgs)
	}
	for i, arg := range capturedArgs {
		if arg != expectedArgs[i] {
			t.Errorf("expected arg at index %d to be %q, got %q", i, expectedArgs[i], arg)
		}
	}
}
