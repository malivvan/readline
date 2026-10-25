package history

import (
	"testing"

	"github.com/malivvan/readline/core"
)

func craftTestSources(maxEntries int) (*Sources, *memory, *core.Line) {
	line := new(core.Line)
	mem := new(memory)

	src := &Sources{
		line:       line,
		list:       map[string]Source{standardSourceName: mem},
		names:      []string{standardSourceName},
		maxEntries: maxEntries,
	}

	return src, mem, line
}

func persistLine(src *Sources, line *core.Line, text string) {
	*line = core.Line([]rune(text))

	src.Write(false)
}

func TestWriteUnlimitedPersists(t *testing.T) {
	src, mem, line := craftTestSources(-1)

	persistLine(src, line, "one")
	persistLine(src, line, "two")
	persistLine(src, line, "three")

	if mem.Len() != 3 {
		t.Fatalf("unlimited history wrote %d lines, want 3", mem.Len())
	}
}

func TestWriteUnderCapPersists(t *testing.T) {
	src, mem, line := craftTestSources(5)

	persistLine(src, line, "one")
	persistLine(src, line, "two")

	if mem.Len() != 2 {
		t.Fatalf("capped history under the limit wrote %d lines, want 2 (inverted-condition regression)", mem.Len())
	}
}

func TestWriteStopsAtCap(t *testing.T) {
	src, mem, line := craftTestSources(2)

	persistLine(src, line, "one")
	persistLine(src, line, "two")
	persistLine(src, line, "three")

	if mem.Len() != 2 {
		t.Fatalf("history grew to %d lines past its cap of 2", mem.Len())
	}
}

func TestWriteZeroMaxIsUnlimited(t *testing.T) {
	src, mem, line := craftTestSources(0)

	persistLine(src, line, "one")
	persistLine(src, line, "two")

	if mem.Len() != 2 {
		t.Fatalf("maxEntries==0 wrote %d lines, want 2 (should be unlimited)", mem.Len())
	}
}

func TestMemoryGetLineOutOfRange(t *testing.T) {
	mem := new(memory)
	_, _ = mem.Write("test")

	if _, err := mem.GetLine(mem.Len()); err == nil {
		t.Fatalf("GetLine(%d) on a %d-line history returned nil error, want out-of-range error", mem.Len(), mem.Len())
	}

	if _, err := mem.GetLine(-1); err == nil {
		t.Fatalf("GetLine(-1) returned nil error, want negative-index error")
	}
}
