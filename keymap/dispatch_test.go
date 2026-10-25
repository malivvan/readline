package keymap

import (
	"testing"

	"github.com/malivvan/readline/core"
	"github.com/malivvan/readline/inputrc"
	"github.com/malivvan/readline/strutil"
)

func builtinBindMaps(t *testing.T) map[string]map[string]inputrc.Bind {
	t.Helper()

	_, cfg := NewEngine(&core.Keys{}, &core.Iterations{})

	return cfg.Binds
}

func collisions(binds map[string]inputrc.Bind) map[string][]inputrc.Bind {
	bySeq := make(map[string][]inputrc.Bind)

	for sequence, bind := range binds {
		seq := strutil.ConvertMeta([]rune(sequence))
		bySeq[seq] = append(bySeq[seq], bind)
	}

	out := make(map[string][]inputrc.Bind)

	for seq, list := range bySeq {
		differ := false

		for _, b := range list[1:] {
			if b != list[0] {
				differ = true

				break
			}
		}

		if differ {
			out[seq] = list
		}
	}

	return out
}

func TestMatchBindSortIsLoadBearing(t *testing.T) {
	binds := builtinBindMaps(t)[string(Emacs)]

	cols := collisions(binds)
	if len(cols) == 0 {
		t.Fatal("expected ConvertMeta collisions in the emacs keymap; if truly none, matchBind no longer needs to order its matches")
	}

	t.Logf("emacs keymap has %d colliding key sequences whose exact match is disambiguated only by matchBind's ordering", len(cols))
}

func TestMatchBindResolvesDeterministically(t *testing.T) {
	eng := &Engine{}
	binds := builtinBindMaps(t)[string(Emacs)]

	cols := collisions(binds)
	if len(cols) == 0 {
		t.Skip("no collisions to probe")
	}

	var probe string

	for seq := range cols {
		if probe == "" || seq < probe {
			probe = seq
		}
	}

	want, _ := eng.matchBind([]byte(probe), binds)

	for range 64 {
		got, _ := eng.matchBind([]byte(probe), binds)
		if got != want {
			t.Fatalf("matchBind(%q) is nondeterministic: got %+v, first saw %+v", probe, got, want)
		}
	}
}
