package readline

import (
	"fmt"
	"io"
	"os"

	"github.com/malivvan/readline/completion"
	"github.com/malivvan/readline/core"
	"github.com/malivvan/readline/display"
	"github.com/malivvan/readline/editor"
	"github.com/malivvan/readline/history"
	"github.com/malivvan/readline/inputrc"
	"github.com/malivvan/readline/keymap"
	"github.com/malivvan/readline/macro"
	"github.com/malivvan/readline/term"
	"github.com/malivvan/readline/ui"
)

type Shell struct {
	line       *core.Line
	cursor     *core.Cursor
	selection  *core.Selection
	Iterations *core.Iterations
	Buffers    *editor.Buffers
	Keys       *core.Keys
	Keymap     *keymap.Engine
	History    *history.Sources
	Macros     *macro.Engine

	Config    *inputrc.Config
	Opts      []inputrc.Option
	Prompt    *ui.Prompt
	Hint      *ui.Hint
	completer *completion.Engine
	Display   *display.Engine

	AcceptMultiline func(line []rune) (accept bool)

	SyntaxHighlighter func(line []rune) string

	PasteTransformer func(text string) string

	Completer func(line []rune, cursor int) Completions

	// Stdin, Stdout and Stderr are the shell streams. They default to the
	// process standard streams and are propagated to every component; assign
	// them (or use SetIO) before calling Readline.
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
}

func NewShell(opts ...inputrc.Option) *Shell {
	shell := new(Shell)

	keys := new(core.Keys)
	line := new(core.Line)
	cursor := core.NewCursor(line)
	selection := core.NewSelection(line, cursor)
	iterations := new(core.Iterations)

	shell.Keys = keys
	shell.line = line
	shell.cursor = cursor
	shell.selection = selection
	shell.Buffers = editor.NewBuffers()
	shell.Iterations = iterations

	keymaps, config := keymap.NewEngine(keys, iterations, opts...)
	keymaps.Register(shell.standardCommands())
	keymaps.Register(shell.viCommands())
	keymaps.Register(shell.historyCommands())
	keymaps.Register(shell.completionCommands())

	shell.Keymap = keymaps
	shell.Config = config
	shell.Opts = opts

	hint := ui.NewHint(keys)
	prompt := ui.NewPrompt(line, cursor, keymaps, config)
	macros := macro.NewEngine(keys, hint)
	history := history.NewSources(line, cursor, hint, config)
	completer := completion.NewEngine(hint, keymaps, config)
	completion.Init(completer, keys, line, cursor, selection, shell.commandCompletion)

	display := display.NewEngine(keys, selection, history, prompt, hint, completer, config)

	shell.Config = config
	shell.Hint = hint
	shell.Prompt = prompt
	shell.completer = completer
	shell.Macros = macros
	shell.History = history
	shell.Display = display

	shell.applyIO()

	return shell
}

// SetIO replaces the shell streams. A nil stream restores the corresponding
// process standard stream.
func (rl *Shell) SetIO(stdin io.Reader, stdout, stderr io.Writer) {
	rl.Stdin = stdin
	rl.Stdout = stdout
	rl.Stderr = stderr

	rl.applyIO()
}

// applyIO propagates the shell streams to the components that read or write.
func (rl *Shell) applyIO() {
	if rl.Stdin == nil {
		rl.Stdin = os.Stdin
	}
	if rl.Stdout == nil {
		rl.Stdout = os.Stdout
	}
	if rl.Stderr == nil {
		rl.Stderr = os.Stderr
	}

	rl.Keys.SetInput(rl.Stdin)
	rl.Keys.SetOutput(rl.Stdout)
	rl.Keys.SetErrorOutput(rl.Stderr)
	rl.Buffers.SetIO(rl.Stdin, rl.Stdout, rl.Stderr)

	term.SetOutput(rl.Stdout)
	term.SetTerminalFile(terminalFile(rl.Stdout))
}

// terminalFile returns the file used for terminal size queries: the output
// stream when it is a terminal, the process stderr otherwise, or nil when
// neither is a terminal.
func terminalFile(out io.Writer) *os.File {
	if f, ok := out.(*os.File); ok && term.IsTerminal(int(f.Fd())) {
		return f
	}

	if term.IsTerminal(int(os.Stderr.Fd())) {
		return os.Stderr
	}

	return nil
}

// output returns the writer used by the shell for its own messages.
func (rl *Shell) output() io.Writer {
	if rl.Stdout != nil {
		return rl.Stdout
	}

	return os.Stdout
}

func (rl *Shell) Cursor() *core.Cursor { return rl.cursor }

func (rl *Shell) Line() *core.Line { return rl.line }

func (rl *Shell) PrintTransientf(msg string, args ...any) (n int, err error) {

	rl.Display.CursorToLineStart()
	term.MoveCursorBackwards(term.GetWidth())
	term.MoveCursorUp(rl.Prompt.PrimaryUsed())
	term.WriteString(term.ClearScreenBelow)

	n, err = fmt.Fprintf(rl.output(), msg+"\n", args...)

	rl.Prompt.PrimaryPrint()
	rl.Display.Refresh()

	return
}

func (rl *Shell) Printf(msg string, args ...any) (n int, err error) {

	rl.Display.CursorBelowLine()
	term.MoveCursorBackwards(term.GetWidth())
	term.WriteString(term.ClearScreenBelow)

	n, err = fmt.Fprintf(rl.output(), msg+"\n", args...)

	rl.Prompt.PrimaryPrint()
	rl.Display.Refresh()

	return
}

func (rl *Shell) Selection() *core.Selection { return rl.selection }
