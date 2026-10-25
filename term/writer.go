package term

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"sync"
)

var (
	outMu    sync.Mutex
	outDepth int
	outBuf   *bufio.Writer
)

// renderOut is the writer used for all rendering. When nil, os.Stdout is used.
var renderOut io.Writer

// SetOutput sets the writer used for all terminal rendering.
// A nil writer restores the default (os.Stdout).
func SetOutput(w io.Writer) {
	outMu.Lock()
	defer outMu.Unlock()

	renderOut = w
}

// Output returns the writer used for terminal rendering.
func Output() io.Writer {
	outMu.Lock()
	defer outMu.Unlock()

	if renderOut != nil {
		return renderOut
	}

	return os.Stdout
}

func BeginBuffer() {
	outMu.Lock()
	defer outMu.Unlock()

	outDepth++
	if outDepth == 1 {
		outBuf = bufio.NewWriterSize(out(), 64*1024)
	}
}

func EndBuffer() {
	outMu.Lock()
	defer outMu.Unlock()

	if outDepth == 0 {
		return
	}

	outDepth--

	if outDepth == 0 && outBuf != nil {
		_ = outBuf.Flush()
		outBuf = nil
	}
}

func FlushBuffer() {
	outMu.Lock()
	defer outMu.Unlock()

	if outBuf != nil {
		_ = outBuf.Flush()
	}
}

func Printf(format string, a ...interface{}) {
	WriteString(fmt.Sprintf(format, a...))
}

func WriteString(s string) {
	if s == "" {
		return
	}

	outMu.Lock()
	defer outMu.Unlock()

	if outDepth > 0 && outBuf != nil {
		_, _ = io.WriteString(outBuf, s)
		return
	}

	_, _ = io.WriteString(out(), s)
}

func out() io.Writer {
	if renderOut != nil {
		return renderOut
	}

	return os.Stdout
}

func printf(format string, a ...interface{}) {
	WriteString(fmt.Sprintf(format, a...))
}
