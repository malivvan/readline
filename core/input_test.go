package core

import (
	"bytes"
	"errors"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/malivvan/readline/inputrc"
)

var errReadFailed = errors.New("read failed")

type installStep struct {
	data []byte
	err  error
}

type stubReadCloser struct {
	steps []installStep
	index int
}

func (s *stubReadCloser) Read(p []byte) (int, error) {
	if s.index >= len(s.steps) {
		return 0, io.EOF
	}

	step := s.steps[s.index]
	s.index++

	n := copy(p, step.data)

	return n, step.err
}

func (s *stubReadCloser) Close() error { return nil }

func stubKeys(t *testing.T, steps ...installStep) *Keys {
	t.Helper()

	keys := &Keys{}
	keys.SetInput(&stubReadCloser{steps: steps})

	return keys
}

func TestWaitAvailableKeysMarksEOF(t *testing.T) {
	keys := stubKeys(t, installStep{err: io.EOF})
	WaitAvailableKeys(keys, &inputrc.Config{})

	if !keys.IsEOF() {
		t.Fatal("expected EOF to be recorded")
	}

	if err := keys.ReadError(); err != nil {
		t.Fatalf("expected no read error on EOF, got %v", err)
	}
}

func TestWaitAvailableKeysRecordsNonEOFReadError(t *testing.T) {
	want := errReadFailed

	keys := stubKeys(t, installStep{err: want})
	WaitAvailableKeys(keys, &inputrc.Config{})

	if keys.IsEOF() {
		t.Fatal("non-EOF failure must not be marked as EOF")
	}

	if err := keys.ReadError(); !errors.Is(err, want) {
		t.Fatalf("expected read error %v, got %v", want, err)
	}
}

func TestReadKeyReturnsAbortOnEOF(t *testing.T) {
	keys := stubKeys(t, installStep{err: io.EOF})

	key, isAbort := keys.ReadKey()

	if !isAbort {
		t.Fatal("expected ReadKey to abort on EOF")
	}

	if key != 0 {
		t.Fatalf("expected zero key on EOF, got %q", key)
	}
}

func TestReadKeySkipsEmptyReads(t *testing.T) {
	keys := stubKeys(t, installStep{}, installStep{data: []byte("x")})

	key, isAbort := keys.ReadKey()

	if isAbort {
		t.Fatal("expected ReadKey to continue after an empty read")
	}

	if key != 'x' {
		t.Fatalf("expected key %q, got %q", 'x', key)
	}
}

func TestKeysStreamDefaults(t *testing.T) {
	keys := &Keys{}

	if keys.stdin() != os.Stdin {
		t.Fatal("default input should be os.Stdin")
	}

	if keys.outputWriter() != os.Stdout {
		t.Fatal("default query output should be os.Stdout")
	}

	if keys.stderrWriter() != os.Stderr {
		t.Fatal("default error output should be os.Stderr")
	}
}

func TestKeysSetStreams(t *testing.T) {
	keys := &Keys{}

	var stdout, stderr bytes.Buffer
	input := strings.NewReader("")

	keys.SetInput(input)
	keys.SetOutput(&stdout)
	keys.SetErrorOutput(&stderr)

	if keys.stdin() != input {
		t.Fatal("input was not stored")
	}

	if keys.outputWriter() != &stdout {
		t.Fatal("query output was not stored")
	}

	if keys.stderrWriter() != &stderr {
		t.Fatal("error output was not stored")
	}
}
