package adapter

import (
	"errors"
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name        string
		target      string
		wantAdapter *Adapter
		wantErrIs   error
	}{
		{
			name:   "valid target",
			target: "opencode:google/gemini-3.5-flash",
			wantAdapter: &Adapter{
				CLIName:  "opencode",
				Provider: "google",
				Model:    "gemini-3.5-flash",
			},
			wantErrIs: nil,
		},
		{
			name:   "valid target with leading/trailing spaces on target",
			target: "  opencode:google/gemini-3.5-flash  ",
			wantAdapter: &Adapter{
				CLIName:  "opencode",
				Provider: "google",
				Model:    "gemini-3.5-flash",
			},
			wantErrIs: nil,
		},
		{
			name:   "valid target with spaces around separators",
			target: "opencode  :  google  /  gemini-3.5-flash",
			wantAdapter: &Adapter{
				CLIName:  "opencode",
				Provider: "google",
				Model:    "gemini-3.5-flash",
			},
			wantErrIs: nil,
		},
		{
			name:        "empty target string",
			target:      "",
			wantAdapter: nil,
			wantErrIs:   ErrEmptyTarget,
		},
		{
			name:        "only spaces target string",
			target:      "    ",
			wantAdapter: nil,
			wantErrIs:   ErrEmptyTarget,
		},
		{
			name:        "missing colon",
			target:      "opencode-google/gemini-3.5-flash",
			wantAdapter: nil,
			wantErrIs:   ErrMissingColon,
		},
		{
			name:        "extra colon",
			target:      "opencode:google:gemini-3.5-flash",
			wantAdapter: nil,
			wantErrIs:   ErrExtraColon,
		},
		{
			name:        "empty CLI name",
			target:      ":google/gemini-3.5-flash",
			wantAdapter: nil,
			wantErrIs:   ErrEmptyCLIName,
		},
		{
			name:        "empty CLI name with spaces",
			target:      "   :google/gemini-3.5-flash",
			wantAdapter: nil,
			wantErrIs:   ErrEmptyCLIName,
		},
		{
			name:        "missing slash",
			target:      "opencode:google-gemini-3.5-flash",
			wantAdapter: nil,
			wantErrIs:   ErrMissingSlash,
		},
		{
			name:        "extra slash",
			target:      "opencode:google/gemini/3.5-flash",
			wantAdapter: nil,
			wantErrIs:   ErrExtraSlash,
		},
		{
			name:        "empty provider",
			target:      "opencode:/gemini-3.5-flash",
			wantAdapter: nil,
			wantErrIs:   ErrEmptyProvider,
		},
		{
			name:        "empty provider with spaces",
			target:      "opencode:  /gemini-3.5-flash",
			wantAdapter: nil,
			wantErrIs:   ErrEmptyProvider,
		},
		{
			name:        "empty model",
			target:      "opencode:google/",
			wantAdapter: nil,
			wantErrIs:   ErrEmptyModel,
		},
		{
			name:        "empty model with spaces",
			target:      "opencode:google/   ",
			wantAdapter: nil,
			wantErrIs:   ErrEmptyModel,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.target)
			if tt.wantErrIs != nil {
				if err == nil {
					t.Fatalf("expected error containing %v, got nil", tt.wantErrIs)
				}
				if !errors.Is(err, tt.wantErrIs) {
					t.Errorf("expected error %v to wrap %v", err, tt.wantErrIs)
				}
				if got != nil {
					t.Errorf("expected nil adapter, got %+v", got)
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if got == nil {
					t.Fatal("expected non-nil adapter, got nil")
				}
				if got.CLIName != tt.wantAdapter.CLIName {
					t.Errorf("CLIName = %q; want %q", got.CLIName, tt.wantAdapter.CLIName)
				}
				if got.Provider != tt.wantAdapter.Provider {
					t.Errorf("Provider = %q; want %q", got.Provider, tt.wantAdapter.Provider)
				}
				if got.Model != tt.wantAdapter.Model {
					t.Errorf("Model = %q; want %q", got.Model, tt.wantAdapter.Model)
				}
			}
		})
	}
}

func TestPlaceholder(t *testing.T) {
	got := Placeholder()
	want := "adapter"
	if got != want {
		t.Errorf("Placeholder() = %q; want %q", got, want)
	}
}
