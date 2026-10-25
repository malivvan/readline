//go:build unix

package display_test

import (
	"strings"
	"testing"
)

func compactScreen(screen string) string {
	return strings.Join(strings.Fields(screen), "")
}

func rowIndex(screen, substr string) int {
	for i, row := range strings.Split(screen, "\n") {
		if strings.Contains(row, substr) {
			return i
		}
	}

	return -1
}

func TestHintProviderTracksLine(t *testing.T) {
	c := beginConsole(t, consoleConfig{
		prompt:       "PROMPT> ",
		cols:         80,
		rows:         24,
		hintProvider: true,
	})
	c.waitForScreen("PROMPT>")

	c.send("abc")
	c.waitForScreen("HINT:abc")

	c.send("d")
	c.waitForScreen("HINT:abcd")
}

func TestHintProviderAboveTransient(t *testing.T) {
	c := beginConsole(t, consoleConfig{
		prompt:       "PROMPT> ",
		cols:         80,
		rows:         24,
		hintProvider: true,
		transient:    "ASYNCMSG",
	})
	c.waitForScreen("PROMPT>")

	c.send("abc")
	screen := c.waitUntil(func(s string) bool {
		return strings.Contains(s, "HINT:abc") && strings.Contains(s, "ASYNCMSG")
	})

	provided := rowIndex(screen, "HINT:abc")
	transient := rowIndex(screen, "ASYNCMSG")

	if provided < 0 || transient < 0 {
		t.Fatalf("both hints must be present (provided=%d transient=%d):\n%s",
			provided, transient, screen)
	}

	if provided >= transient {
		t.Fatalf("provider hint (row %d) must render above transient hint (row %d):\n%s",
			provided, transient, screen)
	}
}

func TestTransientSurvivesIsearchAndRendersAbove(t *testing.T) {
	c := beginConsole(t, consoleConfig{
		prompt:    "PROMPT> ",
		cols:      80,
		rows:      24,
		transient: "ASYNCMSG",
	})
	c.waitForScreen("ASYNCMSG")

	c.send("\x12")
	screen := c.waitForScreen("inc-search")

	transient := rowIndex(screen, "ASYNCMSG")
	isearch := rowIndex(screen, "inc-search")

	if transient < 0 {
		t.Fatalf("transient hint was lost when entering incremental search:\n%s", screen)
	}

	if transient >= isearch {
		t.Fatalf("transient hint (row %d) must render above the isearch hint (row %d):\n%s",
			transient, isearch, screen)
	}
}

func TestAsyncTransientWhileIdle(t *testing.T) {
	c := beginConsole(t, consoleConfig{
		prompt:  "PROMPT> ",
		cols:    80,
		rows:    24,
		asyncMS: 200,
	})
	c.waitForScreen("PROMPT>")

	c.waitForScreen("ASYNCPING")
}

func TestAsyncRefreshKeepsLayout(t *testing.T) {
	c := beginConsole(t, consoleConfig{
		prompt:  "PROMPT> ",
		cols:    80,
		rows:    24,
		asyncMS: 200,
	})
	c.waitForScreen("ASYNCPING")

	c.send("abc")
	screen := c.waitForScreen("PROMPT> abc")

	if got := tallyLine(screen, "PROMPT>"); got != 1 {
		t.Fatalf("prompt should render exactly once after async repaint, got %d:\n%s", got, screen)
	}

	first := strings.TrimRight(strings.SplitN(screen, "\n", 2)[0], " ")
	if first != "PROMPT> abc" {
		t.Fatalf("input misaligned after async repaint:\n  got:  %q\n  want: %q", first, "PROMPT> abc")
	}
}
