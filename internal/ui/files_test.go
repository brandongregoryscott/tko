package ui

import (
	"testing"
)

func TestSanitizeFilename(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"my-project", "my-project"},
		{"my/project", "my"},          // truncates at /
		{"name with spaces", "name with spaces"},
		{"", ""},
		{"normal.mp3", "normal"},      // truncates at .
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := sanitizeFilename(tt.input)
			if got != tt.expected {
				t.Errorf("sanitizeFilename(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestSanitizeFilenameNoUnsafeChars(t *testing.T) {
	// Strings without unsafe chars should pass through unchanged.
	if got := sanitizeFilename("my-beat"); got != "my-beat" {
		t.Errorf("got %q, want my-beat", got)
	}
}

func TestSanitizeFilenameEmpty(t *testing.T) {
	if got := sanitizeFilename(""); got != "" {
		t.Errorf("empty: got %q", got)
	}
	if got := sanitizeFilename("."); got != "" {
		t.Errorf("just a dot: got %q", got)
	}
}
