> **Runnable example:** [`cmd/options`](../cmd/options)


## Introduction

The `.inputrc` file is well established and known from bash users. For those unfamiliar,
the inputrc is a user-wide configuration for specifying input keys behavior for legacy
readline libraries.

The `readline` library offers complete `.inputrc` support, for commands and options.
A few number of options will however have no effect on this library, as some of them
were not needed anymore. Each option/command not implemented will be signaled as such.

For a tour of the original readline/inputrc documentation, see [here](https://man7.org/linux/man-pages/man3/readline.3.html).
This page will help users not yet acquainted with `.inputrc` to setup their own and declare
custom binds, macros and options.

## Setting options in the `.inputrc` file

Below are a few examples of options being set in the init file.

We start with some global settings, applying to all applications 
using readline, regardless of the language it's programmed in:
```bash
# Line & commands
set editing-mode emacs

# Input/Output
set keyseq-timeout 50
set meta-flag on
set input-meta on 
set convert-meta on
set enable-meta-key on
set output-meta on
```

Since this library offers options that are not found in the legacy GNU readline,
those specific options can be set using a `$if` directive, which will apply to
all applications using this Go `readline` library:
```bash
# github.com/malivvan/readline Go readline library
$if Go
    set autocomplete off
    set usage-hint-always on
    set history-autosuggest off
    set autopairs on
    set transient-prompt off
$endif
```

As well and conformingly to the traditional `inputrc` model, each application can
make use of specific options and bindings:
```bash
# The binds/options below only apply when readline is
# used by the MyApplication (case insentive) program.
$if MyApplication

    # Shell options
    set autocomplete on
    set usage-hint-always on
    set autopairs off

    # Keybindings (Emacs)
    set keymap emacs

    "\ez": vi-editing-mode
    "\e[D": backward-char
    "\e[C": forward-char 
    "\M-\C-F": shell-forward-word
    "\M-\C-B": shell-backward-word
    "\M-%": vi-match
    "\M-\C-W": shell-backward-kill-word
    TAB : menu-complete
    "\e[Z": menu-complete-backward

    # Keybindings (Vim command mode)
    set keymap vi-command

    # Changing text
    "\C-t": transpose-chars
    "\M-t": transpose-words
    "\e\C-T": shell-transpose-words
    "\M-s": select-keyword-next
    "\e\C-s": select-keyword-prev

    # Keybindings (Vim insert mode)
    set keymap vi-insert

    TAB: menu-complete 
    "\e[Z": menu-complete-backward
    "\M-'": vi-registers-complete
$endif
```


## Options

The following section describes the options that can be set in the `.inputrc` file.  
Legacy options not implemented have their names followed by '\[NI]'.   
Some options are specific to this library, in which case their name is followed by '\[S]'. 

- `active-region-start-color`    
      A string variable that controls the text color and
      background when displaying the text in the active region
      (see the description of enable-active-region below).  This
      string must not take up any physical character positions
      on the display, so it should consist only of terminal
      escape sequences.  It is output to the terminal before
      displaying the text in the active region.  This variable
      is reset to the default value whenever the terminal type
      changes.  The default value is the string that puts the
      terminal in standout mode, as obtained from the terminal's
      terminfo description.  A sample value might be
      "\e[01;33m".
       
- `active-region-end-color`    
      A string variable that "undoes" the effects of
      active-region-start-color and restores "normal" terminal
      display appearance after displaying text in the active
      region.  This string must not take up any physical
      character positions on the display, so it should consist
      only of terminal escape sequences.  It is output to the
      terminal after displaying the text in the active region.
      This variable is reset to the default value whenever the
      terminal type changes.  The default value is the string
      that restores the terminal from standout mode, as obtained
      from the terminal's terminfo description.  A sample value
      might be "\e[0m".

- `autocomplete` (Off)    
      Asynchronously and continuously display possible completions
      for the current word/cursor position, using the custom tab
      completion callback `TabCompleter`. Autocomplete is only
      triggered when at least one character is in the buffer.

- `autopairs` (Off)    
      If set to On, the shell will use builtin syntax completion
      for all "pair characters" (brackets/bracesparens/quotes):
      When inserting the opening character, the pair will be 
      automatically inserted. When inserting the closing one,
      and if the next character after the cursor is identical,
      the cursor will jump over it instead of inserting a new one.
      When deleting characters (with `backward-delete-char` or Vim
      equivalent commands), if the character to delete is the
      opening character of a pair and that its closer immediately
      follows it, both characters will be removed from the line.

- `bell-style` (audible) '\[NI]'    
      Controls what happens when readline wants to ring the
      terminal bell.  If set to none, readline never rings the
      bell.  If set to visible, readline uses a visible bell if
      one is available.  If set to audible, readline attempts to
      ring the terminal's bell.

- `bind-tty-special-chars` (On)     
      If set to On (the default), readline attempts to bind the
      control characters   treated specially by the kernel's
      terminal driver to their readline equivalents.

- `blink-matching-paren` (Off)    
      If set to On, readline attempts to briefly move the cursor
      to an opening parenthesis when a closing parenthesis is
      inserted.

- `colored-completion-prefix` (Off)    
      If set to On, when listing completions, readline displays
      the common prefix of the set of possible completions using
      a different color.  The color definitions are taken from
      the value of the LS_COLORS environment variable.  If there
      is a color definition in $LS_COLORS for the custom suffix
      "readline-colored-completion-prefix", readline uses this
      color for the common prefix instead of its default.

- `colored-stats` (Off) '\[NI]'    
      If set to On, readline displays possible completions using
      different colors to indicate their file type.  The color
      definitions are taken from the value of the LS_COLORS
      environment variable.

- `comment-begin` (#)    
      The string that is inserted in vi mode when the
      insert-comment command is executed.  This command is bound
      to M-# in emacs mode and to # in vi command mode.

- `completion-description-style` (\e[1;30m)    
      The highlighting/color style to use for completion descriptions.
      The value should only include color/style escaped characters.
      The default description color is a light dark gray color.

- `completion-display-width` (-1)    
      The number of screen columns used to display possible
      matches when performing completion.  The value is ignored
      if it is less than 0 or greater than the terminal screen
      width.  A value of 0 will cause matches to be displayed
      one per line.  The default value is -1.

- `completion-ignore-case` (Off)    
      If set to On, readline performs filename matching and
      completion in a case-insensitive fashion.

- `completion-list-separator` (--)     
    The string to use as a separation between completion 
    candidates and their descriptions, if they have any.

- `completion-map-case` (Off)    
      If set to On, and completion-ignore-case is enabled,
      readline treats hyphens (-) and underscores (_) as
      equivalent when performing case-insensitive filename
      matching and completion.

- `completion-prefix-display-length` (0)  
      The length in characters of the common prefix of a list of
      possible completions that is displayed without
      modification.  When set to a value greater than zero,
      common prefixes longer than this value are replaced with
      an ellipsis when displaying possible completions.

- `completion-query-items` (100)    
      This determines when the user is queried about viewing the
      number of possible completions generated by the
      possible-completions command.  It may be set to any
      integer value greater than or equal to zero.  If the
      number of possible completions is greater than or equal to
      the value of this variable, readline will ask whether or
      not the user wishes to view them; otherwise they are
      simply listed on the terminal.  A negative value causes
      readline to never ask.

- `completion-selection-style` (\e[48;05;255m)    
      The highlighting/color style to use for the currently selected
      completion candidate in completion menu-select mode.
      The value should only include color/style escaped characters.
      The default value is a white background preserving fg colors.

- `convert-meta` (On)    
      If set to On, readline will convert characters with the
      eighth bit set to an ASCII key sequence by stripping the
      eighth bit and prefixing it with an escape character (in
      effect, using escape as the meta prefix).  The default is
      On, but readline will set it to Off if the locale contains
      eight-bit characters.  This variable is dependent on the
      LC_CTYPE locale category, and may change if the locale is
      changed.

- `cursor-<keymap-name>` (default)    
      This option is a pattern option, meaning that you can specify
      a different "option name" for each keymap. Thus by default,
      users can specify `cursor-vi-insert`, `cursor-vi-cmd`, `cursor-vi-opp`,
      `cursor-emacs`, etc, for each of the available keymaps.
      Should the program using readline making use of custom keymaps,
      this option works the same for those. Valid values for this
      option are `block`, `beam`, `underline`, `blinking-block`, `blinking-beam`,
      `blinking-underline` or `default`.

- `cursor-position-probe` (On)    
      If set to On (the default), the display engine queries the terminal
      for the cursor position (the `ESC[6n` Device Status Report) when it
      needs to anchor its layout. This is what lets the shell correctly
      reserve space and avoid overlapping the prompt at the bottom of the
      window. Turn it Off for environments that do not reliably answer the
      query (some minimal emulators, constrained CI, certain PTY test
      harnesses): the engine then falls back to the printed prompt width,
      which is exact when the input line starts at column 0. The documented
      cost of disabling it is that the shell can no longer detect the bottom
      of the window, so a prompt at the very bottom may overlap.

- `disable-completion` (Off)    
      If set to On, readline will inhibit word completion.
      Completion characters will be inserted into the line as if
      they had been mapped to self-insert.

- `echo-control-characters` (On)    
      When set to On, on operating systems that indicate they
      support it, readline echoes a character corresponding to a
      signal generated from the keyboard.

- `editing-mode` (emacs)    
      Controls whether readline begins with a set of key
      bindings similar to Emacs or vi.  editing-mode can be set
      to either emacs or vi.

- `emacs-mode-string` (@)    
      If the show-mode-in-prompt variable is enabled, this
      string is displayed immediately before the last line of
      the primary prompt when emacs editing mode is active.  The
      value is expanded like a key binding, so the standard set
      of meta- and control prefixes and backslash escape
      sequences is available.  Use the \1 and \2 escapes to
      begin and end sequences of non-printing characters, which
      can be used to embed a terminal control sequence into the
      mode string.

- `enable-active-region` (On)    
      The point is the current cursor position, and mark refers
      to a saved cursor position.  The text between the point
      and mark is referred to as the region.  When this variable
      is set to On, readline allows certain commands to
      designate the region as active.  When the region is
      active, readline highlights the text in the region using
      the value of the active-region-start-color, which defaults
      to the string that enables the terminal's standout mode.
      The active region shows the text inserted by bracketed-
      paste and any matching text found by incremental and non-
      incremental history searches.

- `enable-bracketed-paste` (On) '\[NI]'    
      When set to On, readline configures the terminal to insert
      each paste into the editing buffer as a single string of
      characters, instead of treating each character as if it
      had been read from the keyboard.  This prevents readline
      from executing any editing commands bound to key sequences
      appearing in the pasted text.

- `enable-keypad` (Off) '\[NI]'    
      When set to On, readline will try to enable the
      application keypad when it is called.  Some systems need
      this to enable the arrow keys.

- `enable-meta-key` (On) '\[NI]'    
      When set to On, readline will try to enable any meta
      modifier key the terminal claims to support when it is
      called.  On many terminals, the meta key is used to send
      eight-bit characters.

- `expand-tilde` (Off) '\[NI]'    
      If set to On, tilde expansion is performed when readline
      attempts word completion.

- `history-autosuggest` (Off)    
      If set to On, the shell will display the first history line
      matching the current buffer (searching backward) with a dim
      color. When such line is available, commands such as `forward-char`
      and `forward-word` (or their Vim equivalents) will have special
      behavior. See the commands documentation for details.
      This option indeed provides fish-like autosuggestions.

- `history-preserve-point` (Off)    
      If set to On, the history code attempts to place point at
      the same location on each history line retrieved with
      previous-history or next-history.

- `history-size` (unset)    
      Set the maximum number of history entries saved in the
      history list.  If set to zero, any existing history
      entries are deleted and no new entries are saved.  If set
      to a value less than zero, the number of history entries
      is not limited.  By default, the number of history entries
      is not limited.  If an attempt is made to set history-size
      to a non-numeric value, the maximum number of history
      entries will be set to 500.

- `horizontal-scroll-mode` (Off) '\[NI]'    
      When set to On, makes readline use a single line for
      display, scrolling the input horizontally on a single
      screen line when it becomes longer than the screen width
      rather than wrapping to a new line.  This setting is
      automatically enabled for terminals of height 1.

- `input-meta` (Off)    
      If set to On, readline will enable eight-bit input (that
      is, it will not clear the eighth bit in the characters it
      reads), regardless of what the terminal claims it can
      support.  The name meta-flag is a synonym for this
      variable.  The default is Off, but readline will set it to
      On if the locale contains eight-bit characters.  This
      variable is dependent on the LC_CTYPE locale category, and
      may change if the locale is changed.

- `isearch-terminators` (``C-[ C-J'')    
      The string of characters that should terminate an
      incremental search without subsequently executing the
      character as a command.  If this variable has not been
      given a value, the characters ESC and C-J will terminate
      an incremental search.

- `keymap` (emacs)    
      Set the current readline keymap.  The set of legal keymap
      names is emacs, emacs-standard, emacs-meta, emacs-ctlx,
      vi, vi-move, vi-command, and vi-insert.  vi is equivalent
      to vi-command; emacs is equivalent to emacs-standard.  The
      default value is emacs.  The value of editing-mode also
      affects the default keymap.

- `keyseq-timeout` (500) '\[NI]'    
      Specifies the duration readline will wait for a character
      when reading an ambiguous key sequence (one that can form
      a complete key sequence using the input read so far, or
      can take additional input to complete a longer key
      sequence).  If no input is received within the timeout,
      readline will use the shorter but complete key sequence.
      The value is specified in milliseconds, so a value of 1000
      means that readline will wait one second for additional
      input.  If this variable is set to a value less than or
      equal to zero, or to a non-numeric value, readline will
      wait until another key is pressed to decide which key
      sequence to complete.

- `mark-directories` (On) '\[NI]'    
      If set to On, completed directory names have a slash
      appended.

- `mark-modified-lines` (Off) '\[NI]'    
      If set to On, history lines that have been modified are
      displayed with a preceding asterisk (*).

- `mark-symlinked-directories` (Off) '\[NI]'    
      If set to On, completed names which are symbolic links to
      directories have a slash appended (subject to the value of
      mark-directories).

- `match-hidden-files` (On) '\[NI]'    
      This variable, when set to On, causes readline to match
      files whose names begin with a \`.\` (hidden files) when
      performing filename completion.  If set to Off, the
      leading `.' must be supplied by the user in the filename
      to be completed.

- `menu-complete-display-prefix` (Off)    
      If set to On, menu completion displays the common prefix
      of the list of possible completions (which may be empty)
      before cycling through the list.

- `multiline-column` (On)    
      If set to ON, when editing a buffer spaning multiple lines,
      the left-most column of the UI will start displaying column
      signs along with the secondary prompt (a custom or default)

- `multiline-column-numbered` (Off)    
      If set to ON, the left-most column of the UI will print 
      dimmed line numbers if the input buffer spans multiple lines.

- `multiline-column-custom` ("")    
      If this value, specified as a string, is not nil/empty,
      if the input buffer spans multiple lines, the column signs
      of the left-most column will use this display string rather
      than the default graphic columns.

- `output-meta` (Off)    
      If set to On, readline will display characters with the
      eighth bit set directly rather than as a meta-prefixed
      escape sequence.  The default is Off, but readline will
      set it to On if the locale contains eight-bit characters.
      This variable is dependent on the LC_CTYPE locale
      category, and may change if the locale is changed.

- `page-completions` (On) '\[NI]'    
      If set to On, readline uses an internal more-like pager to
      display a screenful of possible completions at a time.

- `print-completions-horizontally` (Off) '\[NI]'    
      If set to On, readline will display completions with
      matches sorted horizontally in alphabetical order, rather
      than down the screen.

- `transient-prompt` (Off)    
      If set to On and if the shell has a registered transient
      prompt function (`shell.Prompt.Transient(func() string)`),
      each readline call loop will redisplay the prompt
      transiently (that is, delete the current one, go above it, 
      print the transient prompt and accepted line, and go below
      before redisplaying the new prompt buffer).

- `revert-all-at-newline` (Off)    
      If set to On, readline will undo all changes to history
      lines before returning when accept-line is executed.  By
      default, history lines may be modified and retain
      individual undo lists across calls to readline.

- `show-all-if-ambiguous` (Off) '\[NI]'    
      This alters the default behavior of the completion
      functions.  If set to On, words which have more than one
      possible completion cause the matches to be listed
      immediately instead of ringing the bell.

- `show-all-if-unmodified` (Off) '\[NI]'    
      This alters the default behavior of the completion
      functions in a fashion similar to show-all-if-ambiguous.
      If set to On, words which have more than one possible
      completion without any possible partial completion (the
      possible completions don't share a common prefix) cause
      the matches to be listed immediately instead of ringing
      the bell.

- `show-mode-in-prompt` (Off)    
      If set to On, add a string to the beginning of the prompt
      indicating the editing mode: emacs, vi command, or vi
      insertion.  The mode strings are user-settable (e.g.,
      emacs-mode-string).

- `skip-completed-text` (Off)    
      If set to On, this alters the default completion behavior
      when inserting a single match into the line.  It's only
      active when performing completion in the middle of a word.
      If enabled, readline does not insert characters from the
      completion that match characters after point in the word
      being completed, so portions of the word following the
      cursor are not duplicated.

- `usage-hint-always` (Off)    
      If set to On, if the completions available at any given
      time include one more usage message hints, those will be
      displayed regardless of the number of completions.
      When set to Off, usage messages are only showed when there
      are no available completions in the given context.

- `vi-cmd-mode-string` ((cmd))    
      If the show-mode-in-prompt variable is enabled, this
      string is displayed immediately before the last line of
      the primary prompt when vi editing mode is active and in
      command mode.  The value is expanded like a key binding,
      so the standard set of meta- and control prefixes and
      backslash escape sequences is available.  Use the \1 and
      \2 escapes to begin and end sequences of non-printing
      characters, which can be used to embed a terminal control
      sequence into the mode string.

- `vi-ins-mode-string` ((ins))    
      If the show-mode-in-prompt variable is enabled, this
      string is displayed immediately before the last line of
      the primary prompt when vi editing mode is active and in
      insertion mode.  The value is expanded like a key binding,
      so the standard set of meta- and control prefixes and
      backslash escape sequences is available.  Use the \1 and
      \2 escapes to begin and end sequences of non-printing
      characters, which can be used to embed a terminal control
      sequence into the mode string.

- `visible-stats` (Off) '\[NI]'    
      If set to On, a character denoting a file's type as
      reported by stat(2) is appended to the filename when
      listing possible completions.

