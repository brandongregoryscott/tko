package ui

import (
	"strings"
	"testing"
	"time"
)

func TestSanitizeFilename(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"my-project", "my-project"},
		{"my/project", "my"}, // truncates at /
		{"name with spaces", "name with spaces"},
		{"", ""},
		{"normal.mp3", "normal"}, // truncates at .
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

func TestNextProjectName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		existing []string
		expected string
	}{
		{"unused name is kept", "gen-boombap-11", nil, "gen-boombap-11"},
		{"trailing number increments", "gen-boombap-11", []string{"gen-boombap-11"}, "gen-boombap-12"},
		{
			"skips names already taken",
			"gen-boombap-11",
			[]string{"gen-boombap-11", "gen-boombap-12", "gen-boombap-13"},
			"gen-boombap-14",
		},
		{"no separator needed", "beat2", []string{"beat2"}, "beat3"},
		{"zero padding is preserved", "beat-009", []string{"beat-009"}, "beat-010"},
		{"all digits", "12", []string{"12"}, "13"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := nextProjectName(tt.input, existsIn(tt.existing))
			if got != tt.expected {
				t.Errorf("nextProjectName(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestNextProjectNameFallsBackToTimestamp(t *testing.T) {
	got := nextProjectName("my-beat", existsIn([]string{"my-beat"}))
	if !strings.HasPrefix(got, "my-beat-") {
		t.Fatalf("got %q, want a my-beat- prefixed name", got)
	}
	if _, err := time.Parse("2006-01-02-150405", strings.TrimPrefix(got, "my-beat-")); err != nil {
		t.Errorf("nextProjectName(%q) = %q, want a timestamp suffix: %v", "my-beat", got, err)
	}
}

func existsIn(names []string) func(string) bool {
	return func(name string) bool {
		for _, n := range names {
			if n == name {
				return true
			}
		}
		return false
	}
}
