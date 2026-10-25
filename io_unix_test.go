//go:build unix

package readline

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/creack/pty"
)

// TestReadlineReturnsWhenOutputIsNotTerminal reproduces a hang: the
// cursor-position probe used to be written to the configured output even when
// that output was a file, while the answer was awaited on the terminal input.
// The answer never arrives, so the shell blocked forever after printing the
// prompt.
func TestReadlineReturnsWhenOutputIsNotTerminal(t *testing.T) {
	ptmx, tty, err := pty.Open()
	if err != nil {
		t.Fatalf("pty: %v", err)
	}
	defer func() { _ = ptmx.Close() }()
	defer func() { _ = tty.Close() }()

	out, err := os.Create(filepath.Join(t.TempDir(), "stdout"))
	if err != nil {
		t.Fatalf("create output file: %v", err)
	}
	defer func() { _ = out.Close() }()

	shell := NewShell()
	defer shell.SetIO(os.Stdin, os.Stdout, os.Stderr)

	shell.SetIO(tty, out, out)
	shell.Prompt.Primary(func() string { return "> " })

	type result struct {
		line string
		err  error
	}

	done := make(chan result, 1)

	go func() {
		line, err := shell.Readline()
		done <- result{line: line, err: err}
	}()

	time.Sleep(200 * time.Millisecond)

	if _, err := ptmx.WriteString("abc\r"); err != nil {
		t.Fatalf("write input: %v", err)
	}

	select {
	case got := <-done:
		if got.err != nil {
			t.Fatalf("Readline: %v", got.err)
		}

		if got.line != "abc" {
			t.Fatalf("line = %q, want %q", got.line, "abc")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Readline blocked waiting for a cursor-position reply although stdout is not a terminal")
	}
}
