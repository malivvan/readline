//go:build unix

package display_test

import (
	"strings"
	"testing"
	"time"
)

func TestAsyncRefreshCompletions(t *testing.T) {
	c := beginConsole(t, consoleConfig{
		prompt:    "P> ",
		cols:      80,
		rows:      24,
		asyncComp: true,
		asyncMS:   400,
	})
	c.waitForScreen("P>")

	c.send("\x1b?")
	screen := c.waitForScreen("alpha")

	if strings.Contains(screen, "alpaca") {
		t.Fatalf("alpaca should not be present before the async refresh:\n%s", screen)
	}

	c.waitForScreen("alpaca")
}

func TestAsyncRefreshCompletionsNoMenuIsNoop(t *testing.T) {
	c := beginConsole(t, consoleConfig{
		prompt:    "P> ",
		cols:      80,
		rows:      24,
		asyncComp: true,
		asyncMS:   300,
	})
	c.waitForScreen("P>")

	time.Sleep(900 * time.Millisecond)

	screen := c.screen()
	if strings.Contains(screen, "alpha") || strings.Contains(screen, "alpaca") {
		t.Fatalf("RefreshCompletions must no-op when no menu is active, but a menu appeared:\n%s", screen)
	}
}

func TestAsyncRefreshKeepsSelection(t *testing.T) {
	c := beginConsole(t, consoleConfig{
		prompt:    "P> ",
		cols:      80,
		rows:      24,
		asyncComp: true,
		asyncMS:   1500,
	})
	c.waitForScreen("P>")

	c.send("\x1b?")
	c.waitForScreen("alpha")

	c.send("\x1b[B")
	c.send("\x1b[B")
	screen := c.waitForScreen("P> alpine")

	if strings.Contains(screen, "alpaca") {
		t.Fatalf("alpaca should not be present before the async refresh:\n%s", screen)
	}

	c.waitForScreen("alpaca")

	final := c.screen()
	first := strings.TrimRight(strings.SplitN(final, "\n", 2)[0], " ")

	if first != "P> alpine" {
		t.Fatalf("selection not preserved across async refresh: first line = %q want %q\n%s",
			first, "P> alpine", final)
	}
}
