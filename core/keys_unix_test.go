//go:build unix

package core

import (
	"testing"

	"github.com/creack/pty"

	"github.com/malivvan/readline/term"
)

func TestGetCursorPosPreservesCoalescedInput(t *testing.T) {
	ptmx, tty, err := pty.Open()
	if err != nil {
		t.Fatalf("pty: %v", err)
	}
	defer func() { _ = ptmx.Close() }()
	defer func() { _ = tty.Close() }()

	if _, err := term.MakeRaw(int(tty.Fd())); err != nil {
		t.Fatalf("raw mode: %v", err)
	}

	if _, err := ptmx.Write([]byte("echo hello\r\x1b[5;10R")); err != nil {
		t.Fatalf("write: %v", err)
	}

	keys := &Keys{}
	keys.SetInput(tty)
	// The cursor query is answered by the terminal that receives it, so the
	// output must be the pty as well.
	keys.SetOutput(tty)

	x, y := keys.GetCursorPos()
	if x != 10 || y != 5 {
		t.Errorf("cursor position = (%d, %d), want (10, 5)", x, y)
	}
	if got := string(keys.buf); got != "echo hello\r" {
		t.Errorf("pending input = %q, want %q (typed bytes must survive the probe)", got, "echo hello\r")
	}
}
