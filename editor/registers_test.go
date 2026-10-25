package editor

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestBuffersStreamDefaults(t *testing.T) {
	reg := NewBuffers()

	if reg.input() != os.Stdin {
		t.Fatal("default input should be os.Stdin")
	}

	if reg.output() != os.Stdout {
		t.Fatal("default output should be os.Stdout")
	}

	if reg.errorOutput() != os.Stderr {
		t.Fatal("default error output should be os.Stderr")
	}
}

func TestBuffersSetIO(t *testing.T) {
	reg := NewBuffers()

	var stdout, stderr bytes.Buffer
	input := strings.NewReader("")

	reg.SetIO(input, &stdout, &stderr)

	if reg.input() != input {
		t.Fatal("input was not stored")
	}

	if reg.output() != &stdout {
		t.Fatal("output was not stored")
	}

	if reg.errorOutput() != &stderr {
		t.Fatal("error output was not stored")
	}
}
