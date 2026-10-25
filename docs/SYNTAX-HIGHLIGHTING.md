> **Runnable example:** [`cmd/highlighting`](../cmd/highlighting)


Syntax highlighting support in readline is quite simple: 
you only have to return a colored string, however you want it to be.
Readline will deal with it as is, and will seamlessly add some highlighting
on top of it for things like visual selection, blinking matching parens, etc.

### Builtin highlighting

Out of the box, readline applies highlighting to the following things:
- Line comments.
- Visual selection (Vim visual mode, or Emacs active region).
- Some special visual selections for commands involving surround characters.
- Blinking matching parens (when enabled), which are shown when the cursor is
  on a closing/opening matching character (brackets/braces/parenthesis).

### Writing the callback

Additionally, developers using readline can write and bind a custom callback which
will apply highlighting to the input buffer, before readline adds its own overlays
when it needs to. The callback is the following:
```go
// SyntaxHighlighter is a helper function to provide syntax highlighting.
// Once enabled, set to nil to disable again.
SyntaxHighlighter func(line []rune) string
```

The following is an example callback which highlights all quoted words in the buffer:
```go
// highlighter applies highlighting to quoted words.
func highlighter(line []rune) string {
    var highlighted string

	openQuote = "\x1b[33m"  // yellow
	closeQuote = "\x1b[39m" // reset

    quoted := false

    for _, char := range line {

        // Start or close highlighting for quoted words.
        if char == '\'' || char == '"' {
            if quoted {
                highlighted += string(char)
                highlighted += closeQuote 
            } else {
                highlighted += openQuote
                highlighted += string(char)
            }
                    
            quoted = !quoted
            continue
        }

        // Or simply add the character
        highlighted += char
    }

    return highlighted
}

// Binding the highlighter
shell := readline.NewShell()
shell.SyntaxHighlighter = highlighter
```

Another example, even simpler, might be to use the famous `chroma` library:
```go
import "github.com/alecthomas/chroma/quick"

// highlighter uses chroma to highlight the line in an SQL dialect.
func highlighter(line []rune) string {
    var highlighted strings.Builder

    err := quick.Highlight(&highlighter, string(input), "sql", "terminal16m", "xcode-dark")
    if err != nil {
        return string(line)
    }

    return highlighted.String()
}

```
