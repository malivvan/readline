// Command paste shows the PasteTransformer hook.
//
// With the "enable-bracketed-paste" option on, the terminal delivers pasted
// text as one chunk. Readline normalizes CRLF/CR to "\n" and then hands the
// payload to PasteTransformer, whose return value is what actually gets
// inserted.
//
// This example replaces multi-line pastes with a compact placeholder, keeps the
// original text aside, and expands the placeholder when the line is accepted.
package main

import (
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/malivvan/readline"
	"github.com/malivvan/readline/color"
)

type pasteStore struct {
	mu     sync.Mutex
	count  int
	pastes map[string]string
}

func (p *pasteStore) transform(text string) string {
	if !strings.Contains(text, "\n") {
		return text
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	p.count++

	reference := fmt.Sprintf("[Pasted text #%d +%d lines]", p.count, strings.Count(text, "\n"))
	p.pastes[reference] = text

	return reference
}

func (p *pasteStore) expand(line string) string {
	p.mu.Lock()
	defer p.mu.Unlock()

	for reference, text := range p.pastes {
		line = strings.ReplaceAll(line, reference, text)
	}

	return line
}

func main() {
	shell := readline.NewShell()

	shell.Prompt.Primary(func() string {
		return "paste> "
	})

	// Bracketed paste is opt-in: the terminal only wraps pastes when the shell
	// asks for it, which this option makes Readline do.
	_ = shell.Config.Set("enable-bracketed-paste", true)

	store := &pasteStore{pastes: make(map[string]string)}

	// Returning an empty string would drop the paste entirely.
	shell.PasteTransformer = store.transform

	fmt.Println(color.Dim + "paste a multi-line block (e.g. from your clipboard); it becomes a placeholder" + color.Reset)

	for {
		line, err := shell.Readline()

		switch {
		case errors.Is(err, readline.ErrInterrupt):
			continue

		case errors.Is(err, io.EOF):
			return

		case err != nil:
			fmt.Printf("error: %v\n", err)
			return
		}

		if line == "quit" {
			return
		}

		fmt.Printf("--- expanded line ---\n%s\n", store.expand(line))
	}
}
