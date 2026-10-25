> **Runnable example:** [`cmd/options`](../cmd/options)



### Loading configurations

At startup, any program using readline will read the user `~/.inputrc` file, or if not found
the system one, `/etc/inputrc`. It is however possible for developers to access the loaded
configuration from the shell API, and to modify it freely (even replacing it altogether,
although this is neither really practical nor recommended).

To access the configuration:
```go
shell := readline.NewShell()
config := shell.Config
```
Developers will want to use readline for an application, which will have a name.
The library thus enables them to pass this name when creating the shell instance,
so that any app-specific binds/options will be parsed/setup accordingly:
```go
appName := "teleport"

// Creating the shell with this application name: $if blocks matching
// "teleport" (case-insensitively) will be applied.
shell := readline.NewShell(inputrc.WithApp(appName))

// Or adding it later and reloading the configuration (not recommended, as this
// re-reads the whole init file).
shell.Opts = append(shell.Opts, inputrc.WithApp(appName))
_ = shell.Keymap.ReloadConfig(shell.Opts...)
```

### Global options

Inputrc options are used to setup the global behavior, display and input/output settings
of the shell. They can be accessed, queried and set like the following:
```go
// Access the map of all options
allOptions := config.Vars

// Querying option values
editingMode := config.GetString("editing-mode")         // Either "vi" or "emacs"
autocomplete := config.GetBool("autocomplete")          // directly as boolean (on/off)
viPromptMode := config.GetString("vi-ins-mode-string")  // Some options are string values
keyTimeout := config.GetInt("keyseq-timeout")           // Others are integers

// Modifying the option values
config.Set("editing-mode", "vi")            
config.Set("autocomplete", true)
config.Set("keyseq-timeout", 100)
```

### Key bindings 

The second part of the `.inputrc` file is used to declare custom keybindings, either for commands or macros.
As well, it's possible to add, delete or overwrite all or some keybindings, per keymap, through the code.

To access each keymap and its binds:
```go
// Get all keymaps and their binds
binds := config.Binds

// Emacs
emacs := binds["emacs"]              // Access all binds for the Emacs keymap
emacsCtrlX := binds["emacs-ctrlx"]   // All binds of the Emacs Ctrl-X (subset of emacs)
emacsMeta := binds["emacs-meta"]     // All binds of the Emacs Ctrl-Meta (subset of emacs)

// Vim
viCommand := binds["vi"]             // You can also query vi/vi-move/vi-main for this keymap
viInsert := binds["vi-insert"]       // Insertion keymap
viOpp := binds["vi-opp"]             // The vi-operator pending keymap is a special one.

// Completion & incremental search
menuSelect := binds["menu-select"]  // The keymap used when selecting completions.
isearch := binds["isearch"]         // Incremental-search (empty by default, see doc for why)
```

To override, add or delete binds in a given keymap, you can proceed in several ways:
```go
// Using the provided configuration method to add a bind. Note that key
// sequences are stored unescaped, exactly as the shell matches incoming keys.
config.Bind("vi", "K", "vi-kill-eol", false)                              // Bind a command
config.Bind("vi", inputrc.Unescape(`\C-K`), "K", true)                     // Bind a macro (doing what the line above does)

// Or adding directly to the map
viCommand["K"] = inputrc.Bind{Action: "vi-kill-eol", Macro: false}       // Bind a command.
viCommand[inputrc.Unescape(`\C-K`)] = inputrc.Bind{Action: "K", Macro: true} // Bind a macro.

// Deleting is as simple
delete(viCommand, "K")
delete(viCommand, `\C-K`)
```
