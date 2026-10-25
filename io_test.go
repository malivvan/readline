package readline

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/malivvan/readline/term"
)

func TestNewShellDefaultStreams(t *testing.T) {
	shell := NewShell()
	defer shell.SetIO(os.Stdin, os.Stdout, os.Stderr)

	if shell.Stdin != os.Stdin || shell.Stdout != os.Stdout || shell.Stderr != os.Stderr {
		t.Fatalf("default streams = (%v, %v, %v), want the process standard streams",
			shell.Stdin, shell.Stdout, shell.Stderr)
	}
}

func TestSetIOPropagatesToComponents(t *testing.T) {
	shell := NewShell()
	defer shell.SetIO(os.Stdin, os.Stdout, os.Stderr)

	var stdout, stderr bytes.Buffer

	shell.SetIO(strings.NewReader("input"), &stdout, &stderr)

	if shell.Stdout != &stdout || shell.Stderr != &stderr {
		t.Fatal("SetIO did not store the streams on the shell")
	}

	if term.Output() != &stdout {
		t.Fatal("terminal rendering does not use the shell stdout")
	}

	if _, err := shell.Printf("hello %s", "world"); err != nil {
		t.Fatalf("Printf: %v", err)
	}

	if !strings.Contains(stdout.String(), "hello world") {
		t.Fatalf("Printf output = %q, want it to contain %q", stdout.String(), "hello world")
	}
}

func TestPrintTransientfUsesShellStdout(t *testing.T) {
	shell := NewShell()
	defer shell.SetIO(os.Stdin, os.Stdout, os.Stderr)

	var stdout bytes.Buffer

	shell.SetIO(nil, &stdout, nil)

	if _, err := shell.PrintTransientf("transient %d", 42); err != nil {
		t.Fatalf("PrintTransientf: %v", err)
	}

	if !strings.Contains(stdout.String(), "transient 42") {
		t.Fatalf("PrintTransientf output = %q, want it to contain %q", stdout.String(), "transient 42")
	}
}

func TestSetIONilRestoresDefaults(t *testing.T) {
	shell := NewShell()
	defer shell.SetIO(os.Stdin, os.Stdout, os.Stderr)

	var stdout bytes.Buffer

	shell.SetIO(nil, &stdout, nil)
	shell.SetIO(nil, nil, nil)

	if shell.Stdout != os.Stdout || shell.Stderr != os.Stderr || shell.Stdin != os.Stdin {
		t.Fatal("nil streams should restore the process standard streams")
	}
}
