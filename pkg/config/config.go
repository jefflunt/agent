package config

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration.
type Config struct {
	Adapters map[string]string
}

// ErrConfigMissing is returned when the configuration file cannot be found.
type ErrConfigMissing struct {
	Path string
}

func (e *ErrConfigMissing) Error() string {
	return fmt.Sprintf("configuration file not found at %q. Please initialize it by creating a file with your adapter configurations, for example:\n\nadapters:\n  primary: <execution-spec>\n  fast: <execution-spec>\n", e.Path)
}

// OS defines the environment and file system operations required to load configuration.
type OS interface {
	ReadFile(name string) ([]byte, error)
	UserHomeDir() (string, error)
	LookupEnv(key string) (string, bool)
}

// RealOS implements the OS interface using the standard Go os package.
type RealOS struct{}

// ReadFile reads the file named by name and returns the contents.
func (RealOS) ReadFile(name string) ([]byte, error) {
	return os.ReadFile(name)
}

// UserHomeDir returns the current user's home directory.
func (RealOS) UserHomeDir() (string, error) {
	return os.UserHomeDir()
}

// LookupEnv retrieves the value of the environment variable named by the key.
func (RealOS) LookupEnv(key string) (string, bool) {
	return os.LookupEnv(key)
}

// DefaultOS is the standard implementation of the OS interface used by Load.
var DefaultOS OS = RealOS{}

// Load finds, reads, and parses the configuration file using the default OS implementation.
func Load() (*Config, error) {
	return LoadWithOS(DefaultOS)
}

// LoadWithOS finds, reads, and parses the configuration file using a custom OS implementation.
func LoadWithOS(sys OS) (*Config, error) {
	path, err := ResolvePath(sys)
	if err != nil {
		return nil, err
	}

	data, err := sys.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) || errors.Is(err, fs.ErrNotExist) {
			return nil, &ErrConfigMissing{Path: path}
		}
		return nil, fmt.Errorf("failed to read configuration file at %q: %w", path, err)
	}

	cfg, err := ParseConfig(data)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

// ResolvePath resolves the absolute path of the configuration file.
// It checks AGENT_CONFIG_PATH env variable, falling back to ~/.agent/config.yml.
func ResolvePath(sys OS) (string, error) {
	if path, ok := sys.LookupEnv("AGENT_CONFIG_PATH"); ok && path != "" {
		return path, nil
	}
	home, err := sys.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to resolve user home directory: %w", err)
	}
	return filepath.Join(home, ".agent", "config.yml"), nil
}

// ParseConfig parses YAML data and extracts the adapter configurations.
// It supports both a flat map structure and a nested structure under an "adapters" key.
func ParseConfig(data []byte) (*Config, error) {
	var nested struct {
		Adapters map[string]string `yaml:"adapters"`
	}
	nestedErr := yaml.Unmarshal(data, &nested)
	if nestedErr == nil && len(nested.Adapters) > 0 {
		return &Config{Adapters: nested.Adapters}, nil
	}

	var flat map[string]string
	flatErr := yaml.Unmarshal(data, &flat)
	if flatErr == nil && len(flat) > 0 {
		return &Config{Adapters: flat}, nil
	}

	if nestedErr != nil && flatErr != nil {
		return nil, fmt.Errorf("malformed YAML: %v", nestedErr)
	}

	return &Config{Adapters: make(map[string]string)}, nil
}

// Placeholder is a temporary function to ensure the package is correctly compiled and imported.
func Placeholder() string {
	fmt.Println("pkg/config placeholder called")
	return "config"
}
