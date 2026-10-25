package readline

import (
	"sort"

	"github.com/malivvan/readline/color"
	"github.com/malivvan/readline/inputrc"
	"github.com/malivvan/readline/term"
)

func (rl *Shell) callLastKeyboardMacro() {
	rl.Macros.RunLastMacro()
}

func (rl *Shell) dumpMacros() {
	rl.Display.ClearHelpers()
	term.WriteString("\n")

	defer func() {
		rl.Prompt.PrimaryPrint()
		rl.Display.Refresh()
	}()

	binds := rl.Config.Binds[string(rl.Keymap.Main())]
	if len(binds) == 0 {
		return
	}

	var macroBinds []string

	for keys, bind := range binds {
		if bind.Macro {
			macroBinds = append(macroBinds, inputrc.Escape(keys))
		}
	}

	sort.Strings(macroBinds)

	if rl.Iterations.IsSet() {
		for _, key := range macroBinds {
			action := inputrc.Escape(binds[inputrc.Unescape(key)].Action)
			term.Printf("\"%s\": \"%s\"\n", key, action)
		}
	} else {
		for _, key := range macroBinds {
			action := inputrc.Escape(binds[inputrc.Unescape(key)].Action)
			term.Printf("%s outputs %s\n", key, action)
		}
	}
}

func (rl *Shell) endKeyboardMacro() {
	rl.Macros.StopRecord()
}

func (rl *Shell) macroRun() {
	done := rl.Keymap.PendingCursor()
	defer done()

	rl.Hint.SetTemporary(color.Dim + "Run (macro arg)")
	rl.Display.Refresh()

	key, isAbort := rl.Keys.ReadKey()
	if isAbort {
		return
	}

	rl.Macros.RunMacro(key)
}

func (rl *Shell) macroToggleRecord() {
	if rl.Macros.Recording() {
		rl.Macros.StopRecord()

		return
	}

	done := rl.Keymap.PendingCursor()
	defer done()

	rl.Hint.SetTemporary(color.Dim + "REC (macro arg)")
	rl.Display.Refresh()

	key, isAbort := rl.Keys.ReadKey()
	if isAbort {
		return
	}

	rl.Macros.StartRecord(key)
}

func (rl *Shell) printLastKeyboardMacro() {
	rl.Display.ClearHelpers()

	rl.Macros.PrintLastMacro()

	rl.Prompt.PrimaryPrint()
	rl.Display.Refresh()
}

func (rl *Shell) startKeyboardMacro() {
	rl.Macros.StartRecord(rune(0))
}
