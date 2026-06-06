package runner

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"
)

func TestPlaceholder(t *testing.T) {
	got := Placeholder()
	want := "runner"
	if got != want {
		t.Errorf("Placeholder() = %q; want %q", got, want)
	}
}

// MockCommand implements the Command interface for testing.
type MockCommand struct {
	StartFunc      func() error
	WaitFunc       func() error
	StdoutPipeFunc func() (io.ReadCloser, error)
	StderrPipeFunc func() (io.ReadCloser, error)
}

func (m *MockCommand) Start() error {
	if m.StartFunc != nil {
		return m.StartFunc()
	}
	return nil
}

func (m *MockCommand) Wait() error {
	if m.WaitFunc != nil {
		return m.WaitFunc()
	}
	return nil
}

func (m *MockCommand) StdoutPipe() (io.ReadCloser, error) {
	if m.StdoutPipeFunc != nil {
		return m.StdoutPipeFunc()
	}
	return io.NopCloser(strings.NewReader("")), nil
}

func (m *MockCommand) StderrPipe() (io.ReadCloser, error) {
	if m.StderrPipeFunc != nil {
		return m.StderrPipeFunc()
	}
	return io.NopCloser(strings.NewReader("")), nil
}

func TestRegistry(t *testing.T) {
	mockName := "test-custom-runner"
	mockR := &OpencodeRunner{}
	Register(mockName, mockR)

	r, found := Get(mockName)
	if !found {
		t.Errorf("expected to find runner %s, but got false", mockName)
	}
	if r != mockR {
		t.Errorf("expected retrieved runner to be %v, got %v", mockR, r)
	}

	_, found = Get("nonexistent-runner-name")
	if found {
		t.Error("expected not to find nonexistent-runner-name, but got true")
	}
}

func TestNewRealCommand(t *testing.T) {
	ctx := context.Background()
	cmd := NewRealCommand(ctx, "echo", "hello", "world")
	realCmd, ok := cmd.(*RealCommand)
	if !ok {
		t.Fatalf("expected NewRealCommand to return *RealCommand, got %T", cmd)
	}
	if realCmd.cmd == nil {
		t.Fatal("expected inner cmd to be initialized")
	}
	if len(realCmd.cmd.Args) != 3 || realCmd.cmd.Args[0] != "echo" || realCmd.cmd.Args[1] != "hello" || realCmd.cmd.Args[2] != "world" {
		t.Errorf("unexpected cmd.Args: %v", realCmd.cmd.Args)
	}
}

func TestOpencodeRunner_Success(t *testing.T) {
	stdoutData := `{"type":"text","part":{"text":"Hello "}}
{"type":"text","part":{"text":"World!"}}
`
	var capturedCtx context.Context
	var capturedExecutable string
	var capturedArgs []string

	r := &OpencodeRunner{
		Executable: "custom-bin",
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
	res, err := r.Run(ctx, "my-model", "my-prompt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res != "Hello World!" {
		t.Errorf("expected response to be %q, got %q", "Hello World!", res)
	}

	if capturedCtx != ctx {
		t.Error("context was not correctly passed to CommandFactory")
	}

	if capturedExecutable != "custom-bin" {
		t.Errorf("expected executable %q, got %q", "custom-bin", capturedExecutable)
	}

	expectedArgs := []string{"run", "my-prompt", "--model", "my-model", "--format", "json"}
	if len(capturedArgs) != len(expectedArgs) {
		t.Fatalf("expected %d args, got %d: %v", len(expectedArgs), len(capturedArgs), capturedArgs)
	}
	for i, arg := range capturedArgs {
		if arg != expectedArgs[i] {
			t.Errorf("expected arg at index %d to be %q, got %q", i, expectedArgs[i], arg)
		}
	}
}

func TestOpencodeRunner_DefaultExecutable(t *testing.T) {
	var capturedExecutable string

	r := &OpencodeRunner{
		CommandFactory: func(ctx context.Context, name string, args ...string) Command {
			capturedExecutable = name
			return &MockCommand{}
		},
	}

	_, _ = r.Run(context.Background(), "model", "prompt")
	if capturedExecutable != "opencode" {
		t.Errorf("expected executable to default to %q, got %q", "opencode", capturedExecutable)
	}
}

func TestOpencodeRunner_StdoutPipeError(t *testing.T) {
	expectedErr := errors.New("stdout pipe error")
	r := &OpencodeRunner{
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

func TestOpencodeRunner_StderrPipeError(t *testing.T) {
	expectedErr := errors.New("stderr pipe error")
	r := &OpencodeRunner{
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

func TestOpencodeRunner_StartError(t *testing.T) {
	expectedErr := errors.New("start failed")
	r := &OpencodeRunner{
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

func TestOpencodeRunner_SubprocessExitErrorWithStderr(t *testing.T) {
	expectedWaitErr := errors.New("exit status 1")
	r := &OpencodeRunner{
		CommandFactory: func(ctx context.Context, name string, args ...string) Command {
			return &MockCommand{
				StderrPipeFunc: func() (io.ReadCloser, error) {
					return io.NopCloser(strings.NewReader("some detailed stderr message\n")), nil
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
	if !strings.Contains(err.Error(), "some detailed stderr message") {
		t.Errorf("expected stderr message in the error, but was: %v", err)
	}
}

func TestOpencodeRunner_SubprocessExitErrorEmptyStderr(t *testing.T) {
	expectedWaitErr := errors.New("exit status 1")
	r := &OpencodeRunner{
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

type errorReader struct {
	err error
}

func (er errorReader) Read(p []byte) (n int, err error) {
	return 0, er.err
}

func TestOpencodeRunner_ScannerError(t *testing.T) {
	expectedScanErr := errors.New("scanner failure")
	r := &OpencodeRunner{
		CommandFactory: func(ctx context.Context, name string, args ...string) Command {
			return &MockCommand{
				StdoutPipeFunc: func() (io.ReadCloser, error) {
					return io.NopCloser(errorReader{err: expectedScanErr}), nil
				},
			}
		},
	}

	_, err := r.Run(context.Background(), "model", "prompt")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "error scanning subprocess stdout") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestOpencodeRunner_InvalidAndMixedJSON(t *testing.T) {
	stdoutData := `{"type":"text","part":{"text":"Hello "}}
invalid-json-line-to-be-ignored
{"type":"other-type","part":{"text":"Ignored type"}}
{"type":"text","part":null}
{"type":"text","part":{"text":"World!"}}
`
	r := &OpencodeRunner{
		CommandFactory: func(ctx context.Context, name string, args ...string) Command {
			return &MockCommand{
				StdoutPipeFunc: func() (io.ReadCloser, error) {
					return io.NopCloser(strings.NewReader(stdoutData)), nil
				},
			}
		},
	}

	res, err := r.Run(context.Background(), "model", "prompt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res != "Hello World!" {
		t.Errorf("expected aggregated response to ignore invalid/mismatched JSON lines and return %q, got %q", "Hello World!", res)
	}
}
