> **Runnable example:** [`cmd/paste`](../cmd/paste)

When a terminal is in *bracketed paste* mode, pasted text arrives as a single
chunk wrapped in special escape sequences. Readline reads that chunk, normalises
line breaks (terminals send `\r` or `\r\n` inside a paste, which readline
rewrites to `\n` so they don't corrupt the multiline buffer), and inserts it into
the input line.

Interactive AI and command consoles often want to handle large multiline pastes
differently from normal typing. A common pattern is a compact placeholder — the
editable prompt shows something like `[Pasted text #1 +6 lines]` while the
application keeps the original payload and expands it later when the line is
executed.

### The paste transformer

To support this, readline exposes a hook that rewrites the pasted payload just
before it is inserted:

```go
// PasteTransformer, when set, rewrites bracketed paste payloads before
// they are inserted into the input buffer.
PasteTransformer func(text string) string
```

The callback receives the paste text **after** CRLF/CR normalisation, so it
always sees `\n` line breaks. Its return value is what gets inserted. Returning
an empty string inserts nothing — useful to drop a paste entirely, or to stash
it elsewhere and insert your own placeholder.

The reference store and later expansion are intentionally left to the
application: readline only provides the hook at the point where the terminal
paste has been read and normalised.

### Example

```go
shell := readline.NewShell()

var pasteCount int
pastes := map[string]string{}

// Replace multiline pastes with a compact placeholder, keeping the payload.
shell.PasteTransformer = func(text string) string {
    if strings.Count(text, "\n") == 0 {
        return text // insert short single-line pastes as-is
    }

    pasteCount++
    ref := fmt.Sprintf("[Pasted text #%d +%d lines]", pasteCount, strings.Count(text, "\n"))
    pastes[ref] = text // keep the original payload for expansion at execution time

    return ref
}
```

Set the field to `nil` to disable the hook and restore the default behaviour
(insert the normalised paste directly).
