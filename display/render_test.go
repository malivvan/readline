//go:build unix

package display_test

import (
	"strings"
	"testing"
	"time"

	"github.com/hinshun/vt10x"
)

func tallyLine(screen, prefix string) int {
	var n int

	for _, row := range strings.Split(screen, "\n") {
		if strings.HasPrefix(row, prefix) {
			n++
		}
	}

	return n
}

func TestRenderPromptAndInput(t *testing.T) {
	c := craftConsole(t, "PROMPT> ", 80, 24)

	c.waitForScreen("PROMPT>")

	c.send("hello world")
	screen := c.waitForScreen("PROMPT> hello world")

	firstLine := strings.SplitN(screen, "\n", 2)[0]
	if got, want := strings.TrimRight(firstLine, " "), "PROMPT> hello world"; got != want {
		t.Fatalf("first line misaligned:\n  got:  %q\n  want: %q", got, want)
	}

	c.send("\r")
	c.waitForScreen("[LINE:hello world]")
}

func TestRenderWrappingAlignment(t *testing.T) {

	c := craftConsole(t, "PROMPT> ", 40, 24)
	c.waitForScreen("PROMPT>")

	c.send(strings.Repeat("x", 40))
	screen := c.waitUntil(func(s string) bool {
		return strings.Count(s, "x") >= 40
	})

	rows := strings.Split(screen, "\n")

	if got := tallyLine(screen, "PROMPT>"); got != 1 {
		t.Fatalf("prompt should render exactly once, got %d times:\n%s", got, screen)
	}

	if want := "PROMPT> " + strings.Repeat("x", 32); strings.TrimRight(rows[0], " ") != want {
		t.Fatalf("row 0 wrong:\n  got:  %q\n  want: %q", strings.TrimRight(rows[0], " "), want)
	}

	if want := strings.Repeat("x", 8); strings.TrimRight(rows[1], " ") != want {
		t.Fatalf("row 1 (wrapped continuation) wrong:\n  got:  %q\n  want: %q",
			strings.TrimRight(rows[1], " "), want)
	}
}

func TestRenderMultilinePromptAtBottom(t *testing.T) {
	const rows = 10

	c := beginConsole(t, consoleConfig{
		prompt:  "L1\nL2\nL3\nL4> ",
		cols:    20,
		rows:    rows,
		prefill: rows - 1,
	})

	c.waitForScreen("L1")
	time.Sleep(250 * time.Millisecond)

	screen := c.screen()

	for _, line := range []string{"L1", "L2", "L3", "L4>"} {
		if !strings.Contains(screen, line) {
			t.Fatalf("prompt line %q was lost when rendering at the bottom of the window:\n%s",
				line, screen)
		}
	}
}

func TestRenderMultilinePromptWrapAtBottom(t *testing.T) {
	const rows = 10

	c := beginConsole(t, consoleConfig{
		prompt:  "TOP\nP> ",
		cols:    20,
		rows:    rows,
		prefill: rows - 1,
	})

	c.waitForScreen("P>")

	c.send(strings.Repeat("y", 50))
	screen := c.waitUntil(func(s string) bool {
		return strings.Count(s, "y") >= 50
	})

	for _, line := range []string{"TOP", "P> "} {
		if !strings.Contains(screen, line) {
			t.Fatalf("prompt line %q was lost when a multi-line prompt wrapped at the bottom:\n%s",
				line, screen)
		}
	}
}

func TestCursorProbeEnabledByDefault(t *testing.T) {
	c := craftConsole(t, "PROMPT> ", 80, 24)
	c.waitForScreen("PROMPT>")

	deadline := time.Now().Add(3 * time.Second)
	for c.probeQueries() == 0 && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}

	if c.probeQueries() == 0 {
		t.Fatal("expected at least one ESC[6n cursor-position query with probing enabled")
	}
}

func TestRenderWithCursorProbeDisabled(t *testing.T) {

	lying := func(vt10x.Cursor) string { return "\x1b[1;1R" }

	c := beginConsole(t, consoleConfig{
		prompt:     "PROMPT> ",
		cols:       40,
		rows:       24,
		noProbe:    true,
		probeReply: lying,
	})
	c.waitForScreen("PROMPT>")

	c.send(strings.Repeat("x", 40))
	screen := c.waitUntil(func(s string) bool {
		return strings.Count(s, "x") >= 40
	})

	if got := c.probeQueries(); got != 0 {
		t.Fatalf("expected no ESC[6n queries when probing is disabled, got %d", got)
	}

	if got := tallyLine(screen, "PROMPT>"); got != 1 {
		t.Fatalf("prompt should render exactly once with probing disabled, got %d:\n%s", got, screen)
	}

	row0 := strings.TrimRight(strings.SplitN(screen, "\n", 2)[0], " ")
	if want := "PROMPT> " + strings.Repeat("x", 32); row0 != want {
		t.Fatalf("row 0 misaligned with probing disabled:\n  got:  %q\n  want: %q", row0, want)
	}
}

func TestRenderPromptAtBottomWithCursorProbeDisabled(t *testing.T) {
	const rows = 10

	c := beginConsole(t, consoleConfig{
		prompt:  "PROMPT> ",
		cols:    80,
		rows:    rows,
		prefill: rows - 1,
		noProbe: true,
	})

	c.waitForScreen("PROMPT>")

	c.send("hello")
	screen := c.waitForScreen("PROMPT> hello")

	if got := c.probeQueries(); got != 0 {
		t.Fatalf("expected no ESC[6n queries when probing is disabled, got %d", got)
	}

	if got := tallyLine(screen, "PROMPT>"); got != 1 {
		t.Fatalf("prompt should render exactly once, got %d:\n%s", got, screen)
	}

	promptRow := -1

	for i, line := range strings.Split(screen, "\n") {
		if strings.HasPrefix(line, "PROMPT>") {
			promptRow = i

			break
		}
	}

	if promptRow == rows-1 {
		t.Fatalf("input line left on the last row, with no spare row below it:\n%s", screen)
	}
}

func TestRenderRightEdgeWrapAtBottomKeepsInputVisiblePerKey(t *testing.T) {
	const (
		cols = 20
		rows = 8
	)
	const prompt = "P> "

	c := beginConsole(t, consoleConfig{
		prompt:  prompt,
		cols:    cols,
		rows:    rows,
		prefill: rows - 1,
	})

	typed := "abcdefghijklmnopqrstu"
	firstWrappedInputLen := cols - len(prompt) + 1
	stableTopRow := -1

	c.waitForScreen(prompt)
	for idx, ch := range typed {
		c.send(string(ch))
		want := typed[:idx+1]

		screen := c.waitUntil(func(screen string) bool {
			return strings.Contains(compactScreen(screen), "P>"+want)
		})

		if idx+1 == firstWrappedInputLen {
			stableTopRow = rowIndex(screen, prompt)
		}
		if idx+1 > firstWrappedInputLen {
			if got := rowIndex(screen, prompt); got != stableTopRow {
				t.Fatalf("prompt scrolled while typing within the same wrapped row after key %d (%q): prompt row %d, want %d\n%s", idx+1, ch, got, stableTopRow, screen)
			}
		}
	}

	c.send("\r")
	c.waitUntil(func(screen string) bool {
		return strings.Contains(compactScreen(screen), "[LINE:"+typed+"]")
	})
}
