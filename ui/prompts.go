package ui

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/malivvan/readline/color"
	"github.com/malivvan/readline/core"
	"github.com/malivvan/readline/inputrc"
	"github.com/malivvan/readline/keymap"
	"github.com/malivvan/readline/strutil"
	"github.com/malivvan/readline/term"
)

type Prompt struct {
	primaryF    func() string
	primaryCols int

	secondaryF func() string
	transientF func() string
	rightF     func() string
	tooltipF   func() string

	refreshing bool

	line    *core.Line
	cursor  *core.Cursor
	keymaps *keymap.Engine
	opts    *inputrc.Config
}

var (
	DefaultSecondaryPrompt = color.FgBlackBright + "\U00002514 " + color.Reset

	DefaultMultilineColumn = color.FgBlackBright + "\U00002502 " + color.Reset
)

func NewPrompt(line *core.Line, cursor *core.Cursor, keymaps *keymap.Engine, opts *inputrc.Config) *Prompt {
	return &Prompt{
		line:    line,
		cursor:  cursor,
		keymaps: keymaps,
		opts:    opts,
	}
}

func (p *Prompt) LastPrint() {
	if p.primaryF == nil {
		return
	}

	lines := strings.Split(p.primaryF(), "\n")

	if len(lines) == 0 {
		return
	}

	prompt := p.formatLastPrompt(lines[len(lines)-1])

	term.WriteString(prompt)

	p.primaryCols = strutil.RealLength(prompt)
}

func (p *Prompt) LastUsed() int {
	if p.primaryF == nil {
		return 0
	}

	lines := strings.Split(p.primaryF(), "\n")
	if len(lines) == 0 {
		return 0
	}

	prompt := p.formatLastPrompt(lines[len(lines)-1])
	p.primaryCols = strutil.RealLength(prompt)

	return p.primaryCols
}

func (p *Prompt) MultilineColumnPrint() {
	numbered := p.opts.GetBool("multiline-column-numbered")
	custom := p.opts.GetString("multiline-column-custom")
	defaultCol := p.opts.GetBool("multiline-column")

	switch {
	case numbered:
		var column strings.Builder
		for pos := range p.line.Lines() {
			fmt.Fprintf(&column, "\n"+color.FgBlackBright+"%d"+color.Reset+" ", pos+2)
		}

		term.WriteString(column.String())
	case len(custom) > 0:
		var column strings.Builder
		for range p.line.Lines() {
			fmt.Fprintf(&column, "\n%s\x1b[0m", custom)
		}

		term.WriteString(column.String())
	case defaultCol:
		var column strings.Builder
		for range p.line.Lines() {
			column.WriteString("\n" + DefaultMultilineColumn)
		}

		term.WriteString(column.String())
	}
}

func (p *Prompt) Primary(prompt func() string) {
	p.primaryF = prompt
}

func (p *Prompt) PrimaryPrint() {
	p.refreshing = false

	if p.primaryF == nil {
		return
	}

	prompt := p.primaryF()

	prompt, lastPrompt := p.formatPrimaryLines(prompt)

	lastPrompt = p.formatLastPrompt(lastPrompt)

	if prompt != "" {
		term.WriteString(prompt)
	}

	term.WriteString(lastPrompt)

	p.primaryCols = strutil.RealLength(lastPrompt)
}

// promptRows returns the number of terminal rows the given prompt fragment
// occupies, including rows produced by line wrapping. The fragment is the part
// of the prompt printed above the last line (it may be empty); a trailing
// newline does not start an extra row because it only moves to the row of the
// last prompt line.
func promptRows(prompt string) int {
	if prompt == "" {
		return 0
	}

	width := term.GetWidth()
	if width <= 0 {
		width = 80
	}

	prompt = strings.TrimSuffix(prompt, "\n")
	if prompt == "" {
		return 0
	}

	rows := 0

	for _, line := range strings.Split(prompt, "\n") {
		lineRows := (strutil.RealLength(line) + width - 1) / width
		if lineRows < 1 {
			lineRows = 1
		}

		rows += lineRows
	}

	return rows
}

// PrimaryUsed returns the number of terminal rows occupied by the prompt lines
// above the last one. It is computed from the current prompt and terminal width
// so a resize is reflected without printing the prompt again.
func (p *Prompt) PrimaryUsed() int {
	if p.primaryF == nil {
		return 0
	}

	upper, _ := p.formatPrimaryLines(p.primaryF())

	return promptRows(upper)
}

func (p *Prompt) Refreshing() bool {
	return p.refreshing
}

func (p *Prompt) Right(prompt func() string) {
	p.rightF = prompt
}

func (p *Prompt) RightPrint(startColumn int, force bool) {
	var rprompt string

	if p.tooltipF != nil && force {
		rprompt = p.tooltipF()
	}

	if rprompt == "" && p.rightF != nil {
		rprompt = p.rightF()
	}

	if rprompt == "" {
		return
	}

	if prompt, canPrint := p.formatRightPrompt(rprompt, startColumn); canPrint {
		term.WriteString(prompt)
	} else {
		term.WriteString(term.ClearLineAfter)
	}
}

func (p *Prompt) Secondary(prompt func() string) {
	p.secondaryF = prompt
}

func (p *Prompt) SecondaryPrint() {
	if p.secondaryF != nil {
		term.WriteString(p.secondaryF())
		return
	}

	term.WriteString(DefaultSecondaryPrompt)
}

func (p *Prompt) Tooltip(prompt func(word string) string) {
	if prompt == nil {
		return
	}

	p.tooltipF = func() string {
		var tooltipWord string

		shellWords := strings.Split(string(*p.line), " ")

		if len(shellWords) > 0 {
			tooltipWord = shellWords[0]
		}

		return prompt(tooltipWord)
	}
}

func (p *Prompt) Transient(prompt func() string) {
	p.transientF = prompt
}

func (p *Prompt) TransientPrint() {
	if p.transientF == nil {
		return
	}

	term.MoveCursorBackwards(term.GetWidth())
	term.MoveCursorUp(p.PrimaryUsed())
	term.WriteString(term.ClearScreenBelow)

	term.WriteString(p.transientF())
}

func (p *Prompt) UpperPrint() {
	if p.primaryF == nil {
		return
	}

	upper, _ := p.formatPrimaryLines(p.primaryF())
	if upper == "" {
		return
	}

	lines := strings.Split(strings.TrimSuffix(upper, "\n"), "\n")
	for _, line := range lines {
		term.WriteString(line + term.ClearLineAfter + term.NewlineReturn)
	}
}

func (p *Prompt) formatLastPrompt(prompt string) string {
	if !p.opts.GetBool("show-mode-in-prompt") {
		return prompt
	}

	var status string

	switch {
	case p.keymaps.IsEmacs():
		status = p.opts.GetString("emacs-mode-string")
	case p.keymaps.Main() == keymap.ViCommand:
		status = p.opts.GetString("vi-cmd-mode-string")
	case p.keymaps.Main() == keymap.ViInsert:
		status = p.opts.GetString("vi-ins-mode-string")
	}

	status = strings.Trim(status, "\"")

	begin := regexp.MustCompile(`\\1`)
	end := regexp.MustCompile(`\\2`)

	status = begin.ReplaceAllString(status, "")
	status = end.ReplaceAllString(status, "")
	status = strings.ReplaceAll(status, "\\e", "\x1b")

	return status + prompt
}

func (p *Prompt) formatPrimaryLines(prompt string) (multi, lastPrompt string) {

	lines := strings.Split(prompt, "\n")

	if len(lines) > 1 {
		multi = strings.Join(lines[:len(lines)-1], "\n") + "\n"
		lastPrompt = lines[len(lines)-1]
	} else {
		lastPrompt, multi = prompt, ""
	}

	return multi, lastPrompt
}

func (p *Prompt) formatRightPrompt(rprompt string, startColumn int) (prompt string, canPrint bool) {

	termWidth := term.GetWidth()
	promptLen := strutil.RealLength(rprompt)
	padLen := termWidth - startColumn - promptLen

	if startColumn == termWidth {
		padLen = startColumn - promptLen
	}

	canPrint = (startColumn+promptLen < termWidth) || startColumn == termWidth
	if canPrint {
		prompt = fmt.Sprintf("%s%s", strings.Repeat(" ", padLen), rprompt)
	}

	return
}
