package config

import "testing"

func TestPlaceholder(t *testing.T) {
	got := Placeholder()
	want := "config"
	if got != want {
		t.Errorf("Placeholder() = %q; want %q", got, want)
	}
}
