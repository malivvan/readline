> **Runnable example:** [`cmd/multiline`](../cmd/multiline)


Readline supports editing buffers spanning on multiple lines.
By default, however, when the user accepts the current input buffer,
the latter will be returned and executed regardless of what it looks like.

Depending on the nature of the application using readline, and the structure
of the buffers it uses, developers can specify a simple callback that will be
used by `readline` to determine whether or not it should:  

- Return the line buffer to the caller for execution.  
- Keep editing the buffer, appending a newline to it first.  

The main objective of this callback is to provide a mechanism that is as
simple as possible to reason of and program for developers using readline.

The callback used by the shell to know what to do upon line acceptance is this:
```go
// This function should return true if the line is deemed complete (thus asking
// the shell to return from its Readline() loop), or false if the shell should
// keep reading input on a newline (thus, insert a newline and read).
AcceptMultiline func(line []rune) (accept bool)
```

### Writing the callback

The most basic example of such a callback might be a bash-like logic, where
if the line ends with a backslash, edition should continue on a new line:
```go
// writing the callback
bashMultiline := func(line []rune) (accept bool) {
    if strings.HasSuffix(string(line), "\\") {
    return false
    }

    return true
}

// and binding it.
shell := readline.NewShell()
shell.AcceptMultiline = bashMultiline
```
There are quite a few other "syntaxes" where the multiline callback might be
almost as simple as this one. One of them is the SQL command syntax, which mandates
a command to end with a semi-colon to be deemed complete and executable.


### Example

Below is a slightly more complicated example still involving sh-like command strings.
Here, the shell will not return the line for execution if the last word in the line is 
an unterminated quoted word (single and/or double quotes)

```go
import "github.com/kballard/go-shellquote"

func acceptMultiline(line []rune) (accept bool) {
    // Errors are either: unterminated quotes, or unterminated escapes.
    _, err := shellquote.Split(string(line))
    if err == nil {
        return true
    }

    switch err {
    case shellquote.UnterminatedDoubleQuoteError, shellquote.UnterminatedSingleQuoteError:
        // Unterminated quotes: keep reading.
        return false
            
    case shellquote.UnterminatedEscapeError:
        // If there no are trailing spaces, keep reading.
        if len(line) > 0 && line[len(line)-1] == '\\' {
            return false
        }
        return true
    }

    return true
}
```
