//go:build unix

package display_test

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestNoPromptDriftWithHint(t *testing.T) {
	c := beginConsole(t, consoleConfig{
		prompt:    "ANCHOR\nP> ",
		cols:      80,
		rows:      24,
		transient: "STATUS-HINT",
	})
	c.waitForScreen("ANCHOR")
	c.waitForScreen("STATUS-HINT")

	chars := []string{"a", "b", "c", "d"}
	rows := make([]int, 0, len(chars))

	for _, ch := range chars {
		c.send(ch)
		time.Sleep(150 * time.Millisecond)

		screen := c.screen()
		rows = append(rows, rowIndex(screen, "ANCHOR"))
	}

	for i, r := range rows {
		if r != rows[0] {
			t.Fatalf("prompt drifted: ANCHOR row per keystroke = %v (moved at step %d)\nlast screen:\n%s",
				rows, i, c.screen())
		}
	}
}

func TestNoPromptDriftWithHintAndAutocomplete(t *testing.T) {
	c := beginConsole(t, consoleConfig{
		prompt:       "ANCHOR\nP> ",
		cols:         80,
		rows:         24,
		autocomplete: true,
		hintProvider: true,
	})
	c.waitForScreen("ANCHOR")

	chars := strings.Split("alp", "")
	rows := make([]int, 0, len(chars))

	for _, ch := range chars {
		c.send(ch)
		time.Sleep(200 * time.Millisecond)

		screen := c.screen()
		rows = append(rows, rowIndex(screen, "ANCHOR"))
	}

	for i, r := range rows {
		if r != rows[0] {
			t.Fatalf("prompt drifted with autocomplete: ANCHOR row per keystroke = %v (moved at step %d)\nlast screen:\n%s",
				rows, i, c.screen())
		}
	}
}

func TestNoPromptDriftAcrossAsyncWakes(t *testing.T) {
	c := beginConsole(t, consoleConfig{
		prompt:      "ANCHOR\nP> ",
		cols:        80,
		rows:        24,
		asyncMS:     150,
		asyncRepeat: 5,
	})
	c.waitForScreen("ANCHOR")

	rows := make([]int, 0, 5)

	for i := range 5 {
		c.waitForScreen(fmt.Sprintf("ASYNCPING-%d", i))
		rows = append(rows, rowIndex(c.screen(), "ANCHOR"))
	}

	for i, r := range rows {
		if r != rows[0] {
			t.Fatalf("prompt drifted across async wakes: ANCHOR row per update = %v (moved at update %d)\nlast screen:\n%s",
				rows, i, c.screen())
		}
	}
}

func TestNoPromptDriftAsyncWakesWithRightPrompt(t *testing.T) {
	c := beginConsole(t, consoleConfig{
		prompt:      "ANCHOR\nP> ",
		cols:        80,
		rows:        24,
		rightPrompt: "12:00:00.000",
		asyncMS:     150,
		asyncRepeat: 5,
	})
	c.waitForScreen("ANCHOR")

	rows := make([]int, 0, 5)

	for i := range 5 {
		c.waitForScreen(fmt.Sprintf("ASYNCPING-%d", i))
		rows = append(rows, rowIndex(c.screen(), "ANCHOR"))
	}

	for i, r := range rows {
		if r != rows[0] {
			t.Fatalf("prompt drifted across async wakes (with right prompt): ANCHOR row = %v (moved at %d)\n%s",
				rows, i, c.screen())
		}
	}
}

func TestNoPromptDriftTwoHintLanes(t *testing.T) {
	c := beginConsole(t, consoleConfig{
		prompt:       "ANCHOR\nP> ",
		cols:         80,
		rows:         24,
		hintProvider: true,
		asyncMS:      150,
		asyncRepeat:  5,
	})
	c.waitForScreen("ANCHOR")

	c.send("x")
	c.waitForScreen("HINT:x")

	rows := make([]int, 0, 5)

	for i := range 5 {
		c.waitForScreen(fmt.Sprintf("ASYNCPING-%d", i))
		rows = append(rows, rowIndex(c.screen(), "ANCHOR"))
	}

	for i, r := range rows {
		if r != rows[0] {
			t.Fatalf("prompt drifted with two hint lanes: ANCHOR row = %v (moved at %d)\n%s",
				rows, i, c.screen())
		}
	}
}

func TestNoPromptDriftWithWrappedUpperPrompt(t *testing.T) {
	long := "example [main] in " + strings.Repeat("/a/very/long/working/directory", 5)

	c := beginConsole(t, consoleConfig{
		prompt: long + "\nP> ",
		cols:   80,
		rows:   24,
	})
	c.waitForScreen("example [main]")

	for _, ch := range []string{"a", "b", "c", "d"} {
		c.send(ch)
		time.Sleep(150 * time.Millisecond)
	}

	c.send("\r")
	c.waitForScreen("[LINE:abcd]")

	screen := c.screen()
	if got := strings.Count(screen, "example [main] in"); got != 1 {
		t.Fatalf("wrapped prompt upper line rendered %d times, want once:\n%s", got, screen)
	}
}
