package config

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type MockOS struct {
	ReadFileFunc    func(name string) ([]byte, error)
	UserHomeDirFunc func() (string, error)
	LookupEnvFunc   func(key string) (string, bool)
}

func (m MockOS) ReadFile(name string) ([]byte, error) {
	if m.ReadFileFunc != nil {
		return m.ReadFileFunc(name)
	}
	return nil, os.ErrNotExist
}

func (m MockOS) UserHomeDir() (string, error) {
	if m.UserHomeDirFunc != nil {
		return m.UserHomeDirFunc()
	}
	return "", errors.New("not implemented")
}

func (m MockOS) LookupEnv(key string) (string, bool) {
	if m.LookupEnvFunc != nil {
		return m.LookupEnvFunc(key)
	}
	return "", false
}

func TestPlaceholder(t *testing.T) {
	got := Placeholder()
	want := "config"
	if got != want {
		t.Errorf("Placeholder() = %q; want %q", got, want)
	}
}

func TestResolvePath_EnvVarSet(t *testing.T) {
	mock := MockOS{
		LookupEnvFunc: func(key string) (string, bool) {
			if key == "AGENT_CONFIG_PATH" {
				return "/custom/path/config.yml", true
			}
			return "", false
		},
	}

	got, err := ResolvePath(mock)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "/custom/path/config.yml"
	if got != want {
		t.Errorf("ResolvePath() = %q; want %q", got, want)
	}
}

func TestResolvePath_DefaultHome(t *testing.T) {
	mock := MockOS{
		LookupEnvFunc: func(key string) (string, bool) {
			return "", false
		},
		UserHomeDirFunc: func() (string, error) {
			return "/home/user", nil
		},
	}

	got, err := ResolvePath(mock)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := filepath.Join("/home/user", ".agent", "config.yml")
	if got != want {
		t.Errorf("ResolvePath() = %q; want %q", got, want)
	}
}

func TestResolvePath_HomeError(t *testing.T) {
	mock := MockOS{
		LookupEnvFunc: func(key string) (string, bool) {
			return "", false
		},
		UserHomeDirFunc: func() (string, error) {
			return "", errors.New("home directory missing")
		},
	}

	_, err := ResolvePath(mock)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "failed to resolve user home directory") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestLoadWithOS_SuccessFlat(t *testing.T) {
	yamlData := `
primary: "adapter-spec-1"
fast: "adapter-spec-2"
`
	mock := MockOS{
		LookupEnvFunc: func(key string) (string, bool) {
			if key == "AGENT_CONFIG_PATH" {
				return "/home/user/.agent/config.yml", true
			}
			return "", false
		},
		ReadFileFunc: func(name string) ([]byte, error) {
			if name == "/home/user/.agent/config.yml" {
				return []byte(yamlData), nil
			}
			return nil, os.ErrNotExist
		},
	}

	cfg, err := LoadWithOS(mock)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg == nil {
		t.Fatal("expected non-nil config")
	}

	if len(cfg.Adapters) != 2 {
		t.Errorf("expected 2 adapters, got %d", len(cfg.Adapters))
	}
	if cfg.Adapters["primary"] != "adapter-spec-1" {
		t.Errorf("expected adapters[primary] to be %q, got %q", "adapter-spec-1", cfg.Adapters["primary"])
	}
	if cfg.Adapters["fast"] != "adapter-spec-2" {
		t.Errorf("expected adapters[fast] to be %q, got %q", "adapter-spec-2", cfg.Adapters["fast"])
	}
}

func TestLoadWithOS_SuccessNested(t *testing.T) {
	yamlData := `
adapters:
  primary: "adapter-spec-1"
  claude: "adapter-spec-3"
`
	mock := MockOS{
		LookupEnvFunc: func(key string) (string, bool) {
			if key == "AGENT_CONFIG_PATH" {
				return "/home/user/.agent/config.yml", true
			}
			return "", false
		},
		ReadFileFunc: func(name string) ([]byte, error) {
			if name == "/home/user/.agent/config.yml" {
				return []byte(yamlData), nil
			}
			return nil, os.ErrNotExist
		},
	}

	cfg, err := LoadWithOS(mock)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg == nil {
		t.Fatal("expected non-nil config")
	}

	if len(cfg.Adapters) != 2 {
		t.Errorf("expected 2 adapters, got %d", len(cfg.Adapters))
	}
	if cfg.Adapters["primary"] != "adapter-spec-1" {
		t.Errorf("expected adapters[primary] to be %q, got %q", "adapter-spec-1", cfg.Adapters["primary"])
	}
	if cfg.Adapters["claude"] != "adapter-spec-3" {
		t.Errorf("expected adapters[claude] to be %q, got %q", "adapter-spec-3", cfg.Adapters["claude"])
	}
}

func TestLoadWithOS_ConfigMissing(t *testing.T) {
	mock := MockOS{
		LookupEnvFunc: func(key string) (string, bool) {
			if key == "AGENT_CONFIG_PATH" {
				return "/home/user/.agent/config.yml", true
			}
			return "", false
		},
		ReadFileFunc: func(name string) ([]byte, error) {
			return nil, os.ErrNotExist
		},
	}

	_, err := LoadWithOS(mock)
	if err == nil {
		t.Fatal("expected missing config error, got nil")
	}

	var missingErr *ErrConfigMissing
	if !errors.As(err, &missingErr) {
		t.Fatalf("expected ErrConfigMissing error, got %T: %v", err, err)
	}

	if missingErr.Path != "/home/user/.agent/config.yml" {
		t.Errorf("expected missing path %q, got %q", "/home/user/.agent/config.yml", missingErr.Path)
	}

	expectedMessage := `configuration file not found at "/home/user/.agent/config.yml"`
	if !strings.Contains(missingErr.Error(), expectedMessage) {
		t.Errorf("unexpected error message: %v", missingErr.Error())
	}
}

func TestLoadWithOS_ConfigMissingFsErr(t *testing.T) {
	mock := MockOS{
		LookupEnvFunc: func(key string) (string, bool) {
			if key == "AGENT_CONFIG_PATH" {
				return "/home/user/.agent/config.yml", true
			}
			return "", false
		},
		ReadFileFunc: func(name string) ([]byte, error) {
			return nil, fs.ErrNotExist
		},
	}

	_, err := LoadWithOS(mock)
	if err == nil {
		t.Fatal("expected missing config error, got nil")
	}

	var missingErr *ErrConfigMissing
	if !errors.As(err, &missingErr) {
		t.Fatalf("expected ErrConfigMissing error, got %T: %v", err, err)
	}
}

func TestLoadWithOS_ReadError(t *testing.T) {
	mock := MockOS{
		LookupEnvFunc: func(key string) (string, bool) {
			if key == "AGENT_CONFIG_PATH" {
				return "/home/user/.agent/config.yml", true
			}
			return "", false
		},
		ReadFileFunc: func(name string) ([]byte, error) {
			return nil, errors.New("permission denied")
		},
	}

	_, err := LoadWithOS(mock)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !strings.Contains(err.Error(), "failed to read configuration file") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestLoadWithOS_MalformedYAML(t *testing.T) {
	yamlData := `
adapters:
  [malformed: yaml
`
	mock := MockOS{
		LookupEnvFunc: func(key string) (string, bool) {
			if key == "AGENT_CONFIG_PATH" {
				return "/home/user/.agent/config.yml", true
			}
			return "", false
		},
		ReadFileFunc: func(name string) ([]byte, error) {
			return []byte(yamlData), nil
		},
	}

	_, err := LoadWithOS(mock)
	if err == nil {
		t.Fatal("expected YAML parsing error, got nil")
	}

	if !strings.Contains(err.Error(), "malformed YAML") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestLoadWithOS_EmptyYAML(t *testing.T) {
	yamlData := `
# Just comments
`
	mock := MockOS{
		LookupEnvFunc: func(key string) (string, bool) {
			if key == "AGENT_CONFIG_PATH" {
				return "/home/user/.agent/config.yml", true
			}
			return "", false
		},
		ReadFileFunc: func(name string) ([]byte, error) {
			return []byte(yamlData), nil
		},
	}

	cfg, err := LoadWithOS(mock)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg == nil {
		t.Fatal("expected non-nil config")
	}

	if len(cfg.Adapters) != 0 {
		t.Errorf("expected empty adapters map, got %d", len(cfg.Adapters))
	}
}
