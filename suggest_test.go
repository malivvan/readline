package readline

import (
	"testing"
)

func TestInlineSuggestAcceptFull(t *testing.T) {
	rl := NewShell()
	storeLine(rl, "sta")
	rl.SetInlineSuggestion("status --verbose")

	rl.inlineSuggestAccept()

	if got := string(*rl.line); got != "status --verbose" {
		t.Fatalf("line after accept = %q, want %q", got, "status --verbose")
	}
	if got := rl.GetInlineSuggestion(); got != "" {
		t.Fatalf("inline suggestion after accept = %q, want cleared", got)
	}
}

func TestInlineSuggestAcceptNotApplicable(t *testing.T) {
	rl := NewShell()
	storeLine(rl, "xyz")
	rl.SetInlineSuggestion("status")

	rl.inlineSuggestAccept()

	if got := string(*rl.line); got != "xyz" {
		t.Fatalf("line = %q, want unchanged %q", got, "xyz")
	}

	storeLine(rl, "sta")
	rl.cursor.Set(1)
	rl.SetInlineSuggestion("status")

	rl.inlineSuggestAccept()

	if got := string(*rl.line); got != "sta" {
		t.Fatalf("line = %q, want unchanged %q (cursor not at end)", got, "sta")
	}
}

func TestInlineSuggestAcceptWord(t *testing.T) {
	rl := NewShell()
	storeLine(rl, "sta")
	rl.SetInlineSuggestion("status --verbose")

	rl.inlineSuggestAcceptWord()

	if got := string(*rl.line); got != "status" {
		t.Fatalf("line after word accept = %q, want %q", got, "status")
	}

	if got := rl.GetInlineSuggestion(); got != "status --verbose" {
		t.Fatalf("inline suggestion after word accept = %q, want unchanged", got)
	}

	rl.inlineSuggestAcceptWord()
	if got := string(*rl.line); got != "status --verbose" {
		t.Fatalf("line after second word accept = %q, want %q", got, "status --verbose")
	}
	if got := rl.GetInlineSuggestion(); got != "" {
		t.Fatalf("inline suggestion after full consumption = %q, want cleared", got)
	}
}

func TestInlineSuggestionAPI(t *testing.T) {
	rl := NewShell()

	rl.SetInlineSuggestion("status --verbose")
	if got := rl.GetInlineSuggestion(); got != "status --verbose" {
		t.Fatalf("inline suggestion = %q, want status --verbose", got)
	}

	rl.ClearInlineSuggestion()
	if got := rl.GetInlineSuggestion(); got != "" {
		t.Fatalf("inline suggestion after clear = %q, want empty", got)
	}
}

func storeLine(rl *Shell, s string) {
	rl.line.Set([]rune(s)...)
	rl.cursor.Set(rl.line.Len())
}
