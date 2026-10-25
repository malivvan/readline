package completion

import (
	"strings"

	"github.com/malivvan/readline/color"
	"github.com/malivvan/readline/term"
)

func (e *Engine) hintCompletions(comps Values) {
	hint := ""

	if len(comps.values) == 0 || e.config.GetBool("usage-hint-always") {
		if comps.Usage != "" {
			hint += color.Dim + comps.Usage + color.Reset + term.NewlineReturn
		}
	}

	messages := strings.Join(comps.Messages.Get(), term.NewlineReturn)
	messages = strings.TrimSuffix(messages, term.NewlineReturn)

	if messages != "" {
		hint = hint + color.Dim + messages
	}

	if e.Matches() == 0 && hint == "" && !e.auto {
		hint = e.hintNoMatches()
	}

	hint = strings.TrimSuffix(hint, term.NewlineReturn)
	if hint == "" {
		return
	}

	e.hint.Set(hint + color.Reset)
}

func (e *Engine) hintNoMatches() string {
	noMatches := color.Dim + "no matching"

	var groups []string

	for _, group := range e.groups {
		if group.tag == "" {
			continue
		}

		groups = append(groups, group.tag)
	}

	if len(groups) > 0 {
		groupsStr := strings.Join(groups, ", ")
		noMatches += "'" + groupsStr + "'"
	}

	return noMatches + " completions"
}
