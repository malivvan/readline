package readline

import (
	"maps"
	"strings"
	"testing"

	"github.com/malivvan/readline/core"
	"github.com/malivvan/readline/history"
)

func TestArrowNavigationCommandsRegistered(t *testing.T) {
	all := allCommands()

	for _, action := range []string{
		"up-line-or-search",
		"down-line-or-search",
		"up-line-or-history",
		"down-line-or-history",
		"down-line-or-select",
	} {
		if _, ok := all[action]; !ok {
			t.Errorf("navigation action %q is bound in the default keymaps but has no registered command", action)
		}
	}
}

func TestBracketedPasteNormalizesCarriageReturns(t *testing.T) {
	cases := []struct {
		name    string
		payload string
		want    string
	}{
		{"crlf", "a\r\nb", "a\nb"},
		{"bare cr", "a\rb", "a\nb"},
		{"mixed", "one\r\ntwo\rthree", "one\ntwo\nthree"},
		{"no newline", "plain", "plain"},
		{"already lf", "a\nb", "a\nb"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := supplyPaste(t, c.payload); got != c.want {
				t.Fatalf("paste %q => %q, want %q", c.payload, got, c.want)
			}
		})
	}
}

func TestBracketedPasteTransformer(t *testing.T) {
	got := supplyPasteWithTransformer(t, "a\r\nb", func(text string) string {
		if text != "a\nb" {
			t.Fatalf("transformer saw %q, want normalized text", text)
		}

		return "rewritten"
	})

	if got != "rewritten" {
		t.Fatalf("transformed paste = %q, want rewritten", got)
	}
}

func TestBracketedPasteTransformerCanDropText(t *testing.T) {
	got := supplyPasteWithTransformer(t, "secret", func(string) string {
		return ""
	})

	if got != "" {
		t.Fatalf("dropped paste = %q, want empty line", got)
	}
}

func TestCommandCompletionNilCompleter(t *testing.T) {
	rl := NewShell()
	rl.Completer = nil

	values := rl.commandCompletion()

	if !values.Messages.IsEmpty() {
		t.Fatalf("nil completer produced messages %v, want none", values.Messages.Get())
	}
}

func TestCommandCompletionRecoversFromPanic(t *testing.T) {
	rl := NewShell()
	rl.Completer = func([]rune, int) Completions {
		panic("boom")
	}

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("commandCompletion did not recover, panicked with: %v", r)
		}
	}()

	values := rl.commandCompletion()

	if values.Messages.IsEmpty() {
		t.Fatal("recovered completion produced no message, want a completion error message")
	}

	var found bool
	for _, msg := range values.Messages.Get() {
		if strings.Contains(msg, "completion error") && strings.Contains(msg, "boom") {
			found = true
		}
	}

	if !found {
		t.Fatalf("recovered messages = %v, want one containing the panic value", values.Messages.Get())
	}
}

func TestSkipCsiSequence(t *testing.T) {
	cases := []struct {
		name     string
		feed     string
		wantLeft string
	}{
		{"function key params and final", "15~", ""},
		{"single final byte", "Z", ""},
		{"modified arrow with params", "1;5D", ""},
		{"stops at final, keeps following keys", "15~abc", "abc"},
		{"empty buffer is a no-op", "", ""},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			rl := &Shell{Keys: new(core.Keys), History: &history.Sources{}}

			if c.feed != "" {
				rl.Keys.Feed(false, []rune(c.feed)...)
			}

			rl.skipCsiSequence()

			if got := drainKeys(rl); got != c.wantLeft {
				t.Fatalf("skip of %q left %q, want %q", c.feed, got, c.wantLeft)
			}
		})
	}
}

func allCommands() commands {
	rl := new(Shell)

	all := make(commands)
	for _, set := range []commands{
		rl.standardCommands(),
		rl.viCommands(),
		rl.historyCommands(),
		rl.completionCommands(),
	} {
		maps.Copy(all, set)
	}

	return all
}

func drainKeys(rl *Shell) string {
	var b []byte

	for {
		key, empty := core.PopKey(rl.Keys)
		if empty {
			return string(b)
		}

		b = append(b, key)
	}
}

func supplyPaste(t *testing.T, payload string) string {
	t.Helper()

	return supplyPasteWithTransformer(t, payload, nil)
}

func supplyPasteWithTransformer(t *testing.T, payload string, transformer func(string) string) string {
	t.Helper()

	line := new(core.Line)
	rl := &Shell{
		Keys:   new(core.Keys),
		line:   line,
		cursor: core.NewCursor(line),
	}
	rl.PasteTransformer = transformer

	rl.Keys.Feed(false, []rune(payload+"\x1b[201~")...)
	rl.bracketedPasteBegin()

	return string(*line)
}
