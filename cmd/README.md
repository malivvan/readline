# Readline examples

Every feature of `github.com/malivvan/readline` has a small, self-contained
program in this directory. Each example is a `main` package that builds and runs
on its own, with no dependency other than the library and its (few) direct
dependencies.

```console
# run any example straight from the repository
go run ./cmd/basics

# or build them all
go build ./cmd/...
```

The examples are meant to be read as much as run: each file starts with a
comment describing what it demonstrates, and stays short enough to be read in
one sitting.

## Index

| Example | What it demonstrates | Related documentation |
|---------|----------------------|-----------------------|
| [`basics`](./basics) | The `Readline()` loop, the primary prompt, `ErrInterrupt` and `io.EOF`. | [Introduction](../docs/INTRODUCTION.md) |
| [`completions`](./completions) | The `Completions` container: values, descriptions, tags, styles, usage hints, error messages, and asynchronous menu refresh. | [Completions & Hints](../docs/COMPLETIONS-&-HINTS.md) |
| [`history`](./history) | File-backed and in-memory history sources, multiple sources at once, cycling and deletion. | [History Sources](../docs/HISTORY-SOURCES.md) |
| [`keymaps`](./keymaps) | Registering custom commands and binds, reading/mutating line and cursor, switching Emacs/Vi at runtime. | [Programmable Commands](../docs/PROGRAMMABLE-COMMANDS.md), [Keymaps & Commands](../docs/KEYMAPS-&-COMMANDS.md) |
| [`highlighting`](./highlighting) | The `SyntaxHighlighter` callback (commands, quotes and comments). | [Syntax Highlighting](../docs/SYNTAX-HIGHLIGHTING.md) |
| [`multiline`](./multiline) | The `AcceptMultiline` callback and the secondary prompt. | [Multiline](../docs/MULTILINE.md) |
| [`prompts`](./prompts) | Primary/secondary/right/transient prompts, tooltips, and both logging methods. | [Prompts](../docs/PROMPTS.md) |
| [`suggestions`](./suggestions) | Inline (ghost text) suggestions, recomputed by wrapping builtin commands. | [Inline Suggestions](../docs/INLINE-SUGGESTIONS.md) |
| [`paste`](./paste) | The `PasteTransformer` hook: multi-line pastes become expandable placeholders. | [Bracketed Paste](../docs/BRACKETED-PASTE.md) |
| [`options`](./options) | Reading and writing options, loading an extra `.inputrc`, reloading the configuration. | [Options](../docs/OPTIONS.md), [Configuration File](../docs/CONFIGURATION-FILE.md) |
| [`macros`](./macros) | Macro binds, the builtin recording keys, and replaying the last macro from the application. | [Macros](../docs/MACROS.md) |

## basics

The smallest useful program: prompt, read, print.

```go
shell := readline.NewShell()
shell.Prompt.Primary(func() string { return "basics> " })

line, err := shell.Readline()
```

`Readline` returns `readline.ErrInterrupt` when the line is aborted with
Ctrl-C, and `io.EOF` when Ctrl-D is pressed on an empty buffer. A shell loop
must handle both and keep the shell usable after an interrupt.

## completions

The completer is a function on the shell:

```go
shell.Completer = func(line []rune, cursor int) readline.Completions {
    return readline.CompleteValues("help", "show", "quit").
        Tag("commands").
        Usage("show <host>")
}
```

The example builds a small fake CLI and shows the main pieces of the API:

- `CompleteValues` / `CompleteValuesDescribed` / `CompleteMessage` — build a
  container of candidates, with or without descriptions, or with a message
  instead of candidates.
- `Tag` groups candidates, so that they can be displayed differently.
- `DisplayList` renders a group as a list rather than as a grid;
  `JustifyDescriptions` aligns the descriptions with the configured
  `completion-list-separator`.
- `Style` colors the values that are inserted.
- `Usage` prints a hint below the input line when there is nothing to complete.

A background goroutine grows the host list and calls `shell.RefreshCompletions()`
to rebuild the open menu in place — any selected candidate is preserved.

Try `show ` + Tab, then wait a few seconds with the menu open to watch it
refresh.

## history

```go
source, err := readline.NewHistoryFromFile("/path/to/history")
shell.History.Add("home", source)
shell.History.AddFromFile("project", ".project_history")
shell.History.Add("temporary", readline.NewInMemoryHistory())
```

Adding the first source replaces the default in-memory history. Any number of
sources can then coexist; `history-source-next` / `history-source-prev` cycle
through them (the example binds them to Ctrl-X n / Ctrl-X p), and
`shell.History.Delete(name)` removes one.

The example falls back to the in-memory source when the history file cannot be
created, so it also runs in read-only environments.

## keymaps

Commands are `func()` values registered by name; binds map key sequences to
those names, exactly like an `.inputrc` file would:

```go
shell.Keymap.Register(map[string]func(){
    "insert-date": func() {
        shell.Cursor().InsertAt([]rune(time.Now().Format(time.RFC3339))...)
    },
})

shell.Config.Bind("emacs", inputrc.Unescape(`\C-t`), "insert-date", false)
```

The example also shows how to switch editing modes at runtime with
`Keymap.SetMain` + `Keymap.UpdateCursor`, and how a command reads the line,
cursor and hint area. Available keys:

| Key | Command |
|-----|---------|
| `Ctrl-T` | insert the current date and time at the cursor |
| `Ctrl-X Ctrl-U` | uppercase the whole line |
| `Ctrl-G` | toggle between Emacs and Vi insert mode |
| `Ctrl-X Ctrl-S` | describe the line in the hint area |

## highlighting

```go
shell.SyntaxHighlighter = func(line []rune) string {
    return color.FgYellow + string(line) + color.Reset
}
```

The callback returns the string to render, with any escape sequences; readline
keeps its own overlays (visual selection, matching parens) on top of it.

The example highlighter colors the command word (green when known, red
otherwise), quoted strings in yellow, and `#` comments in dim gray.

## multiline

```go
shell.AcceptMultiline = func(line []rune) (accept bool) {
    _, err := shellquote.Split(string(line))
    return err == nil
}
```

Returning `false` makes the shell append a newline and keep reading, using the
secondary prompt for continuation lines. The example accepts the buffer only
when its quotes are balanced and it does not end with a backslash.

## prompts

```go
shell.Prompt.Primary(func() string { return "prompts> " })
shell.Prompt.Secondary(func() string { return "... " })
shell.Prompt.Right(func() string { return time.Now().Format("15:04:05") })
shell.Prompt.Transient(func() string { return "prompts (done)" })
shell.Prompt.Tooltip(func(word string) string { return "word: " + word })
```

The transient prompt replaces the primary prompt once the line is accepted, and
is enabled with the `transient-prompt` option (the example turns it on). The
tooltip is recomputed for the word under the cursor on every keypress.

Two logging methods are demonstrated:

- `shell.Printf(...)` prints below the current line and redraws the prompt.
- `shell.PrintTransientf(...)` clears the current line, prints in its place, and
  redraws the prompt/line below.

Use the `log`, `transient` and `ticker` commands to try them; `ticker` starts a
background goroutine printing notifications every two seconds.

## suggestions

```go
shell.SetInlineSuggestion("deploy")
```

The suggestion is rendered as dim text after the cursor when the current line
is a prefix of it, and `forward-char` (Ctrl-F or the right arrow) accepts it.

The example recomputes the suggestion by wrapping the builtin `self-insert`
and `backward-delete-char` commands — a key event is just a command of your
own — and binds Ctrl-G to accept the whole suggestion.

## paste

```go
shell.PasteTransformer = func(text string) string {
    // multi-line payloads are replaced by a placeholder
    return "[Pasted text #1 +6 lines]"
}
```

The hook receives the paste payload after line-ending normalization (`\n`) and
returns the text that is actually inserted. Bracketed paste is opt-in: the
example enables the `enable-bracketed-paste` option so that the terminal wraps
pastes in the first place. The example keeps the original
payload in a store and expands the placeholder when the line is accepted.

## options

```go
_ = shell.Config.Set("autocomplete", true)
_ = shell.Config.Set("completion-ignore-case", true)

autocomplete := shell.Config.GetBool("autocomplete")
```

Options are the same as those of an `.inputrc` file. The example sets the
library-specific options (`autopairs`, `autocomplete`, `completion-ignore-case`,
`usage-hint-always`, `history-autosuggest`, `transient-prompt`,
`completion-list-separator`), optionally loads an extra
`~/.readline-cmd-inputrc` with `inputrc.ParseFile`, and binds a command showing
the live configuration.

`Ctrl-X Ctrl-R` (builtin `re-read-init-file`) reloads the configuration; the
`reload` command does the same through `Keymap.ReloadConfig`.

## macros

```go
shell.Config.Bind("emacs", inputrc.Unescape(`\C-x\C-d`),
    inputrc.EscapeMacro("date +%F"), true)
```

A bind with `macro=true` feeds its action back to the shell instead of running a
command. Builtin recording keys are available in Emacs mode:

| Key | Command |
|-----|---------|
| `Ctrl-X (` | start recording |
| `Ctrl-X )` | stop recording |
| `Ctrl-X e` | replay the last recorded macro |

In Vi mode, `q<register>` records and `@<register>` replays. The example also
registers a command that replays the last macro through `shell.Macros`, bound to
`Ctrl-X Ctrl-L`, and a static macro bound to `Ctrl-X Ctrl-D`.
