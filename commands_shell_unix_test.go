//go:build unix

package readline

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestEditCommandWithEmptiedBufferDoesNotPanic guards the editor result
// handling: an editor that deletes the whole buffer returns an empty result
// with a nil error, which used to be treated as an error and dereferenced.
func TestEditCommandWithEmptiedBufferDoesNotPanic(t *testing.T) {
	script := filepath.Join(t.TempDir(), "empty-editor")

	if err := os.WriteFile(script, []byte("#!/bin/sh\n: > \"$1\"\n"), 0o755); err != nil {
		t.Fatalf("write editor script: %v", err)
	}

	t.Setenv("VISUAL", "")
	t.Setenv("EDITOR", script)
	t.Setenv("TMPDIR", t.TempDir())

	shell := NewShell()
	shell.Line().Set([]rune("hello")...)

	shell.editAndExecuteCommand()

	if got := shell.Hint.Text(); !strings.Contains(got, "Editor error") {
		t.Fatalf("hint = %q, want an editor error hint", got)
	}

	if got := string(*shell.Line()); got != "hello" {
		t.Fatalf("line = %q, want the original buffer to be kept", got)
	}
}
