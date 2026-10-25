//go:build unix

package editor

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeRecorder creates an executable shell script that records its arguments
// in $EDITOR_ARGS and leaves the edited file untouched.
func writeRecorder(t *testing.T, name string) (dir, script string) {
	t.Helper()

	dir = t.TempDir()
	script = filepath.Join(dir, name)

	body := "#!/bin/sh\nprintf '%s\\n' \"$@\" > \"$EDITOR_ARGS\"\n"

	if err := os.WriteFile(script, []byte(body), 0o755); err != nil {
		t.Fatalf("write editor script: %v", err)
	}

	return dir, script
}

func editWithEditor(t *testing.T, editor, filetype string) []string {
	t.Helper()

	argsFile := filepath.Join(t.TempDir(), "args")
	t.Setenv("VISUAL", "")
	t.Setenv("EDITOR", editor)
	t.Setenv("EDITOR_ARGS", argsFile)
	t.Setenv("TMPDIR", t.TempDir())

	buffers := NewBuffers()

	edited, err := buffers.EditBuffer([]rune("hello"), "", filetype, false)
	if err != nil {
		t.Fatalf("EditBuffer: %v", err)
	}

	if string(edited) != "hello" {
		t.Fatalf("edited buffer = %q, want %q", string(edited), "hello")
	}

	raw, err := os.ReadFile(argsFile)
	if err != nil {
		t.Fatalf("read recorded arguments: %v", err)
	}

	return strings.Split(strings.TrimRight(string(raw), "\n"), "\n")
}

func TestEditBufferSplitsEditorArguments(t *testing.T) {
	_, script := writeRecorder(t, "fake-editor")

	args := editWithEditor(t, script+" --marker", "")

	if len(args) != 2 || args[0] != "--marker" {
		t.Fatalf("editor arguments = %q, want [--marker <file>]", args)
	}

	if !strings.HasPrefix(args[1], os.TempDir()) {
		t.Fatalf("edited file %q is not in the temp directory", args[1])
	}
}

func TestEditBufferPassesFiletypeToVi(t *testing.T) {
	_, script := writeRecorder(t, "vi")

	args := editWithEditor(t, script, "go")

	if len(args) != 3 || args[0] != "-c" || args[1] != "set filetype=go" {
		t.Fatalf("editor arguments = %q, want [-c, set filetype=go, <file>]", args)
	}
}

func TestEditBufferDoesNotPassFiletypeToOtherEditors(t *testing.T) {
	_, script := writeRecorder(t, "nano")

	args := editWithEditor(t, script, "go")

	if len(args) != 1 {
		t.Fatalf("editor arguments = %q, want only the file name", args)
	}
}
