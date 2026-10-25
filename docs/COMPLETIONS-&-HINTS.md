> **Runnable example:** [`cmd/completions`](../cmd/completions)


The readline library offers an advanced -yet quite simple to use- completion API.
As said before, its behavior and features are mostly similar to the Z-Shell completion 
system, while offering a much narrower API to attain those features.

The latter is heavily inspired (if not outright copied) from the [carapace completion API](https://github.com/rsteube/carapace),
but trimmed down to its most essential methods.

Related documentation resources:
- The file containing the completions API is [here](../completion_values.go).
- A runnable demonstration lives in [`cmd/completions`](../cmd/completions).
- The list of available completion commands is [here](./KEYMAPS-&-COMMANDS.md#completion).

# Table of Contents

<details>
<summary>Click to expand Table of Contents</summary>
<ul>
        <li><a href="#table-of-contents">Table of Contents</a></li>
        <li><a href="#the-completions-container">The <code>Completions</code> container</a></li>
        <li><a href="#completion-candidates">Completion Candidates</a></li>
        <li><a href="#autosuffix-removal">Autosuffix removal</a></li>
        <li><a href="#hints-and-usage-strings">Hints and usage strings</a></li>
        <li><a href="#the-hint-area-and-its-lanes">The hint area and its lanes</a></li>
        <li><a href="#asynchronous-updates">Asynchronous updates</a></li>
    </ul></li>
</ul>
</details>

## The `Completions` container
The readline shell uses the following signature as a valid completer function
(the `line` parameter is the whole input line, and `cursor` is the cursor position index):
```go
Completer func(line []rune, cursor int) Completions
```

The `Completions` type will store all completion candidates and all meta-parameters,
which are various settings that are used when either building completions, or managing
how they are inserted, trimmed, etc.

Here is a snippet of its internals, for the sake of detail:
```go
type Completions struct {
    values   rawValues          // All completion candidates
    messages messages           // All messages to be printed as hints below the line, above completions.
    noSpace  suffixMatcher      // Stores a list of runes against which to perform autosuffix-removal
    usage    string             // A single usage message (identical to messages in its working).
    listLong map[string]bool    // A list of tags for which to force listing comps below each other.
    noSort   map[string]bool    // A list of tags for which we should not sort the completions.

    // Initially this will be set to the part of the current word
    // from the beginning of the word up to the position of the cursor;
    // it may be altered to give a common prefix for all matches.
    PREFIX string

	// Initially this will be set to the part of the current word,
	// starting from the cursor position up to the end of the word.
	// It may be altered so that inserted completions don't overwrite
	// entirely any suffix when completing in the middle of a word.
	SUFFIX string
}
```


## Completion Candidates
In this section we explain the various components of a single completion candidate, 
and the effect they have on the completion system. This is the code:
```go
type Completion struct {
    Value       string // Value is the value of the completion as actually inserted in the line
    Display     string // When display is not nil, this string is used to display the completion in the menu.
    Description string // A description to display next to the completion candidate.
    Style       string // An arbitrary string of color/text effects to use when displaying the completion, in SGR notation.
    Tag         string // All completions with the same tag are grouped together and displayed under the tag heading.
}
```

Notes additional to code comments:
- `Value`: In most cases, you don't need to take care of adding the current word prefix to them. 
  Readline takes care of it. See the next section for details.
- `Display`: When the display is empty, readline automatically uses the `Value` for it.
- `Style`: Be sure to pass style strings in [ANSI-compliant](https://en.wikipedia.org/wiki/ANSI_escape_code) notation **without escape sequences**.
  Additionally, note that there is full support for colors in display and description strings.
  The readline shell will automatically handle the correct size/display computations and related
  problems, even for strings with arbitrary use of escaped color sequences.
  Thus, the `Style` field is simply a way to more formally/simply specify styling for a given candidate.
- `Tag`: Other than grouping results together under a heading, the tag can also be used for
  specific listing/sorting behaviors, as seen in the previous section.


## Autosuffix removal
As seen in the `Completions` container type above, there is a `suffixMatcher` type.
This is simply a list of runes, which can populated with the following function:
```go
// NoSpace disables space suffix for given characters (or all if none are given).
// These suffixes will be used for all completions that have not specified their
// own suffix-matching patterns.
// This is used for slash-autoremoval in path completions, comma-separated completions, etc.
func (c Completions) NoSpace(suffixes ...rune) Completions {
	if len(suffixes) == 0 {
		c.noSpace.Add('*')
	}
	c.noSpace.Add(suffixes...)
	return c
}
```
As its documentation implies, the list of runes that it stores are compared against
the last rune of a to-be-inserted completion candidate, and removes this last rune
when the next entered key conforms to specific requirements.

For instance, in general, if a completion ending with a space is to be inserted,
and that the next entered key is a space, the net result is 0: one space pruned,
and one space inserted.

This behavior is especially noticable for directory paths, where you can either
"abandon" the current path completion with a space (removing a slash if the candidate
was a directory), or typing a slash because you're good with the current candidate,
and you want to complete what is inside it.

Lastly, the [console application] library makes heavy use of this for completing
flags, interspersed or not, and completing lists of arguments when the flags are
arrays or maps.


## Hints and usage strings

The readline shell supports printing various hints below the current input line,
just above any completions. The `Completions` container also allows to provide
such hints for a given completion context.

It allows you to return generally two types/purposes of messages:
- Usage messages, set to indicate what is expected (when no completions are given)
- Error messages, when you attempted to provide completions but that you could not 
  for some reason. Note the `Suppress()` method below to ignore some errors.

Those hints above can be provided by two different functions:
```go
func CompleteMessage(msg string, args ...any) Completions
func (c Completions) Usage(usage string, args ...any)
```

An example of using each:
```go
// Message is generally used to notify something went wrong.
err := getCompsFromSomwehere
if err != nil {
    comps := readline.CompleteMessage("Failed to fetch a list of completions")
}

// Usage is more used when you know that you don't have completions
// to give, but you still want to give the user a hint on what is wanted.
comps := comps.Usage("An IPv4 address used as positional argument")
```

Both functions -and the messages they return- will actually have the same behavior:
printing those messages below the current input line. Note that these strings can
contain arbitrary escaped colors/text effects: the shell will both render them and
correctly take them into account when padding/displaying.


## The hint area and its lanes

The usage/error messages above are returned by the completer. In addition, the shell
exposes the hint area directly through `rl.Hint`, as a set of **independent lanes** that
are rendered top to bottom. Keeping them separate lets passive hinting, asynchronous
status reporting and completion hints coexist without overwriting one another:

| Lane | Set with | Purpose & lifetime |
|------|----------|--------------------|
| persistent | `rl.Hint.Persist` / `ResetPersist` | sticky status (e.g. Vim register/iteration); set by commands |
| provided   | `rl.Hint.SetProvider` | passive hint, recomputed from the current line on every refresh |
| transient  | `rl.Hint.SetTransient` / `ClearTransient` | one-shot/async status; **not** cleared by completion or isearch |
| completion | the completer's `Usage`/`Message`, `rl.Hint.Set` / `SetTemporary` | completion & command hints |

**Passive provider hint** — recomputed on every redisplay, so it tracks the input as it
changes. It renders above the transient and completion lanes. Return `nil`/empty for no hint:
```go
rl.Hint.SetProvider(func(line []rune, cursor int) []rune {
    return []rune("hint computed from: " + string(line))
})
```

**Transient (status) hint** — unlike the completion lane, it is not cleared by completion
or incremental-search activity; it persists until replaced or cleared. It is the right
channel for explicit status messages, typically pushed asynchronously (see below). It is
safe to call from any goroutine:
```go
rl.Hint.SetTransient("\x1b[33m⣾ working…\x1b[0m")
rl.Hint.ClearTransient()
```


## Asynchronous updates

The `Readline()` loop blocks while waiting for a keypress. To let a background goroutine
update the interface while the shell is idle, the shell provides a wake primitive:
```go
rl.Keys.RequestRefresh() // safe from any goroutine; repaints an idle shell
```
You rarely need to call it directly: the asynchronous hint and completion APIs call it for
you. In all cases the actual rendering happens on the `Readline()` goroutine (the wake only
unblocks it), so background producers never write to the display themselves — they only
mutate their own data and call one of the methods below.

**Async status in the hint lane** — set a transient hint from a goroutine; the idle shell
repaints on its own, with no keystroke:
```go
go func() {
    for _, stage := range stages {
        time.Sleep(step)
        rl.Hint.SetTransient(stage)
    }
    rl.Hint.ClearTransient()
}()
```

**Async completions** — when a completion menu is open and a background producer has new
results, ask the shell to regenerate the menu in place:
```go
func (rl *Shell) RefreshCompletions()
```
- Safe from any goroutine; a clean no-op when no menu is active.
- Re-runs the completer and rebuilds the open menu, repainting without a keystroke.
- If the user has a candidate selected, the selection is **preserved**: the same candidate
  (matched by tag+value) is re-selected in the rebuilt menu. If it no longer exists in the
  new results, the menu stays open with no selection.

Typical pattern — a producer grows a cache the completer reads, then calls `RefreshCompletions`:
```go
rl.Completer = func(line []rune, cursor int) readline.Completions {
    return readline.CompleteValues(cache.snapshot()...) // reads the shared cache
}

go func() {
    for host := range discovered {
        cache.add(host)
        rl.RefreshCompletions() // menu fills in live; any selection is kept
    }
}()
```



