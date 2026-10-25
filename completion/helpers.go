package completion

import (
	"strings"
	"unicode"

	"github.com/malivvan/readline/color"
	"github.com/malivvan/readline/keymap"
	"github.com/malivvan/readline/term"
)

const (
	trailingDescLen  = 3
	trailingValueLen = 4
)

var sanitizer = strings.NewReplacer(
	"\n", ``,
	"\r", ``,
	"\t", ``,
)

func containsUpper(line []rune) bool {
	for _, r := range line {
		if unicode.IsUpper(r) {
			return true
		}
	}

	return false
}

func longest(vals []string, trimEscapes bool) int {
	var length int

	for _, val := range vals {
		if trimEscapes {
			val = color.Strip(val)
		}

		if len(val) > length {
			length = len(val)
		}
	}

	return length
}

func sum(vals []int) (sum int) {
	for _, val := range vals {
		sum += val
	}

	return
}

func (e *Engine) adjustCycleKeys(row, column int) (int, int) {
	cur := e.currentGroup()

	keyRunes := e.keys.Caller()
	keys := string(keyRunes)

	if row > 0 {
		if cur.aliased && keys != term.ArrowRight && keys != term.ArrowDown {
			row, column = 0, row
		} else if keys == term.ArrowDown {
			row, column = 0, row
		}
	} else {
		if cur.aliased && keys != term.ArrowLeft && keys != term.ArrowUp {
			row, column = 0, 1*row
		} else if keys == term.ArrowUp {
			row, column = 0, 1*row
		}
	}

	return row, column
}

func (e *Engine) adjustSelectKeymap() {
	if e.keymap.Local() != keymap.Isearch {
		e.keymap.SetLocal(keymap.MenuSelect)
	}
}

func (e *Engine) completionCount() (comps int, used int) {
	for _, group := range e.groups {

		for _, row := range group.rows {
			comps += len(row)
		}

		if group.tag != "" {
			used++
		}

		if group.maxY > len(group.rows) {
			used += group.maxY
		} else {
			used += len(group.rows)
		}
	}

	return comps, used
}

func (e *Engine) currentGroup() (grp *group) {
	for _, g := range e.groups {
		if g.isCurrent {
			return g
		}
	}

	if len(e.groups) > 0 {
		for _, g := range e.groups {
			if len(g.rows) > 0 {
				g.isCurrent = true
				return g
			}
		}
	}

	return
}

func (e *Engine) cycleNextGroup() {
	for pos, g := range e.groups {
		if g.isCurrent {
			g.isCurrent = false

			if pos == len(e.groups)-1 {
				e.groups[0].isCurrent = true
			} else {
				e.groups[pos+1].isCurrent = true
			}

			break
		}
	}

	for {
		next := e.currentGroup()
		if len(next.rows) == 0 {
			e.cycleNextGroup()
			continue
		}

		return
	}
}

func (e *Engine) cyclePreviousGroup() {
	for pos, g := range e.groups {
		if g.isCurrent {
			g.isCurrent = false

			if pos == 0 {
				e.groups[len(e.groups)-1].isCurrent = true
			} else {
				e.groups[pos-1].isCurrent = true
			}

			break
		}
	}

	for {
		prev := e.currentGroup()
		if len(prev.rows) == 0 {
			e.cyclePreviousGroup()
			continue
		}

		return
	}
}

func (e *Engine) generate(completions Values) {

	defer func() {
		e.hintCompletions(completions)
	}()

	if len(completions.values) == 0 {
		return
	}

	matchCase := e.config.GetBool("completion-ignore-case")
	completions.values = completions.values.FilterPrefix(e.prefix, !matchCase)

	completions.values.EachTag(e.generateGroup(completions))
	e.justifyGroups(completions)
}

func (e *Engine) generateGroup(comps Values) func(tag string, values RawValues) {
	return func(tag string, values RawValues) {

		vals, noDescVals, descriptions := e.groupNonDescribed(&comps, values)

		e.newCompletionGroup(comps, tag, vals, descriptions)

		if len(noDescVals) > 0 {
			e.newCompletionGroup(comps, "", noDescVals, descriptions)
		}
	}
}

func (e *Engine) getAbsPos() int {
	var prev int
	var foundCurrent bool

	for _, grp := range e.groups {
		groupComps := 0

		for _, row := range grp.rows {
			groupComps += len(row)
		}

		if groupComps == 0 {
			continue
		}

		if grp.tag != "" {
			prev++
		}

		if grp.isCurrent {
			prev += grp.posY
			foundCurrent = true

			break
		}

		prev += grp.maxY
	}

	if !foundCurrent {
		return 0
	}

	return prev
}

func (e *Engine) groupNonDescribed(comps *Values, values RawValues) (vals, noDescVals RawValues, descs []string) {
	var descriptions []string

	prefix := ""
	if e.prefix != "\"\"" && e.prefix != "''" {
		prefix = e.prefix
	}

	for _, val := range values {

		if val.Display == "" {
			val.Display = val.Value
		}

		if strings.HasPrefix(val.Value, prefix+"ERR") && val.Value == prefix+"_" {
			comps.Messages.Add(color.FgRed + val.Display + val.Description)

			continue
		}

		if val.Description == "" {
			noDescVals = append(noDescVals, val)

			continue
		}

		descriptions = append(descriptions, val.Description)
		vals = append(vals, val)
	}

	if len(vals) == 0 {
		vals = noDescVals
		noDescVals = make(RawValues, 0)
	}

	return vals, noDescVals, descriptions
}

func (e *Engine) hasUniqueCandidate() bool {
	switch len(e.groups) {
	case 0:
		return false

	case 1:
		cur := e.currentGroup()
		if cur == nil {
			return false
		}

		if len(cur.rows) == 1 {
			return len(cur.rows[0]) == 1
		}

		return len(cur.rows) == 1

	default:
		var count int

	GROUPS:
		for _, group := range e.groups {
			for _, row := range group.rows {
				for range row {
					count++
				}

				if count > 1 {
					break GROUPS
				}
			}
		}

		return count == 1
	}
}

func (e *Engine) justifyGroups(values Values) {
	commandGroups := make([]*group, 0)
	maxCellLength := 0

	for _, group := range e.groups {

		justify := values.Pad[group.tag]
		if !justify {
			justify = values.Pad["*"]
		}

		if !justify {
			continue
		}

		if group.aliased || len(group.columnsWidth) > 1 {
			continue
		}

		commandGroups = append(commandGroups, group)

		if group.longestValue > maxCellLength {
			maxCellLength = group.longestValue
		}
	}

	for _, group := range commandGroups {
		group.columnsWidth[0] = maxCellLength
		group.longestValue = maxCellLength
	}
}

func (e *Engine) needsAutoComplete() bool {

	needsComplete := e.config.GetBool("autocomplete") &&
		e.keymap.Local() != keymap.MenuSelect &&
		e.keymap.Local() != keymap.Isearch &&
		e.line.Len() > 0

	isCorrectMenu := e.keymap.Main() != keymap.ViCommand &&
		e.keymap.Local() != keymap.Isearch

	if needsComplete && isCorrectMenu && len(e.selected.Value) == 0 {
		return true
	}

	if e.keymap.Local() != keymap.MenuSelect && e.autoForce {
		return true
	}

	return false
}

func (e *Engine) noCompletions() bool {
	for _, group := range e.groups {
		if len(group.rows) > 0 {
			return false
		}
	}

	return true
}

func (e *Engine) prepare(completions Values) {
	e.prefix = ""
	e.groups = make([]*group, 0)

	e.setPrefix(completions)
	e.setSuffix(completions)
	e.generate(completions)
}

func (e *Engine) resetValues(comps, cached bool) {
	e.selected = Candidate{}

	if comps {
		e.usedY = 0
		e.groups = make([]*group, 0)
	}

	if cached {
		e.cached = nil
	}

	if len(e.groups) > 0 {
		for _, g := range e.groups {
			g.isCurrent = false
		}

		e.groups[0].isCurrent = true
	}
}

func (e *Engine) setPrefix(completions Values) {
	switch completions.PREFIX {
	case "":

		cpos := e.cursor.Pos() - 1
		if cpos < 0 {
			cpos = 0
		}

		bpos, _ := e.line.SelectBlankWord(cpos)

		if bpos > cpos {
			bpos, cpos = cpos, bpos
		}

		if cpos < e.line.Len() {
			cpos++
		}

		e.prefix = strings.TrimSpace(string((*e.line)[bpos:cpos]))

	default:
		e.prefix = completions.PREFIX
	}
}

func (e *Engine) setSuffix(completions Values) {
	switch completions.SUFFIX {
	case "":
		cpos := e.cursor.Pos()
		_, epos := e.line.SelectBlankWord(cpos)

		if epos < e.line.Len() {
			epos++
		}

		if epos < cpos {
			epos, cpos = cpos, epos
		}

		if epos < e.line.Len() {
			if (*e.line)[epos] == '\'' || (*e.line)[epos] == '"' {
				epos++
			}
		}

		e.suffix = strings.TrimSpace(string((*e.line)[cpos:epos]))

	default:
		e.suffix = completions.SUFFIX
	}
}
