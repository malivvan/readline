> **Runnable example:** [`cmd/keymaps`](../cmd/keymaps)


## Introduction

Underneath, readline programs and libraries are one of their kind.
They face several challenges that many other pieces of software don't:
- They must handle a plethora of details and tell absolutely everything
  the program and system should do, for everything.
- They make heavy use of arrays and indexing, since the input buffer is
  an array by essence. Therefore, the shell must be very cautious about
  indexing errors.
- Unlike many other programs, even the smallest mistake can make the whole
  session entirely unusable.

For all these reasons, programming readline libraries or extending them is
quite a risky endeavour, and even more in legacy libraries written in C.
And while Go helps having an easier codebase and programming practices, it 
cannot safeguard programmers from various mistakes, such as wrong indexing.

Accordingly, this `readline` library goes a step further so as to provide bold programmers 
the means to safely (and to some extent, easily) extend the shell's functionality, by 
encapsulating the various shell components and offering various methods to manipulate them.
This goes for things like the input line, the cursor, the selections, history sources, macro
and display engines, the completion engine, the register buffers, keymaps and commands.

This page thus explains how to access those components and use them to extend the shell.
For developers who would want to develop such extensions, it also highly advised to go through
the codebase: the latter and its components are heavily commented, so as to facilitate using them.


## Accessing readline components

Below are a few examples of the various readline components and some of the methods
they might offer. We start with the core things: the line, cursor and selection.
Several things should be noted:
- The line offers methods to set, modify and access itself safely. This is so that
  there is no risk of using invalid index positions when doing so.
- The same applies to the cursor: moving it through methods ensures that it can NEVER
  be in an invalid position (negative, or longer than the line buffer).

```go
// Since in some special modes, such as incremental-search, the shell uses a minibuffer,
// the input line and cursor of interest may vary, accessing those can only be done by
// asking the shell to return the relevant line/cursor/selection in the current context:
line, cursor, sel := shell.Line(), shell.Cursor(), shell.Selection()

// Line
line.Set([]rune("an input line"))     // Set the input line buffer.
line.Cut(0, 10)                       // Delete the first 10 characters in the line.
line.Insert(100, []rune("foo")...)    // Insert `foo` at the index 100 of the line, if possible.

// Cursor
cursor.Set(line.Len()) // Put the cursor in append mode.
cursor.InsertAt('b')   // Insert a character at the current cursor position.
cursor.Move(-100)      // Move the cursor backward by 100 (or as much of it as possible).
cursor.LineMove(-1)    // Move up a line if the input buffer is a multiline one.
cursor.Char()          // Return the character under cursor, if any.

// Selection
sel.Mark(cursor.Pos()) // Start a selection from the current cursor position.
sel.Visual(false)      // Set the selection as the visual one.
sel.Text()             // Get the text of the current selection.
sel.MarkRange(0, 10)   // Mark a range of text.
sel.SelectABlankWord() // Select the bigword under the cursor.
```

We continue with some other core components:
```go
// Buffers (eg. Vim registers)
buf := shell.Buffers

buf.SetActive('b') // Set the active register to 'b' (if 'B', use it in append-mode)
buf.Active()       // The currently selected register (user just typed "<key>)
buf.Get('c')       // Get the contents of register 'c', if any.

buf.Write([]rune("hello world")...)        // Set the contents of the active register
buf.WriteTo('c', []rune("hello world")...) // Set the contents of register 'c'
buf.Reset()                                // Reset the active register (do not clear it, but unselect it)


// History sources
hist := rl.History

// Line buffer state
hist.Undo() // Undo the last line change/insertion.
hist.Redo() // Redo the last line undo change, if any.
hist.Save() // Save the current state of the line in the undo history.

// Previous command lines
hist.Walk(-1)     // Walk to the previous history entry.
hist.Walk(+1)     // Walk to the next history entry.
hist.Write(false) // Write the current line to the history.
```

Finally, we show a few things about some display-related helpers:
```go
// Hint section, allowing to display usage indications.
hint := shell.Hint

hint.Set("usage about this command")   // Set the hint message.
hint.SetTemporary("temporary message") // Will be cleared on the next keypress.
hint.Persist("persisted message")      // Will remain until hint.ResetPersist()
hint.Reset()                           // Reset the non-persistent part of the hint.

// The display engine allows to refresh the interface, move the cursor 
// to some specific places, clear some helpers (hints and completions), etc.
rl.Display.Refresh()      // Refresh the entire shell interface.
rl.Display.ClearHelpers() // Clears all hints and completions until next recomputation.
```


## The keymap engine

The keymap engine is another component of the shell, accessible and used just like the others.
It stores all available keymaps, the currently active ones, as well as all the registered command 
functions. It also provides methods to dispatch keys against binds, run pending commands, and some other stuff.

Many of the methods it offers will likely not be needed by developers, except a few ones:
```go
keymap := shell.Keymap

// Accessing and setting the local and main keymaps
// Note that even these calls will likely not be needed unless you are using custom keymaps: 
// the shell already has builtin commands to use/switch to builtin keymaps (emacs/vim modes/etc). 
// Please refer to the `Keymaps & Commands` documentation page for a list of them.

keymap.SetLocal("menu-select")  // Set the local keymap to the completion keymap.
keymap.Local()                  // Get the name of the local keymap
keymap.SetMain("vi-move")       // Set the main keymap.
keymap.Main()                   // Get the name of the main keymap

// Display-related
keymap.UpdateCursor()           // Refreshes the cursor style depending on the current keymap.
keymap.PrintCursor("vi-move")   // Print a specific cursor
keymap.PendingCursor()          // Temporarily print a pending (underline) cursor for argument commands/key reading.

// More advanced calls
keymap.Pending() // If called from within a command function, this command will be reexecuted after an argument command is found.
```

You might however need to add some commands.

Adding command functions to the shell can be done with the following code:
```go
// Commands is map of "command-name":func(), storing all commands.
// The builtin commands func() are all methods of the *Shell type.
// Example: "vi-movement-mode": func (rl *Shell) vimMovementMode() {...}
//          "down-history":     func (rl *Shell) downHistory() {...}
commands := shell.Keymap.Commands()

// The most simple (and useless) example of adding a command would be this:
shell.Keymap.Commands()["custom-command"] = func() {}

// Or adding sets of commands at once.
keymap.Register(map[string]func(){
    "completion-command":   func() {...},
    "insertion-command":    func() {...},
    "modification-command": func() {...},
})
```

Your commands might also need to register new bind mappings (or new keymaps altogether).

This is done with the `inputrc` configuration object. See [configuration file](./CONFIGURATION-FILE.md) page for more
details on this.


## Example 1: Adding a URL-selection command

We finally go through an example demonstrating a custom command registered with the shell.
This command's role is to select some region of text under/around the cursor, and to find
some pattern within this region and, if any match is found, to narrow the selection to this.

This is essentially what the builtin `select-keyword-next`/`select-keyword-prev` commands do: 
it selects the blank word under the cursor, and uses regular expressions to find common patterns
within this word: URLs, and subsequently each component of it: domain, path and parameters. 

Users can then cycle through these various groups/subgroups by repetitively calling this command.

Here is the code:
```go
func (rl *Shell) selectKeywordNext() {
    // Don't add a new state to the line changes history.
    rl.History.SkipSave()

    // Get the start and end positions of the blank word under cursor.
    bpos, epos := rl.line.SelectBlankWord(rl.cursor.Pos())

    // Use an internal (quite long/complex) function to narrow 
    // down the selection to some URL regexp group/subgroup.
    _, epos, match := rl.selection.SelectKeyword(bpos, epos, true)
    if !match {
        return
    }

    // The matchers succeeded, we now have a selection active,
    // but the cursor should be moved to the end of it.
    rl.cursor.Set(epos)
    rl.selection.Visual(false)
}
```

We could then register this command to the shell, and users could either add a bind for it 
in their inputrc file, or you could add a bind to the in-memory inputrc configuration for it.
```go
// Or adding sets of commands at once.
shell.Keymap.Register(map[string]func(){
    "url-select":   rl.selectKeywordNext,
})

// Add a bind in the Vim command keymap
// Key sequences are stored unescaped, exactly as the shell matches them.
shell.Config.Bind("vi", inputrc.Unescape(`\C-X\C-u`), "url-select", false)
```

## Example 2: Adding a surround-change command

This other example is also taken from the readline code.
The following command is `vi-select-surround`: it reads a key from the keyboard, and attempts
to find a pair of such character surrounding the current cursor position. 

An example use in Vim command mode is `cs"'`:
- `c` triggers the `vi-change-to` command, which marks itself as waiting for a pending/argument command to run first.
- `s` triggers the `vi-select-surround` command.
- The latter reads a key from the keyboard, here `"`, and attempts to select the closest surrounding double quotes.
- The command exits, and control is returned to `vi-change-to`, which detects the surround selection.
- `vi-change-to` reads a last key, here `'`, and replaces the double quotes by single ones.

For the sake of precision, and because the two commands `vi-change-to` and `vi-select-surround` illustrate
yet more of the shell potential, both commands' code is reproduced below.

Starting with `vi-change-to`:
```go
func (rl *Shell) viChangeTo() {
	switch {
	case rl.Keymap.IsPending():
		// In vi operator pending mode, it's that we've been called
		// twice in a row (eg. `cc`), so copy the entire current line.
		rl.Keymap.CancelPending()
		rl.History.Save()

		rl.selection.Mark(rl.cursor.Pos())
		rl.selection.Visual(true)
		rl.selection.Cut()
		rl.viInsertMode()

	case len(rl.selection.Surrounds()) == 2:
		// In surround selection mode, change the surrounding chars.
		rl.Display.Refresh()
		defer rl.selection.Reset()

		// Now read another key
		done := rl.Keymap.PendingCursor()
		defer done()

		rchar, isAbort := rl.Keys.ReadKey()
		if isAbort {
			return
		}

		rl.History.Save()

		// There might be a matching equivalent.
		bchar, echar := strutil.MatchSurround(rchar)

		surrounds := rl.selection.Surrounds()

		bpos, _ := surrounds[0].Pos()
		epos, _ := surrounds[1].Pos()

		(*rl.line)[bpos] = bchar
		(*rl.line)[epos] = echar

	case rl.selection.Active():
		// In visual mode, we have just have a selection to delete.
		rl.History.Save()

		rl.adjustSelectionPending()
		cpos := rl.selection.Cursor()
		cut := rl.selection.Cut()
		rl.Buffers.Write([]rune(cut)...)
		rl.cursor.Set(cpos)

		rl.viInsertMode()

	default:
		// Since we must emulate the default readline behavior,
		// we vary our behavior depending on the caller key.
		keys := rl.Keys.Caller()

		switch keys[0] {
		case 'c':
			rl.Keymap.Pending()
			rl.selection.Mark(rl.cursor.Pos())
		case 'C':
			rl.viChangeEol()
		}
	}
}
```

Following, the `vi-select-surround` code:
```go
func (rl *Shell) viSelectSurround() {
	rl.History.SkipSave()

	// Read a key as a rune to search for
	done := rl.Keymap.PendingCursor()
	defer done()

	char, isAbort := rl.Keys.ReadKey()
	if isAbort { // If escape has been pressed
		return
	}

	// Find the corresponding enclosing chars
	bpos, epos, _, _ := rl.line.FindSurround(char, rl.cursor.Pos())
	if bpos == -1 || epos == -1 {
		return
	}

	// Add those two positions to highlighting and give back
    // control to whatever command is waiting for us to return,
    // or simply go back to normal execution workflow.
	rl.selection.MarkSurround(bpos, epos)
}
```

## Key-event callbacks

A recurring need when embedding the shell is to *observe a key sequence, inspect a snapshot
of the editor state, optionally modify the line/cursor/hint, and optionally fall through to
the key's original binding* — without patching keymaps indirectly or replaying synthetic
keystrokes.

Everything needed for this is already covered above: a **key event is just a command of your
own, bound to the sequence you care about**. Inside it you read the editor through the
accessors shown earlier (`Line`, `Cursor`, `Keys.Caller`, `Keymap.Main`), mutate it through
the same setters, and — to "continue with the original binding" — look up the command that
was bound on that sequence before and call it (`Keymap.Commands()[name]`).

The small helper below packages that into an `AddKeyEvent`/`RemoveKeyEvent` API. It is pure
user code; **the library needs no additions for this**:

```go
import (
    "github.com/malivvan/readline"
    "github.com/malivvan/readline/inputrc"
)

type KeyEventState struct {
    Line      string // current input line
    CursorPos int    // cursor position in the line
    KeyPress  string // the sequence that triggered the callback
    InputMode string // active editing mode ("emacs", "vi-insert", ...)
}

type KeyEventResult struct {
    SetLine  []rune // when non-nil, replaces the input line
    SetPos   int    // when >= 0, moves the cursor to this position
    HintText []rune // when non-nil, shown in the hint area
    Continue bool   // when true, also run the key's original binding
}

type KeyEvents struct {
    rl     *readline.Shell
    keymap string
    orig   map[string]inputrc.Bind
}

func NewKeyEvents(rl *readline.Shell) *KeyEvents {
    return &KeyEvents{rl: rl, keymap: string(rl.Keymap.Main()), orig: make(map[string]inputrc.Bind)}
}

func (k *KeyEvents) AddKeyEvent(seq string, cb func(arg int, state *KeyEventState) *KeyEventResult) error {
    useq := inputrc.Unescape(seq)
    k.orig[useq] = k.rl.Config.Binds[k.keymap][useq] // remember the original bind
    name := "key-event:" + useq

    k.rl.Keymap.Register(map[string]func(){name: func() {
        orig := k.orig[useq]

        state := &KeyEventState{ // inspect a snapshot
            Line:      string(*k.rl.Line()),
            CursorPos: k.rl.Cursor().Pos(),
            KeyPress:  string(k.rl.Keys.Caller()),
            InputMode: string(k.rl.Keymap.Main()),
        }

        res := cb(k.rl.Iterations.Get(), state)
        if res == nil {
            return
        }

        if res.SetLine != nil {            // modify line
            k.rl.Line().Set(res.SetLine...)
        }
        if res.SetPos >= 0 {               // modify cursor
            k.rl.Cursor().Set(res.SetPos)
        }
        if res.HintText != nil {           // modify hint
            k.rl.Hint.Set(string(res.HintText))
        }
        if res.Continue && orig.Action != "" && !orig.Macro { // continue w/ original
            if cmd := k.rl.Keymap.Commands()[orig.Action]; cmd != nil {
                cmd()
            }
        }
    }})

    return k.rl.Config.Bind(k.keymap, useq, name, false)
}

func (k *KeyEvents) RemoveKeyEvent(seq string) error {
    useq := inputrc.Unescape(seq)
    orig := k.orig[useq]
    delete(k.orig, useq)

    return k.rl.Config.Bind(k.keymap, useq, orig.Action, orig.Macro) // restore the original
}
```

Usage:
```go
events := NewKeyEvents(shell)

// Observe Ctrl-G: show a hint, then continue with whatever Ctrl-G did before.
events.AddKeyEvent(`\C-g`, func(arg int, s *KeyEventState) *KeyEventResult {
    return &KeyEventResult{HintText: []rune("Ctrl-G on: " + s.Line), SetPos: -1, Continue: true}
})

// Transform the line on Ctrl-U (no fall-through).
events.AddKeyEvent(`\C-u`, func(arg int, s *KeyEventState) *KeyEventResult {
    return &KeyEventResult{SetLine: []rune(strings.ToUpper(s.Line)), SetPos: len(s.Line)}
})
```

A few notes:
- **Escaping**: pass sequences through `inputrc.Unescape` before binding; the keymap matches
  raw key bytes, not the `\C-g` spelling (the helper does this for you).
- **Which keymap**: the helper binds into the current main keymap (`Keymap.Main()`); register
  into each keymap you care about (`emacs`, `vi-insert`, `vi-command`) if users switch modes.
- **Continuing a macro bind**: if the original bind was a macro (`bind.Macro`), there is no
  command to call — re-inject its keys with `rl.Keys.Feed(true, ...)` instead.
- **Reading more keys** for multi-key gestures, and `Keymap.PendingCursor()`, work exactly as
  in *Example 2* above.

This is the same set of building blocks as the two examples above, composed into a reusable
hook — which is why no new library API is required to support programmable key events.
