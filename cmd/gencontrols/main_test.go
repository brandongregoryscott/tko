package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGeneratedFilesAreCurrent(t *testing.T) {
	if err := run("../..", true); err != nil {
		t.Fatal(err)
	}
}

func TestConfigValidation(t *testing.T) {
	data, err := os.ReadFile("../../controls.json")
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct{ name, old, replacement, want string }{
		{"duplicate key", `"end"`, `"ctrl+e"`, "shared"},
		{"unknown help action", `"CursorUp",`, `"Missing",`, "unknown"},
		{"unknown field", `"sections":`, `"sectons":`, "unknown field"},
		{"empty key", `"end"`, `""`, "empty key"},
		{"duplicate ID", `"id": "CursorDown"`, `"id": "CursorUp"`, "duplicate action"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := parseConfig([]byte(strings.Replace(string(data), tc.old, tc.replacement, 1)))
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("got %v, want error containing %q", err, tc.want)
			}
		})
	}
}

func TestGenerationAndCheck(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "internal/ui"), 0755); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile("../../controls.json")
	if err != nil {
		t.Fatal(err)
	}
	// A config edit must update both runtime bindings and documentation.
	data = []byte(strings.Replace(string(data), `"ctrl+s"`, `"ctrl+w"`, 1))
	if err := os.WriteFile(filepath.Join(root, "controls.json"), data, 0644); err != nil {
		t.Fatal(err)
	}
	readme := "handwritten introduction\n" + begin + "\nold tables\n" + end + "\nhandwritten ending\n"
	if err := os.WriteFile(filepath.Join(root, "README.md"), []byte(readme), 0644); err != nil {
		t.Fatal(err)
	}
	if err := run(root, true); err == nil {
		t.Fatal("check accepted stale outputs")
	}
	// Check mode must leave files untouched.
	unchanged, _ := os.ReadFile(filepath.Join(root, "README.md"))
	if string(unchanged) != readme {
		t.Fatal("check modified README")
	}
	if err := run(root, false); err != nil {
		t.Fatal(err)
	}
	code, _ := os.ReadFile(filepath.Join(root, "internal/ui/keymap.go"))
	if !strings.Contains(string(code), `key.WithKeys("ctrl+w")`) {
		t.Fatal("runtime key was not updated")
	}
	doc, _ := os.ReadFile(filepath.Join(root, "README.md"))
	if !strings.Contains(string(doc), "<code>Ctrl+W</code> | Save project") || !strings.HasPrefix(string(doc), "handwritten introduction\n") || !strings.HasSuffix(string(doc), "handwritten ending\n") {
		t.Fatalf("incorrect README: %s", doc)
	}
	if err := run(root, true); err != nil {
		t.Fatal(err)
	}
	if err := run(root, false); err != nil {
		t.Fatal(err)
	}
	after, _ := os.ReadFile(filepath.Join(root, "README.md"))
	if string(after) != string(doc) {
		t.Fatal("generation is not idempotent")
	}
}

func TestReadmeMarkersAndEscaping(t *testing.T) {
	for _, doc := range []string{"no markers", end + begin, begin + end + begin} {
		if _, err := updateReadme(doc, "tables"); err == nil {
			t.Fatalf("accepted malformed markers: %s", doc)
		}
	}
	c := config{Sections: []section{{Title: "Test", Bindings: []binding{{Keys: []string{"|", "<", "`"}, Description: "A | B"}}}}}
	markdown := renderMarkdown(c)
	for _, want := range []string{"<code>&#124;</code>", "<code>&lt;</code>", "<code>`</code>", "A &#124; B"} {
		if !strings.Contains(markdown, want) {
			t.Errorf("missing escaped content %q", want)
		}
	}
}
