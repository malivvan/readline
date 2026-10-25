package readline

import (
	"fmt"
	"io"
	"sort"
	"strings"
	"unicode"

	"github.com/malivvan/readline/color"
	"github.com/malivvan/readline/inputrc"
	"github.com/malivvan/readline/keymap"
	"github.com/malivvan/readline/strutil"
	"github.com/malivvan/readline/term"
)

func (rl *Shell) abort() {

	rl.Iterations.Reset()
	rl.selection.Reset()

	if rl.completer.AutoCompleting() || rl.completer.IsInserting() {
		rl.Hint.Reset()
		rl.completer.ResetForce()

		return
	}

	searching, _, _ := rl.completer.NonIncrementallySearching()
	if searching {
		rl.completer.NonIsearchStop()
		return
	}

	if !rl.Keymap.InputIsTerminator() {
		return
	}

	if rl.Config.GetBool("echo-control-characters") {
		key := rl.Keys.Caller()
		if key[0] == rune(inputrc.Unescape(`\C-C`)[0]) {
			quoted, _ := strutil.Quote(key[0])
			term.WriteString(string(quoted))
		}
	}

	rl.Display.AcceptLine()
	rl.History.Accept(false, false, ErrInterrupt)
}

func (rl *Shell) clearDisplay() {
	rl.History.SkipSave()

	term.WriteString(term.CursorTopLeft)
	term.WriteString(term.ClearDisplay)

	rl.Display.PrintPrimaryPrompt()
}

func (rl *Shell) clearScreen() {
	rl.History.SkipSave()

	term.WriteString(term.CursorTopLeft)
	term.WriteString(term.ClearScreen)

	rl.Display.PrintPrimaryPrompt()
}

func (rl *Shell) doLowercaseVersion() {
	rl.History.SkipSave()

	keys := rl.Keys.Caller()

	escapePrefix := false

	if len(keys) > 1 && keys[0] == inputrc.Esc {
		escapePrefix = true
		keys = keys[1:]
	} else if len(keys) == 1 && inputrc.IsMeta(keys[0]) {
		keys = []rune{inputrc.Demeta(keys[0])}
	}

	if unicode.IsLower(keys[0]) {
		return
	}

	keys[0] = unicode.ToLower(keys[0])

	if escapePrefix {
		input := append([]rune{inputrc.Esc}, keys...)
		rl.Keys.Feed(false, input...)
	} else {
		rl.Keys.Feed(false, inputrc.Enmeta(keys[0]))
	}
}

func (rl *Shell) dumpFunctions() {
	rl.Display.ClearHelpers()
	term.WriteString("\n")

	defer func() {
		rl.Prompt.PrimaryPrint()
		rl.Display.Refresh()
	}()

	inputrcFormat := rl.Iterations.IsSet()
	rl.Keymap.PrintBinds(string(rl.Keymap.Main()), inputrcFormat)
}

func (rl *Shell) dumpVariables() {
	rl.Display.ClearHelpers()
	term.WriteString("\n")

	defer func() {
		rl.Prompt.PrimaryPrint()
		rl.Display.Refresh()
	}()

	variables := make([]string, 0, len(rl.Config.Vars))

	for variable := range rl.Config.Vars {
		variables = append(variables, variable)
	}

	sort.Strings(variables)

	if rl.Iterations.IsSet() {
		for _, variable := range variables {
			value := rl.Config.Vars[variable]
			term.Printf("set %s %v\n", variable, value)
		}
	} else {
		for _, variable := range variables {
			value := rl.Config.Vars[variable]
			term.Printf("%s is set to `%v'\n", variable, value)
		}
	}
}

func (rl *Shell) editAndExecuteCommand() {
	buffer := *rl.line

	edited, err := rl.Buffers.EditBuffer(buffer, "", "", rl.Keymap.IsEmacs())
	if rl.editorFailed(buffer, edited, err) {
		return
	}

	rl.line.Set(edited...)
	rl.Display.AcceptLine()
	rl.History.Accept(false, false, nil)
}

func (rl *Shell) editCommandLine() {
	buffer := *rl.line
	keymapCur := rl.Keymap.Main()

	edited, err := rl.Buffers.EditBuffer(buffer, "", "", rl.Keymap.IsEmacs())
	if rl.editorFailed(buffer, edited, err) {
		return
	}

	rl.line.Set(edited...)

	switch keymapCur {
	case keymap.Emacs, keymap.EmacsStandard, keymap.EmacsMeta, keymap.EmacsCtrlX:
		rl.emacsEditingMode()
	}
}

func (rl *Shell) emacsEditingMode() {

	rl.selection.Reset()
	rl.Iterations.Reset()
	rl.Buffers.Reset()

	rl.Hint.Reset()
	rl.completer.Reset()
	rl.line, rl.cursor, rl.selection = rl.completer.GetBuffer()

	rl.Keymap.SetMain(keymap.Emacs)
}

func (rl *Shell) endOfFile() {
	switch rl.line.Len() {
	case 0:
		rl.Display.AcceptLine()
		rl.History.Accept(false, false, io.EOF)
	default:
		rl.deleteChar()
	}
}

func (rl *Shell) prefixMeta() {
	rl.History.SkipSave()

	done := rl.Keymap.PendingCursor()
	defer done()

	key, isAbort := rl.Keys.ReadKey()
	if isAbort {
		return
	}

	keys := append([]rune{inputrc.Esc}, key)
	rl.Keys.Feed(false, keys...)
}

func (rl *Shell) reReadInitFile() {
	main := rl.Keymap.Main()

	err := rl.Keymap.ReloadConfig(rl.Opts...)
	if err != nil {
		rl.Hint.SetTemporary(color.FgRed + "Inputrc reload error: " + err.Error())
		return
	}

	defer rl.Keymap.UpdateCursor()

	newMain := rl.Keymap.Main()

	if main != newMain {
		switch newMain {
		case keymap.Emacs, keymap.EmacsStandard, keymap.EmacsMeta, keymap.EmacsCtrlX:
			rl.emacsEditingMode()
		case keymap.Vi, keymap.ViCommand, keymap.ViMove:
			rl.viCommandMode()
		case keymap.ViInsert:
			rl.viInsertMode()
		}
	}

	rl.Hint.SetTemporary(color.FgGreen + "Inputrc reloaded")
}

// editorFailed reports whether the buffer returned by the system editor must be
// discarded and prints the matching hint. An editor that empties the buffer is
// treated as an error, but that case has a nil error and must not dereference
// it.
func (rl *Shell) editorFailed(buffer, edited []rune, err error) bool {
	empty := len(edited) == 0 && len(buffer) != 0

	if err == nil && !empty {
		return false
	}

	rl.History.SkipSave()

	message := "editor returned an empty buffer"
	if err != nil {
		message = strings.ReplaceAll(err.Error(), "\n", "")
	}

	rl.Hint.SetTemporary(fmt.Sprintf(color.FgRed+"Editor error: %s", message))

	return true
}
