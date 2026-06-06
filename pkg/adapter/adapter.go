package adapter

import (
	"errors"
	"fmt"
	"strings"
)

// Adapter represents the parsed components of an adapter configuration target.
type Adapter struct {
	CLIName  string
	Provider string
	Model    string
}

// Custom error types or sentinel errors representing parsing issues.
var (
	ErrEmptyTarget   = errors.New("target configuration string is empty")
	ErrMissingColon  = errors.New("missing colon separator (':')")
	ErrExtraColon    = errors.New("extra colon separator (':')")
	ErrMissingSlash  = errors.New("missing slash separator ('/')")
	ErrExtraSlash    = errors.New("extra slash separator ('/')")
	ErrEmptyCLIName  = errors.New("CLI name component is empty")
	ErrEmptyProvider = errors.New("provider component is empty")
	ErrEmptyModel    = errors.New("model component is empty")
)

// Parse parses a target configuration string in the format "cliName:provider/model"
// and returns the parsed components. If the format is invalid or missing components,
// it returns a formatted error wrapping one of the sentinel error values.
func Parse(target string) (*Adapter, error) {
	trimmedTarget := strings.TrimSpace(target)
	if trimmedTarget == "" {
		return nil, fmt.Errorf("invalid target: %w", ErrEmptyTarget)
	}

	colonCount := strings.Count(trimmedTarget, ":")
	if colonCount == 0 {
		return nil, fmt.Errorf("invalid target %q: %w", target, ErrMissingColon)
	}
	if colonCount > 1 {
		return nil, fmt.Errorf("invalid target %q: %w", target, ErrExtraColon)
	}

	parts := strings.Split(trimmedTarget, ":")
	cliName := strings.TrimSpace(parts[0])
	providerModelPart := strings.TrimSpace(parts[1])

	if cliName == "" {
		return nil, fmt.Errorf("invalid target %q: %w", target, ErrEmptyCLIName)
	}

	slashCount := strings.Count(providerModelPart, "/")
	if slashCount == 0 {
		return nil, fmt.Errorf("invalid target %q: %w", target, ErrMissingSlash)
	}
	if slashCount > 1 {
		return nil, fmt.Errorf("invalid target %q: %w", target, ErrExtraSlash)
	}

	subParts := strings.Split(providerModelPart, "/")
	provider := strings.TrimSpace(subParts[0])
	model := strings.TrimSpace(subParts[1])

	if provider == "" {
		return nil, fmt.Errorf("invalid target %q: %w", target, ErrEmptyProvider)
	}
	if model == "" {
		return nil, fmt.Errorf("invalid target %q: %w", target, ErrEmptyModel)
	}

	return &Adapter{
		CLIName:  cliName,
		Provider: provider,
		Model:    model,
	}, nil
}

// Placeholder is a temporary function to ensure the package is correctly compiled and imported.
func Placeholder() string {
	fmt.Println("pkg/adapter placeholder called")
	return "adapter"
}
