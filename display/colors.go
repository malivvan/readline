package display

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/malivvan/readline/color"
	"github.com/malivvan/readline/core"
)

var ansiEscapeRegex = regexp.MustCompile(`\x1b\[[0-9;]+m`)

// ansiEscapes returns the length in runes of the ANSI escape sequence starting
// at each rune index of line. The regexp works on the UTF-8 encoding of the
// line, so its byte offsets cannot be compared to rune positions directly.
func ansiEscapes(line []rune) map[int]int {
	escapes := make(map[int]int)

	for index := 0; index < len(line); index++ {
		if line[index] != 0x1b {
			continue
		}

		rest := string(line[index:])

		loc := ansiEscapeRegex.FindStringIndex(rest)
		if loc == nil || loc[0] != 0 {
			continue
		}

		length := utf8.RuneCountInString(rest[:loc[1]])
		escapes[index] = length

		index += length - 1
	}

	return escapes
}

func orderHighlights(vhl core.Selection) []core.Selection {
	all := make([]core.Selection, 0)
	sorted := make([]core.Selection, 0)
	bpos := make([]int, 0)

	for _, reg := range vhl.Surrounds() {
		all = append(all, reg)
		rbpos, _ := reg.Pos()
		bpos = append(bpos, rbpos)
	}

	if vhl.Active() && vhl.IsVisual() {
		all = append(all, vhl)
		vbpos, _ := vhl.Pos()
		bpos = append(bpos, vbpos)
	}

	sort.Ints(bpos)

	prevIsMatcher := false
	prevPos := 0

	for _, pos := range bpos {
		for _, reg := range all {
			bpos, _ := reg.Pos()
			isMatcher := reg.Type == "matcher"

			if bpos != pos || !reg.Active() || !reg.IsVisual() {
				continue
			}

			if prevIsMatcher && isMatcher && prevPos == pos {
				continue
			}

			sorted = append(sorted, reg)
			prevIsMatcher = reg.Type == "matcher"
			prevPos = bpos

			break
		}
	}

	return sorted
}

func (e *Engine) getHighlights(line []rune, sorted []core.Selection) map[int][]rune {
	highlights := make(map[int][]rune)

	colors := ansiEscapes(line)

	regions := make([]core.Selection, 0)
	pos := -1
	skip := 0

	for rawIndex := range line {
		var posHl []rune
		var newHl core.Selection

		if skip > 0 {
			skip--
			continue
		}

		if length, found := colors[rawIndex]; found {
			skip += length - 1

			continue
		}

		pos++

		for _, hl := range sorted {
			bpos, _ := hl.Pos()

			if bpos == pos {
				newHl = hl
				regions = append(regions, hl)
			}
		}

		regions, posHl = e.hlReset(regions, posHl, pos)
		posHl = e.hlAdd(regions, newHl, posHl)

		if len(posHl) > 0 {
			highlights[rawIndex] = posHl
		}
	}

	return highlights
}

func (e *Engine) highlightLine(line []rune, selection core.Selection) string {

	sorted := orderHighlights(selection)
	colors := e.getHighlights(line, sorted)

	var highlighted string

	var highlightedSb25 strings.Builder

	for i, r := range line {
		if highlight, found := colors[i]; found {
			highlightedSb25.WriteString(string(highlight))
		}

		highlightedSb25.WriteRune(r)
	}

	highlighted += highlightedSb25.String()

	comment := strings.Trim(e.opts.GetString("comment-begin"), "\"")
	if comment != e.commentToken || e.commentRegex == nil {
		e.commentToken = comment
		e.commentRegex, _ = regexp.Compile(fmt.Sprintf(`(^|\s)%s.*`, comment))
	}

	if e.commentRegex != nil {
		commentColor := color.SGRStart + color.Fg + "244" + color.SGREnd
		highlighted = e.commentRegex.ReplaceAllString(highlighted, fmt.Sprintf("%s${0}%s", commentColor, color.Reset))
	}

	highlighted += color.Reset

	return highlighted
}

func (e *Engine) hlAdd(regions []core.Selection, newHl core.Selection, line []rune) []rune {
	var (
		fg, bg  string
		matcher bool
		hl      core.Selection
	)

	if newHl.Active() {
		hl = newHl
	} else if len(regions) > 0 {
		hl = regions[len(regions)-1]
	}

	fg, bg = hl.Highlights()
	matcher = hl.Type == "matcher"

	if bg != "" && !matcher {
		background := color.UnquoteRC(e.opts.GetString("active-region-start-color"))
		if background != "" {
			bg = background
		} else {
			bg = color.Reverse
		}
	}

	line = append(line, []rune(bg)...)
	line = append(line, []rune(fg)...)

	return line
}

func (e *Engine) hlReset(regions []core.Selection, line []rune, pos int) ([]core.Selection, []rune) {
	remaining := make([]core.Selection, 0, len(regions))

	for _, reg := range regions {
		_, epos := reg.Pos()
		foreground, background := reg.Highlights()

		if epos != pos {
			remaining = append(remaining, reg)

			continue
		}

		if foreground != "" {
			line = append(line, []rune(color.FgDefault)...)
		}

		if background != "" {

			line = append(line, []rune(color.ReverseReset)...)
			line = append(line, []rune(color.BgDefault)...)
		}
	}

	return remaining, line
}
