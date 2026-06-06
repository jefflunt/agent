package runner

import "testing"

func TestPlaceholder(t *testing.T) {
	got := Placeholder()
	want := "runner"
	if got != want {
		t.Errorf("Placeholder() = %q; want %q", got, want)
	}
}
