package adapter

import "testing"

func TestPlaceholder(t *testing.T) {
	got := Placeholder()
	want := "adapter"
	if got != want {
		t.Errorf("Placeholder() = %q; want %q", got, want)
	}
}
