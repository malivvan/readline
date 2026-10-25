package completion

import (
	"bufio"
	"fmt"
	"strings"

	"github.com/malivvan/readline/color"
	"github.com/malivvan/readline/term"
)

func Coordinates(e *Engine) int {
	return e.usedY
}

func Display(eng *Engine, maxRows int) {
	eng.usedY = 0

	if eng.Matches() == 0 || eng.skipDisplay {
		term.WriteString(term.ClearLineAfter)
		return
	}

	completions := term.ClearLineAfter

	var completionsSb31 strings.Builder
	for _, group := range eng.groups {
		completionsSb31.WriteString(eng.renderCompletions(group))
	}

	completions += completionsSb31.String()

	completions, eng.usedY = eng.cropCompletions(completions, maxRows)

	if completions != "" {
		term.WriteString(completions)
	}
}

func (e *Engine) cropCompletions(comps string, maxRows int) (cropped string, usedY int) {

	absPos := e.getAbsPos()

	scanner := bufio.NewScanner(strings.NewReader(comps))

	if absPos < maxRows-1 {
		return e.cutCompletionsBelow(scanner, maxRows)
	}

	if absPos >= maxRows-1 {
		return e.cutCompletionsAboveBelow(scanner, maxRows, absPos)
	}

	return
}

func (e *Engine) cutCompletionsAboveBelow(scanner *bufio.Scanner, maxRows, absPos int) (string, int) {
	cutAbove := absPos - maxRows + 1

	var cropped string
	var count int

	var croppedSb231 strings.Builder

	for scanner.Scan() {
		line := scanner.Text()

		if count <= cutAbove {
			count++

			continue
		}

		if count > cutAbove && count <= absPos {
			croppedSb231.WriteString(line + term.NewlineReturn)

			count++
		} else {
			break
		}
	}

	cropped += croppedSb231.String()

	cropped = strings.TrimSuffix(cropped, term.NewlineReturn)
	count -= cutAbove + 1

	_, used := e.completionCount()
	remain := used - (maxRows + cutAbove)

	if remain <= 0 {
		return cropped, count - 1
	}

	cropped += fmt.Sprintf(term.NewlineReturn+color.Dim+color.FgYellow+" %d more completion rows... (scroll down to show)"+color.Reset, remain)

	return cropped, count
}

func (e *Engine) cutCompletionsBelow(scanner *bufio.Scanner, maxRows int) (string, int) {
	var count int
	var cropped string

	var croppedSb200 strings.Builder

	for scanner.Scan() {
		line := scanner.Text()
		if count < maxRows-1 {
			croppedSb200.WriteString(line + term.NewlineReturn)

			count++
		} else {
			break
		}
	}

	cropped += croppedSb200.String()

	cropped = strings.TrimSuffix(cropped, term.NewlineReturn)

	_, used := e.completionCount()
	remain := used - count

	if remain <= 0 {
		return cropped, count - 1
	}

	cropped += fmt.Sprintf(term.NewlineReturn+color.Dim+color.FgYellow+" %d more completion rows... (scroll down to show)"+color.Reset, remain)

	return cropped, count
}

func (e *Engine) highlightDesc(grp *group, val Candidate, pad, row, col int, selected bool) (desc string) {
	if val.Description == "" {
		return color.Reset
	}

	desc, padded := grp.trimDesc(val, pad)

	if len(grp.rows) > row+1 && grp.rows[row+1][0].Description == val.Description {
		desc = "|"
	} else if e.IsearchRegex != nil && e.isearchBuf.Len() > 0 && !selected {
		match := e.IsearchRegex.FindString(desc)
		match = color.Fmt(color.Bg+"244") + match + color.Reset + color.Dim
		desc = e.IsearchRegex.ReplaceAllLiteralString(desc, match)
	}

	if row == grp.posY && col == grp.posX && grp.isCurrent && !grp.aliased {
		userDescStyle := color.UnquoteRC(e.config.GetString("completion-selection-style"))
		selectionHighlightStyle := color.Fmt(color.Bg+"255") + userDescStyle
		desc = strings.ReplaceAll(desc, color.BgDefault, userDescStyle)
		desc = selectionHighlightStyle + desc
	}

	compDescStyle := color.UnquoteRC(e.config.GetString("completion-description-style"))

	return compDescStyle + desc + color.Reset + padded
}

func (e *Engine) highlightDisplay(grp *group, val Candidate, pad, col int, selected bool) (candidate string) {

	if val.Display == "" {
		return padSpace(pad)
	}

	style := color.Fmt(val.Style)
	candidate, padded := grp.trimDisplay(val, pad, col)

	if e.IsearchRegex != nil && e.isearchBuf.Len() > 0 && !selected {
		match := e.IsearchRegex.FindString(candidate)
		match = color.Fmt(color.Bg+"244") + match + color.Reset + style
		candidate = e.IsearchRegex.ReplaceAllLiteralString(candidate, match)
	}

	if selected {

		userStyle := color.UnquoteRC(e.config.GetString("completion-selection-style"))
		selectionHighlightStyle := color.Fmt(color.Bg+"255") + userStyle
		candidate = selectionHighlightStyle + candidate

		if grp.aliased {
			candidate += color.Reset
		}
	} else {

		if e.config.GetBool("colored-completion-prefix") && e.prefix != "" && strings.HasPrefix(candidate, e.prefix) {
			prefixColored := color.Bold + color.FgBlue + e.prefix + color.BoldReset + color.FgDefault + style
			candidate = prefixColored + candidate[len(e.prefix):]
		}

		candidate = style + candidate + color.Reset
	}

	return candidate + padded
}

func (e *Engine) renderCompletions(grp *group) string {
	var builder strings.Builder

	if len(grp.rows) == 0 {
		return ""
	}

	if grp.tag != "" {
		tag := fmt.Sprintf("%s%s%s %s", color.Bold, color.FgYellow, grp.tag, color.Reset)
		builder.WriteString(tag + term.ClearLineAfter + term.NewlineReturn)
	}

	for rowIndex, row := range grp.rows {
		for columnIndex := range grp.columnsWidth {
			var value Candidate

			if len(row) > columnIndex {
				value = row[columnIndex]
			}

			padding := grp.getPad(value, columnIndex, false)
			isSelected := rowIndex == grp.posY && columnIndex == grp.posX && grp.isCurrent
			display := e.highlightDisplay(grp, value, padding, columnIndex, isSelected)

			builder.WriteString(display)

			onLast := columnIndex == len(grp.columnsWidth)-1
			if grp.aliased && onLast && value.Description == "" {
				value = row[0]
			}

			if !grp.aliased || onLast {
				grp.maxDescAllowed = grp.setMaximumSizes(columnIndex)

				descPad := grp.getPad(value, columnIndex, true)
				desc := e.highlightDesc(grp, value, descPad, rowIndex, columnIndex, isSelected)
				builder.WriteString(desc)
			}
		}

		builder.WriteString(term.ClearLineAfter + term.NewlineReturn)
	}

	return builder.String()
}
