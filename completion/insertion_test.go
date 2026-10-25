package completion

import (
	"testing"

	"github.com/malivvan/readline/core"
)

func craftPrefixEngine(typed string, values ...string) (*Engine, *core.Line) {
	l := core.Line([]rune(typed))
	line := &l
	cur := core.NewCursor(line)
	cur.Set(line.Len())

	var groups []*group

	for i, v := range values {

		gi := i % 2
		for len(groups) <= gi {
			groups = append(groups, &group{})
		}

		groups[gi].rows = append(groups[gi].rows, []Candidate{{Value: v}})
	}

	e := &Engine{
		line:   line,
		cursor: cur,
		prefix: typed,
		groups: groups,
	}

	return e, line
}

func TestCommonStringPrefix(t *testing.T) {
	cases := []struct{ a, b, want string }{
		{"foobar", "foobaz", "fooba"},
		{"foo", "foobar", "foo"},
		{"abc", "xyz", ""},
		{"", "foo", ""},
		{"café", "cafard", "caf"},
	}

	for _, c := range cases {
		if got := commonStringPrefix(c.a, c.b); got != c.want {
			t.Errorf("commonStringPrefix(%q, %q) = %q, want %q", c.a, c.b, got, c.want)
		}
	}
}

func TestInsertCommonPrefix(t *testing.T) {
	t.Run("extends to shared prefix", func(t *testing.T) {
		e, line := craftPrefixEngine("foo", "foobar", "foobaz", "foobat")
		e.InsertCommonPrefix()

		if got := string(*line); got != "fooba" {
			t.Fatalf("line = %q, want %q", got, "fooba")
		}

		if e.prefix != "fooba" {
			t.Fatalf("engine prefix = %q, want %q", e.prefix, "fooba")
		}

		if e.cursor.Pos() != 5 {
			t.Fatalf("cursor = %d, want 5 (end of extended word)", e.cursor.Pos())
		}
	})

	t.Run("no shared prefix beyond typed is a no-op", func(t *testing.T) {
		e, line := craftPrefixEngine("f", "foo", "fbar")
		e.InsertCommonPrefix()

		if got := string(*line); got != "f" {
			t.Fatalf("line = %q, want unchanged %q", got, "f")
		}

		if e.prefix != "f" {
			t.Fatalf("engine prefix = %q, want unchanged %q", e.prefix, "f")
		}
	})

	t.Run("single candidate extends fully", func(t *testing.T) {
		e, line := craftPrefixEngine("re", "readline")
		e.InsertCommonPrefix()

		if got := string(*line); got != "readline" {
			t.Fatalf("line = %q, want %q", got, "readline")
		}
	})

	t.Run("no candidates is a no-op", func(t *testing.T) {
		e, line := craftPrefixEngine("foo")
		e.InsertCommonPrefix()

		if got := string(*line); got != "foo" {
			t.Fatalf("line = %q, want unchanged %q", got, "foo")
		}
	})
}
