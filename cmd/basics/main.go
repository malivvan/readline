// Command basics is the smallest useful readline program: a prompt, a reading
// loop, and a line of feedback for every accepted input.
//
// It also shows the two errors a shell loop must handle: Ctrl-C aborts the
// current line (and returns ErrInterrupt), Ctrl-D on an empty line returns
// io.EOF and should end the program.
package main

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/malivvan/readline"
)

func main() {
	shell := readline.NewShell()

	// The primary prompt is recomputed on every redisplay.
	shell.Prompt.Primary(func() string {
		return "basics> "
	})

	for {
		line, err := shell.Readline()

		switch {
		case errors.Is(err, readline.ErrInterrupt):
			fmt.Println("^C - press Ctrl-D to quit")
			continue

		case errors.Is(err, io.EOF):
			fmt.Println("bye")
			return

		case err != nil:
			fmt.Fprintf(os.Stderr, "readline: %v\n", err)
			return
		}

		fmt.Printf("you typed: %q\n", line)
	}
}
