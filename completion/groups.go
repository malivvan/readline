package completion

import (
	"math"
	"slices"
	"strings"

	"github.com/malivvan/readline/color"
	"github.com/malivvan/readline/term"
)

type group struct {
	tag               string
	rows              [][]Candidate
	noSpace           SuffixMatcher
	columnsWidth      []int
	descriptionsWidth []int
	listSeparator     string
	list              bool
	noSort            bool
	aliased           bool
	preserveEscapes   bool
	isCurrent         bool
	longestValue      int
	longestDesc       int
	maxDescAllowed    int
	termWidth         int

	posX int
	posY int
	maxX int
	maxY int
}

func completionsAreAliases(values []Candidate) bool {
	oddValueMap := make(map[string]bool)

	for _, value := range values {
		if value.Description == "" {
			continue
		}

		if _, found := oddValueMap[value.Description]; found {
			return true
		}

		oddValueMap[value.Description] = true
	}

	return false
}

func spawnGrid(values []Candidate, rowCount, maxColumns int) [][]Candidate {
	if rowCount < 0 {
		rowCount = 0
	}

	grid := make([][]Candidate, rowCount)

	for i := range rowCount {
		grid[i] = spawnRow(values, maxColumns, i)
	}

	return grid
}

func spawnRow(domains []Candidate, maxColumns, rowIndex int) []Candidate {
	rowStart := rowIndex * maxColumns
	rowEnd := (rowIndex + 1) * maxColumns
	rowEnd = min(len(domains), rowEnd)

	return domains[rowStart:rowEnd]
}

func padSpace(times int) string {
	if times > 0 {
		return strings.Repeat(" ", times)
	}

	return ""
}

func (e *Engine) newCompletionGroup(comps Values, tag string, vals RawValues, descriptions []string) {
	grp := &group{
		tag:          tag,
		noSpace:      comps.NoSpace,
		posX:         -1,
		posY:         -1,
		columnsWidth: []int{0},
		termWidth:    term.GetWidth(),
		longestDesc:  longest(descriptions, true),
	}

	grp.initOptions(e, &comps, tag, vals)

	if !grp.noSort {
		vals.sortStable()
	}

	grp.prepareValues(vals)

	if completionsAreAliases(vals) {
		grp.initCompletionAliased(vals)
	} else {
		grp.initCompletionsGrid(vals)
	}

	e.groups = append(e.groups, grp)
}

func (g *group) calculateMaxColumnWidths(grid [][]Candidate) {
	var numColumns int

	for _, row := range grid {
		if len(row) > numColumns {
			numColumns = len(row)
		}
	}

	values := make([]int, numColumns)
	descriptions := make([]int, numColumns)

	for _, row := range grid {
		for columnIndex, value := range row {
			if value.displayLen+1 > values[columnIndex] {
				values[columnIndex] = value.displayLen + 1
			}

			if value.descLen+1 > descriptions[columnIndex] {
				descriptions[columnIndex] = value.descLen + 1
			}
		}
	}

	if len(grid) == 1 && len(grid[0]) <= numColumns && sum(descriptions) == 0 {
		for i := range values {
			values[i] = g.longestValue
		}
	}

	shouldPad := len(grid) > 1 && numColumns > 1 && sum(descriptions) == 0
	intraColumnSpace := (numColumns * 2)
	totalSpaceUsed := sum(values) + sum(descriptions) + intraColumnSpace
	freeSpace := g.termWidth - totalSpaceUsed

	if shouldPad && !g.aliased && freeSpace >= numColumns {
		each := freeSpace / numColumns

		for i := range values {
			values[i] += each
		}
	}

	g.maxY = len(g.rows)
	g.maxX = len(values)
	g.columnsWidth = values
	g.descriptionsWidth = descriptions
}

func (g *group) createDescribedRows(values []Candidate) ([][]Candidate, []string) {
	descriptionMap := make(map[string][]Candidate)
	uniqueDescriptions := make([]string, 0)

	for i, description := range values {
		if slices.Contains(uniqueDescriptions, description.Description) {
			descriptionMap[description.Description] = append(descriptionMap[description.Description], values[i])
		} else {
			uniqueDescriptions = append(uniqueDescriptions, description.Description)
			descriptionMap[description.Description] = []Candidate{values[i]}
		}
	}

	rows := make([][]Candidate, 0, len(uniqueDescriptions))

	for _, description := range uniqueDescriptions {
		row := descriptionMap[description]
		rows = append(rows, row)
	}

	return rows, uniqueDescriptions
}

func (g *group) findFirstCandidate(x, y int) (done, next bool) {
	for g.posX > len(g.rows[g.posY])-1 {
		g.posY += y
		g.posY += x

		if g.posY < 0 {
			if g.posX == 0 {
				g.posX = 0
				g.posY = 0

				return true, false
			}

			g.posY = len(g.rows) - 1
			g.posX--
		}

		if g.posY > g.maxY-1 {
			g.posY = 0
			if g.posX < len(g.columnsWidth)-1 {
				g.posX++
			} else {
				return true, true
			}
		}
	}

	return
}

func (g *group) firstCell() {
	g.posX = 0
	g.posY = 0
}

func (g *group) getPad(value Candidate, columnIndex int, desc bool) int {
	columns := g.columnsWidth
	valLen := value.displayLen - 1

	if desc {
		columns = g.descriptionsWidth
		valLen = value.descLen
	}

	column := columns[columnIndex]
	column = min(g.termWidth-1, column)
	padding := column - valLen

	if padding < 0 {
		return 0
	}

	return padding
}

func (g *group) initCompletionAliased(domains []Candidate) {
	g.aliased = true

	grid, _ := g.createDescribedRows(domains)
	g.calculateMaxColumnWidths(grid)
	g.wrapExcessAliases(grid)

	g.maxY = len(g.rows)
	g.maxX = len(g.columnsWidth)
}

func (g *group) initCompletionsGrid(comps RawValues) {
	if len(comps) == 0 {
		return
	}

	pairLength := g.longestValueDescribed(comps)
	pairLength = min(g.termWidth, pairLength)

	maxColumns := g.termWidth / pairLength
	if g.list || maxColumns < 0 {
		maxColumns = 1
	}

	rowCount := int(math.Ceil(float64(len(comps)) / (float64(maxColumns))))

	g.rows = spawnGrid(comps, rowCount, maxColumns)
	g.calculateMaxColumnWidths(g.rows)
}

func (g *group) initOptions(eng *Engine, comps *Values, tag string, vals RawValues) {

	_, g.list = comps.ListLong[tag]
	if _, all := comps.ListLong["*"]; all && len(comps.ListLong) == 1 {
		g.list = true
	}

	g.listSeparator = color.UnquoteRC(eng.config.GetString("completion-list-separator"))
	if g.listSeparator == "" {
		g.listSeparator = "--"
	}

	g.preserveEscapes = comps.Escapes[g.tag]
	if !g.preserveEscapes {
		g.preserveEscapes = comps.Escapes["*"]
	}

	if strings.HasSuffix(g.tag, "commands") && len(vals) > 0 && vals[0].Description != "" {
		g.list = true
	}

	listSep, found := comps.ListSep[tag]
	if !found {
		if allSep, found := comps.ListSep["*"]; found {
			g.listSeparator = allSep
		}
	} else {
		g.listSeparator = listSep
	}

	g.noSort = comps.NoSort[tag]
	if noSort, all := comps.NoSort["*"]; noSort && all && len(comps.NoSort) == 1 {
		g.noSort = true
	}
}

func (g *group) lastCell() {
	g.posY = len(g.rows) - 1
	g.posX = len(g.columnsWidth) - 1

	if g.aliased {
		g.findFirstCandidate(0, -1)
	} else {
		g.posX = len(g.rows[g.posY]) - 1
	}
}

func (g *group) listSep() string {
	return g.listSeparator + " "
}

func (g *group) longestValueDescribed(vals []Candidate) int {
	var longestDesc, longestVal int

	descSeparatorLen := 1 + len(g.listSeparator) + 1

	for _, val := range vals {
		if val.displayLen > longestVal {
			longestVal = val.displayLen
		}

		if val.descLen > longestDesc {
			longestDesc = val.descLen
		}
	}

	if longestDesc > 0 {
		longestDesc += descSeparatorLen
	}

	return longestVal + longestDesc + 2
}

func (g *group) moveSelector(x, y int) (done, next bool) {

	if g.posX == -1 && g.posY == -1 {
		if x != 0 {
			g.posY++
		} else {
			g.posX++
		}
	}

	g.posX += x
	g.posY += y
	reverse := (x < 0 || y < 0)

	if g.posX < 0 {
		if g.posY == 0 && reverse {
			g.posX = 0
			g.posY = 0

			return true, false
		}

		g.posY--
		g.posX = len(g.rows[g.posY]) - 1
	}

	if g.posY < 0 {
		if g.posX == 0 {
			g.posX = 0
			g.posY = 0

			return true, false
		}

		g.posY = len(g.rows) - 1
		g.posX--
	}

	if g.posY > g.maxY-1 {
		g.posY = 0
		if g.posX < g.maxX-1 {
			g.posX++
		} else {
			return true, true
		}
	}

	if g.posX > len(g.rows[g.posY])-1 {
		if g.aliased {
			return g.findFirstCandidate(x, y)
		}

		g.posX = 0

		if g.posY < g.maxY-1 {
			g.posY++
		} else {
			return true, true
		}
	}

	return false, false
}

func (g *group) prepareValues(vals RawValues) RawValues {
	for pos, value := range vals {
		if value.Display == "" {
			value.Display = value.Value
		}

		value.displayLen = len(color.Strip(value.Display))
		value.descLen = len(color.Strip(value.Description))

		if value.displayLen > g.longestValue {
			g.longestValue = value.displayLen
		}

		if value.descLen > g.longestDesc {
			g.longestDesc = value.descLen
		}

		vals[pos] = value
	}

	return vals
}

func (g *group) selected() (comp Candidate) {
	defer func() {
		if !g.preserveEscapes {
			comp.Value = color.Strip(comp.Value)
		}
	}()

	if g.posY == -1 || g.posX == -1 {
		return g.rows[0][0]
	}

	return g.rows[g.posY][g.posX]
}

func (g *group) setMaximumSizes(col int) int {

	maxDescLen := g.descriptionsWidth[col]
	valuesRealLen := sum(g.columnsWidth) + len(g.columnsWidth) + len(g.listSep())

	if valuesRealLen+maxDescLen > g.termWidth {
		maxDescLen = g.termWidth - valuesRealLen
	} else if valuesRealLen+maxDescLen < g.termWidth {
		maxDescLen = g.termWidth - valuesRealLen
	}

	return maxDescLen
}

func (g *group) trimDesc(val Candidate, pad int) (desc, padded string) {
	desc = val.Description
	if desc == "" {
		return desc, padSpace(pad)
	}

	if pad > g.maxDescAllowed {
		pad = g.maxDescAllowed - val.descLen
	}

	desc = sanitizer.Replace(desc)

	if val.descLen > g.maxDescAllowed && g.maxDescAllowed > 0 {
		desc = color.Trim(desc, g.maxDescAllowed-trailingDescLen)
		desc += "..."

		return g.listSep() + desc, ""
	}

	if val.descLen+pad > g.maxDescAllowed {
		pad = g.maxDescAllowed - val.descLen
	}

	return g.listSep() + desc, padSpace(pad)
}

func (g *group) trimDisplay(comp Candidate, pad, col int) (candidate, padded string) {
	val := comp.Display

	if val == "" {
		return "", padSpace(pad)
	}

	maxDisplayWidth := g.columnsWidth[col] + 1
	maxDisplayWidth = min(g.termWidth, maxDisplayWidth)

	val = sanitizer.Replace(val)

	if comp.displayLen > maxDisplayWidth {
		val = color.Trim(val, maxDisplayWidth-trailingValueLen)
		val += "..."

		return val, " "
	}

	return val, padSpace(pad)
}

func (g *group) updateIsearch(eng *Engine) {
	if eng.IsearchRegex == nil {
		return
	}

	suggs := make([]Candidate, 0)

	for i := range g.rows {
		row := g.rows[i]

		for _, val := range row {
			if eng.IsearchRegex.MatchString(val.Value) {
				suggs = append(suggs, val)
			} else if val.Description != "" && eng.IsearchRegex.MatchString(val.Description) {
				suggs = append(suggs, val)
			}
		}
	}

	g.rows = make([][]Candidate, 0)
	g.posX = -1
	g.posY = -1

	suggs = g.prepareValues(suggs)

	if completionsAreAliases(suggs) {
		g.initCompletionAliased(suggs)
	} else {
		g.initCompletionsGrid(suggs)
	}
}

func (g *group) wrapExcessAliases(grid [][]Candidate) {
	breakeven := 0
	maxColumns := len(g.columnsWidth)

	for i, width := range g.columnsWidth {
		if (breakeven + width + 1) > g.termWidth/2 {
			maxColumns = i
			break
		}

		breakeven += width + 1
	}

	if maxColumns < 1 {
		maxColumns = 1
	}

	var rows [][]Candidate

	for rowIndex := range grid {
		row := grid[rowIndex]

		for len(row) > maxColumns {
			rows = append(rows, row[:maxColumns])
			row = row[maxColumns:]
		}

		rows = append(rows, row)
	}

	g.rows = rows
	g.columnsWidth = g.columnsWidth[:maxColumns]
}
