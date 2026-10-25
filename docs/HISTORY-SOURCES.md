> **Runnable example:** [`cmd/history`](../cmd/history)


Each newly created readline instance will use an in-memory  history source, which
does not save accepted lines on the filesystem, and is wiped when the application exits.

Additionally, the instance supports an arbitrary number of different history sources
to be bound and used: see below for the list of widgets available for this.

A list of history-related commands and their descriptions is also available [here](./KEYMAPS-&-COMMANDS.md#history-control).


## File-based command history

The readline library offers 2 ways to create a history source from a file, shown below.
Each of those calls will replace the default in-memory history source.
The name of the history is used in the completions and incremental search hints.
```go
// Either create the history separately from the shell...
history, err := readline.NewHistoryFromFile("/path/to/.app_history")

// ... and bind it in another call. 
shell.History.Add("history name", history)

// Or create and bind it in one call.
shell.History.AddFromFile("history name", "/path/to/.app_history")
```


## Adding and removing histories

You can remove one or more history sources with the following calls:
```go
shell.History.Delete()                                     // Delete all sources
shell.History.Delete("local history", "another history")   // Or some of them
```
The shell offers commands to cycle through history sources if more than one is available.


## Implementing history sources

You can also implement the `readline.History` interface. It is an alias of
`history.Source`, so both names can be used interchangeably:
```go
type History interface {
	// Write appends the line and returns the updated number of lines, or an error
	Write(string) (int, error)

	// GetLine takes the historic line number and returns the line or an error
	GetLine(int) (string, error)

	// Len returns the number of history lines
	Len() int

	// Dump returns everything in readline. The return is an interface{} because
	// not all LineHistory implementations will want to structure the history in
	// the same way. And since Dump() is not actually used by the readline API
	// internally, this methods return can be structured in whichever way is most
	// convenient for your own applications (or even just create an empty
	//function which returns `nil` if you don't require Dump() either)
	Dump() interface{}
}
```

The following code shows how to write a history satisfying the `History` interface, 
sending history to a gRPC server:

```go
// ClientHistory - Writes and queries only the Client's history
type ClientHistory struct {
	LinesSinceStart int // Keeps count of line since session
	items           []string
}

// Write - Sends the last command to the server for saving
func (h *ClientHistory) Write(s string) (int, error) {

	res, err := transport.RPC.AddToHistory(context.Background(),
		&clientpb.AddCmdHistoryRequest{Line: s})
	if err != nil {
		return 0, err
	}

	// The server sent us back the whole user history,
	// so we give it to the user history (the latter never
	// actually uses its Write() method.
	UserHist.cache = res.Lines

	h.items = append(h.items, s)
	return len(h.items), nil
}

// GetLine returns a line from history
func (h *ClientHistory) GetLine(i int) (string, error) {
	if len(h.items) == 0 {
		return "", nil
	}
	return h.items[i], nil
}

// Len returns the number of lines in history
func (h *ClientHistory) Len() int {
	return len(h.items)
}

// Dump returns the entire history
func (h *ClientHistory) Dump() interface{} {
	return h.items
}
```
