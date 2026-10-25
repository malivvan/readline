// Package readline provides a pure Go, cross-platform line editor with full
// .inputrc support and extended Emacs and Vi editing modes.
//
// A shell is created with NewShell and driven by a Readline loop:
//
//	shell := readline.NewShell()
//	shell.Prompt.Primary(func() string { return "> " })
//
//	for {
//		line, err := shell.Readline()
//		if err != nil {
//			return
//		}
//		fmt.Println(line)
//	}
//
// Readline returns io.EOF when the input is closed (Ctrl-D on an empty line)
// and ErrInterrupt when a line is aborted with Ctrl-C.
//
// Behavior is configured through .inputrc options and binds (see Config and
// Shell.Config), and extended through the exported components of the Shell:
//
//   - Line, Cursor and Selection give safe access to the input buffer.
//   - Completer returns Completions for the current line; the container offers
//     values, descriptions, tags, styles, usage hints and messages.
//   - History holds one or more history sources (see NewHistoryFromFile and
//     NewInMemoryHistory).
//   - Keymap registers commands and dispatches keys; Config binds them.
//   - Prompt and Hint control the interface; Display exposes rendering.
//   - Macros records and replays keyboard macros; Buffers holds Vi registers.
//
// The hooks AcceptMultiline, SyntaxHighlighter, PasteTransformer and
// Completions, along with the inline suggestion methods, let applications
// customize edition without touching the editor internals.
//
// Runnable examples for every feature live in the cmd/ directory.
package readline
