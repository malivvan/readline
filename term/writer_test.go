package term

import (
	"bytes"
	"io"
	"os"
	"testing"
)

type tallyWriter struct {
	writes int
	buf    bytes.Buffer
}

func (c *tallyWriter) Write(p []byte) (int, error) {
	c.writes++

	return c.buf.Write(p)
}

func TestWriteStringDirect(t *testing.T) {
	cw := &tallyWriter{}
	renderOut = cw

	defer func() { renderOut = nil }()

	WriteString("hello")
	WriteString("")
	WriteString(" world")

	if got := cw.buf.String(); got != "hello world" {
		t.Fatalf("content = %q, want %q", got, "hello world")
	}

	if cw.writes != 2 {
		t.Fatalf("direct writes = %d, want 2", cw.writes)
	}
}

func TestBufferCoalescesToSingleWrite(t *testing.T) {
	cw := &tallyWriter{}
	renderOut = cw

	defer func() { renderOut = nil }()

	BeginBuffer()
	WriteString("\x1b[?25l")
	WriteString("prompt> ")
	WriteString("line")

	if cw.writes != 0 {
		t.Fatalf("expected nothing on screen before EndBuffer, got %d writes", cw.writes)
	}

	EndBuffer()

	if cw.writes != 1 {
		t.Fatalf("expected a single coalesced write, got %d", cw.writes)
	}

	if got := cw.buf.String(); got != "\x1b[?25lprompt> line" {
		t.Fatalf("content = %q", got)
	}
}

func TestBufferNestingFlushesOnce(t *testing.T) {
	cw := &tallyWriter{}
	renderOut = cw

	defer func() { renderOut = nil }()

	BeginBuffer()
	WriteString("a")
	BeginBuffer()
	WriteString("b")
	EndBuffer()

	if cw.writes != 0 {
		t.Fatalf("inner EndBuffer flushed early: %d writes", cw.writes)
	}

	WriteString("c")
	EndBuffer()

	if cw.writes != 1 {
		t.Fatalf("expected one flush, got %d", cw.writes)
	}

	if got := cw.buf.String(); got != "abc" {
		t.Fatalf("content = %q, want abc", got)
	}
}

func TestFlushBufferMidFrame(t *testing.T) {
	cw := &tallyWriter{}
	renderOut = cw

	defer func() { renderOut = nil }()

	BeginBuffer()
	WriteString("prompt")
	FlushBuffer()

	if cw.writes != 1 || cw.buf.String() != "prompt" {
		t.Fatalf("FlushBuffer did not flush partial frame: writes=%d content=%q", cw.writes, cw.buf.String())
	}

	WriteString("line")
	EndBuffer()

	if cw.writes != 2 {
		t.Fatalf("expected 2 writes (mid-flush + end), got %d", cw.writes)
	}

	if got := cw.buf.String(); got != "promptline" {
		t.Fatalf("content = %q", got)
	}
}

func TestSetOutput(t *testing.T) {
	cw := &tallyWriter{}

	SetOutput(cw)
	t.Cleanup(func() { SetOutput(nil) })

	WriteString("hello")

	if got := cw.buf.String(); got != "hello" {
		t.Fatalf("content = %q, want %q", got, "hello")
	}

	if Output() != io.Writer(cw) {
		t.Fatal("Output() does not return the configured writer")
	}
}

func TestOutputDefaultsToStdout(t *testing.T) {
	SetOutput(nil)
	t.Cleanup(func() { SetOutput(nil) })

	if Output() != os.Stdout {
		t.Fatal("Output() should default to os.Stdout")
	}
}
