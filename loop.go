package readline

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/malivvan/readline/color"
	"github.com/malivvan/readline/completion"
	"github.com/malivvan/readline/core"
	"github.com/malivvan/readline/display"
	"github.com/malivvan/readline/history"
	"github.com/malivvan/readline/inputrc"
	"github.com/malivvan/readline/keymap"
	"github.com/malivvan/readline/macro"
	"github.com/malivvan/readline/term"
)

var ErrInterrupt = errors.New(os.Interrupt.String())

// inputDescriptor returns the file descriptor of the given reader, if any.
func inputDescriptor(input io.Reader) (int, bool) {
	if f, ok := input.(interface{ Fd() uintptr }); ok {
		return int(f.Fd()), true
	}

	return 0, false
}

func (rl *Shell) Readline() (string, error) {
	rl.applyIO()

	descriptor, hasDescriptor := inputDescriptor(rl.Stdin)
	if hasDescriptor && term.IsTerminal(descriptor) {
		state, err := term.MakeRaw(descriptor)
		if err != nil {
			return "", err
		}

		defer func() { _ = term.Restore(descriptor, state) }()

		if rl.Config.GetBool("enable-bracketed-paste") {
			term.EnableBracketedPaste()
			defer term.DisableBracketedPaste()
		}
	}

	rl.Display.PrintPrimaryPrompt()
	defer rl.Display.RefreshTransient()
	defer term.WriteString(keymap.CursorStyle("default").String())

	rl.init()

	resize := display.WatchResize(rl.Display)
	defer close(resize)

	rl.Keys.InitWake()
	defer rl.Keys.CloseWake()

	for {

		macro.RecordKeys(rl.Macros)

		core.FlushUsed(rl.Keys)

		rl.completer.ApplyRegen()

		rl.Display.Refresh()

		core.WaitAvailableKeys(rl.Keys, rl.Config)

		if err := rl.Keys.ReadError(); err != nil {
			return "", err
		}

		if rl.Keys.IsEOF() {
			return "", io.EOF
		}

		if rl.Keys.Empty() {
			continue
		}

		bind, command, prefixed := keymap.MatchLocal(rl.Keymap)
		if prefixed {
			continue
		}

		accepted, line, err := rl.run(false, bind, command)
		if accepted {
			return line, err
		} else if command != nil {
			continue
		}

		completion.UpdateInserted(rl.completer)

		bind, command, prefixed = keymap.MatchMain(rl.Keymap)
		if prefixed {
			continue
		}

		accepted, line, err = rl.run(true, bind, command)
		if accepted {
			return line, err
		}

		rl.handleUndefined(bind, command)
	}
}

func (rl *Shell) execute(command func()) {
	if command != nil {
		command()
	}

	if !rl.Iterations.IsPending() {
		rl.Keymap.RunPending()
	}

	switch rl.Keymap.Main() {
	case keymap.ViCommand, keymap.ViMove, keymap.Vi:
		rl.cursor.CheckCommand()
	default:
		rl.cursor.CheckAppend()
	}
}

func (rl *Shell) handleUndefined(bind inputrc.Bind, cmd func()) {
	if bind.Action != "" || cmd != nil {
		return
	}

	if rl.Keymap.Local() == keymap.Isearch {
		rl.Hint.Reset()
		rl.completer.Reset()
	}
}

func (rl *Shell) init() {

	core.FlushUsed(rl.Keys)
	rl.line.Set()
	rl.cursor.Set(0)
	rl.cursor.ResetMark()
	rl.selection.Reset()
	rl.Buffers.Reset()
	rl.History.Reset()
	rl.Iterations.Reset()

	history.Init(rl.History)
	rl.History.Save()

	rl.Hint.Reset()
	rl.completer.ResetForce()
	display.Init(rl.Display, rl.SyntaxHighlighter)
}

func (rl *Shell) run(main bool, bind inputrc.Bind, command func()) (bool, string, error) {

	if !main && bind.Action == "" {
		return false, "", nil
	}

	if bind.Macro {
		macro := inputrc.Unescape(bind.Action)
		rl.Keys.Feed(false, []rune(macro)...)
	}

	rl.line, rl.cursor, rl.selection = rl.completer.GetBuffer()

	rl.execute(command)

	rl.updatePosRunHints()

	rl.completer.UpdateIsearch()

	rl.line, rl.cursor, rl.selection = rl.completer.GetBuffer()

	rl.History.SaveWithCommand(bind)

	return rl.History.LineAccepted()
}

func (rl *Shell) updatePosRunHints() {
	hint := core.ResetPostRunIterations(rl.Iterations)
	register, selected := rl.Buffers.IsSelected()

	if hint == "" && !selected && !rl.Macros.Recording() {
		rl.Hint.ResetPersist()
		return
	}

	if hint != "" {
		rl.Hint.Persist(hint)
	} else if selected {
		rl.Hint.Persist(color.Dim + fmt.Sprintf("(register: %s)", register))
	}
}
