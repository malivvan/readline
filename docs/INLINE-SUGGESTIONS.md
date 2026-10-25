> **Runnable example:** [`cmd/suggestions`](../cmd/suggestions)

Inline suggestions let an application display its own *ghost text* after the
cursor — AI command predictions, remote completions, or any application-specific
suggestion engine. This is separate from history autosuggestions (which readline
computes from its history sources): the application owns the suggestion text and
readline only renders and accepts it.

### The API

The suggestion is stored on the shell and rendered as dim text after the current
line. Three methods drive it:

```go
// Set the text to display as a suggestion after the cursor.
func (rl *Shell) SetInlineSuggestion(suggestion string)

// Return the current inline suggestion (empty if none).
func (rl *Shell) GetInlineSuggestion() string

// Clear the current inline suggestion.
func (rl *Shell) ClearInlineSuggestion()
```

The suggestion is only rendered when:
- The cursor is at the end of the line, and
- The suggestion extends the current line (the line is a prefix of it).

History autosuggestions keep priority: when `history-autosuggest` is enabled and
a history suggestion is available, it is shown instead of the inline suggestion.

### Setting a suggestion

Applications typically recompute the suggestion whenever the line changes, for
example from a [key-event callback](./PROGRAMMABLE-COMMANDS.md#key-event-callbacks)
or a custom command:

```go
shell := readline.NewShell()

// Somewhere your prediction engine produces a full-line suggestion.
shell.SetInlineSuggestion("git status --short --branch")
```

Pass an empty string (or call `ClearInlineSuggestion`) to remove it.

### Accepting a suggestion

Two commands insert the suggestion into the real input buffer (see the
[Commands](./KEYMAPS-&-COMMANDS.md#history-control) reference):

- `inline-suggest-accept` — accept the whole suggestion.
- `inline-suggest-accept-word` — accept only the next word, leaving the rest as a
  suggestion.

Bind them like any other command, e.g. in your `.inputrc`:

```
"\C-f": inline-suggest-accept
"\ef":  inline-suggest-accept-word
```

In addition, the `forward-char` (Emacs) and `vi-forward-char` (Vim insert)
movement commands fall back to accepting an inline suggestion when the cursor is
at the end of the line and history autosuggest accepted nothing. This means an
existing `forward-char` binding accepts either kind of ghost text, mirroring the
history-autosuggest behaviour.
