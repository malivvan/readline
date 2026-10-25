package keymap

import (
	"testing"

	"github.com/malivvan/readline/core"
	"github.com/malivvan/readline/inputrc"
)

func TestEmacsBackspaceDeletesChar(t *testing.T) {
	keys := &core.Keys{}
	iters := &core.Iterations{}

	_, cfg := NewEngine(keys, iters)

	got := cfg.Binds[string(Emacs)][inputrc.Unescape(`\C-h`)].Action

	if got == "backward-kill-word" {
		t.Fatal(`\C-h is bound to backward-kill-word: Backspace would delete a whole word on terminals that send ^H`)
	}

	if got != "backward-delete-char" {
		t.Fatalf(`\C-h resolved to %q, want "backward-delete-char"`, got)
	}
}
