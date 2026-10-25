> **Runnable example:** [`cmd/keymaps`](../cmd/keymaps)


This page lists the various commands for the readline library, and gives explanations for its keymap model.
- If you want to see the list of default binds per keymap, see [this page](./DEFAULT-BINDS-&-OPTIONS.md).
- If you want to learn more about `.inputrc` files, see [this page](https://man7.org/linux/man-pages/man3/readline.3.html).

# Table of Contents

- [Table of Contents](#table-of-contents)
- [Keymaps](#keymaps)
    - [Available keymaps](#available-keymaps)
    - [Key dispatch & matching](#key-dispatch--matching)
    - [Example](#example)
- [Commands](#commands)
    - [Movement](#movement)
    - [Modifying Text](#modifying-text)
    - [Arguments](#arguments)
    - [Macros](#macros)
    - [History Control](#history-control)
    - [Completion](#completion)
    - [Miscellaneous](#miscellaneous)
    - [Text Objects](#text-objects)

-----
# Keymaps 

As explained in the `README`, the readline library model is strongly influenced by the Z-Shell architecture,
which rests on a local (potentially empty) and a main keymap against which to test the keys input by the user.

This model is quite powerful and efficient, as it allows for combinatory use of various line editing commands, 
with a pretty reliable fallback behavior, as well as a strong integration and support of completion actions, 
incremental search, history control, etc.

The shell has a `local` keymap (empty by default) and a `main` keymap (never empty):
- `local`: can be `vi-opp`, `vi-visual`, `isearch` or `menu-select`.
- `main`: can be `emacs-*`, `vi-insert` or `vi-command`.

## Available keymaps 
Initially, there are 8 **editing** keymaps:
* `emacs` - EMACS emulation 
* `emacs-standard` - Alias for `emacs` 
* `emacs-ctlx` - Subset of the `emacs` keymap, containing `Ctrl-X`-based binds 
* `emacs-meta` - Subset of the `emacs` keymap, containing `Meta-key`-based binds 
* `vi-insert` - Vim insert mode 
* `vi` - Vim command/normal mode (aliases `vi-move`, `vi-command`)
* `vi-opp` - Vim operator pending 
* `vi-visual` - Vim visual selection

Additionally, two other keymaps:
* `isearch` - Incremental search mode, used either for history or tab completion results incremental search (eg. fuzzy search).
* `menu-select` - Completion keymap, active when the user asks for completions.

## Key dispatch & matching
When a user enters a key, it is first tested against the local keymap. If the key matches one of the binds, and that no
other commands in the local keymap match this key by prefix, the widget is executed, and the shell goes back to read another key.

When no widget is matched in the `local` keymap (or if it's empty), the same process happens in the `main` keymap.
When no keys are matched here either, the key is litteraly inserted in the command line.

The readline library does _not_ make use of timeouts to read keys and match commands against them: when a given key
matches one widget bind exactly, but also one or more other commands by prefix, the perfectly matched widget is
stored, and the shell reads another key. If the aggregate keys match a widget, it will be executed, and the one
saved previously is dropped. Otherwise, if the keys don't match anything, the saved widget is executed with the
second key passed as argument to it.

## Example
An example is the `vaW` combination in Vim mode. Here is the process explained step by step:
- User enters `v`: the shell enters Visual mode.
- User enters `a`: the `vi-select-surround` widget matches exactly, but `select-a-word`, `select-in-word`, etc. 
  also match by prefix: `vi-select-surround` is cached, and the `a` key is kept in the stack.
- User enters `"`: The aggregate keys are now `a"`. Since no bind exists in Vim 
  operator pending mode for it, the `vi-select-surround` widget is executed with
  `"` as the character to search as a surround element.
- Alternatively, user enters `W`: The aggregate keys are `aW`, which matches `select-a-blank-word`:
  this widget is thus executed, and the previously cached `vi-select-surround` is dropped.
  

----
# Commands 

As this library is fully compatible with the legacy readline library, while enhancing it with a Z-Shell 
architecture and many of the latter's specific commands, several things should be noted: 
- Commands already existing in legacy readline will have the same name (eg. `vi-change-to`).
- Commands only existing in the Z-Shell will have the same name as Z-Shell ones.
- When commands exist in both, command will have the legacy readline name.
- Some commands do not exist in either library (eg. `keyword-increase`, `autosuggest-toggle`, etc) and have names of their own.

The notation is as follows :
- `widget-name` `(alias)` `[zsh name]`
    Description


----
## Movement

- `clear-screen`    
    Clear the current screen and redisplay the prompt and input line.
    This does not clear the terminal's output buffer.

- `clear-display`    
    Clear the current screen and redisplay the prompt and input line.
    This does clear the terminal's output buffer.

- `character-search`    
    A character is read and point is moved to the next occurrence of that character.  
    A negative argument searches for previous occurrences.

- `character-search-backward`    
    A character is read and point is moved to the previous occurrence of that character.  
    A negative argument searches for subsequent occurrences.

- `redraw-current-line`    
    Refreshes the current prompt, input line and any helpers in place.

- `vi-backward-bigword` `[vi-backward-blank-word]`     
    Move backward one word, where a word is defined as a series of non-blank characters.

- `vi-backward-end-bigword` `[vi-backward-blank-word-end]`    
    Move to the end of the previous word, where a word is defined as a series of non-blank characters.

- `backward-char`    
    Move backward one character.

- `vi-backward-char`    
    Move backward one character, without changing lines.

- `backward-word` `(vi-prev-word)`     
    Move to the beginning of the previous word.

- `vi-backward-word`    
    Move to the beginning of the previous word, vi-style.

- `vi-backward-end-word`    
    Move to the end of the previous word, vi-style.

- `beginning-of-line`    
    Move to the beginning of the line. If already at the beginning of the line, move to the beginning of the previous line, if any.

- `end-of-line`    
    Move to the end of the line. If already at the end of the line, move to the end of the next line, if any.

- `vi-end-of-line`    
    Move to the end of the line.

- `vi-forward-bigword` `[vi-forward-blank-word]`    
    Move forward one word, where a word is defined as a series of non-blank characters.

- `vi-end-word` `[vi-forward-word-end]`     
    Move to the end of the next word.

- `vi-end-bigword` `[vi-forward-blank-word-end]`    
    Move to the end of the current word, or, if at the end of the current word, to the end 
    of the next word, where a word is defined as a series of non-blank characters.

- `forward-char`    
    Move forward one character, if not at the end of the line.
    If history autosuggestions are enabled and a suggestion exists, and if the cursor is currently at the end 
    of the line, the suggested line is inserted into the buffer and the cursor moved to the end of it. 

- `vi-forward-char`    
    Move forward one character, if not at the end of the line.
    If history autosuggestions are enabled and a suggestion exists, and if the cursor is currently at the end 
    of the line, the suggested line is inserted into the buffer and the cursor moved to the end of it. 

- `vi-find-next-char`    
    Read a character from the keyboard, and move to the next occurrence of it in the line.

- `vi-char-search`    
    All in one combination of `vi-find-*char` commands. 

- `vi-find-next-char-skip`    
    Read a character from the keyboard, and move to the position just before the next occurrence of it in the line.

- `vi-find-prev-char`    
    Read a character from the keyboard, and move to the previous occurrence of it in the line.

- `vi-find-prev-char-skip`    
    Read a character from the keyboard, and move to the position just after the previous occurrence of it in the line.

- `vi-first-print` `[vi-first-non-blank]`     
    Move to the first non-blank character in the line.

- `vi-forward-word` `(vi-next-word)`    
    Move forward one word, vi-style. 
    If history autosuggestions are enabled and a suggestion exists, and if the cursor is currently at the end 
    of the line, a word from the suggested line is inserted into the buffer and the cursor moved to its beginning.

- `forward-word`    
    Move to the beginning of the next word. The editor’s idea of a word is specified with the WORDCHARS parameter.
    If history autosuggestions are enabled and a suggestion exists, and if the cursor is currently at the end 
    of the line, a word from the suggested line is inserted into the buffer and the cursor moved to its beginning.

- `vi-column` `[vi-goto-column]`    
    Move to the column specified by the numeric argument.

- `vi-goto-mark`    
    Move to the specified mark.

- `next-screen-line`    
    Move down one line if the current buffer has more than one line.
 
- `previous-screen-line`    
    Move up one line if the current buffer has more than one line.

- `shell-backward-word`    
    Move to the beginning of the current or previous word.
    The editor's idea of a word is defined by classic sh-style word splitting:
    any non-spaced sequence of characters, or a quoted sequence.

- `shell-forward-word`    
    Move forward to the beginning of the next word.
    The editor's idea of a word is defined by classic sh-style word splitting:
    any non-spaced sequence of characters, or a quoted sequence.

----
## Modifying Text

- `vi-append-eol` `[vi-add-eol]`     
    Move to the end of the line and enter insert mode.

- `vi-append-mode` `[vi-add-next]`    
    Enter insert mode after the current cursor position, without changing lines.

- `emacs-editing-mode`    
    When in vi command mode, this causes a switch to emacs editing mode.

- `backward-delete-char`    
    Delete the character behind the cursor.

- `vi-backward-delete-char`    
    Delete the character behind the cursor, without changing lines. 

- `forward-backward-delete-char`    
    Delete the character under the cursor, unless the cursor is at the
    end of the line, in which case the character behind the cursor is deleted.

- `backward-kill-line` `(unix-line-discard)`    
    Kill from the beginning of the line to the cursor position.

- `backward-kill-word` `(unix-word-rubout)`     
    Kill the word behind the cursor.

- `vi-unix-word-rubout` `[vi-backward-kill-word]`    
    Kill the word behind the cursor, without going past the point where insert mode was last entered.

- `capitalize-word`    
    Capitalize the current word and move past it.

- `vi-change-to` `[vi-change]`     
    Read a movement command from the keyboard, and kill from the cursor position to the endpoint of the movement. 
    Then enter insert mode. If the command is vi-change, change the current line.

- `vi-change-eol`    
    Kill to the end of the line and enter insert mode.

- `copy-region-as-kill`    
    Copy the area from the cursor to the mark to the kill buffer.

- `copy-backward-word` `[copy-prev-word]`     
    Duplicate the word to the left of the cursor.

- `copy-prev-shell-word`    
    Like copy-prev-word, but the word is found by using shell parsing, whereas copy-prev-word looks for blanks. 
    This makes a difference when the word is quoted and contains spaces.

- `vi-delete-to` `[vi-delete]`    
    Read a movement command from the keyboard, and kill from the cursor position to the endpoint of the movement. 
    If the command is vi-delete, kill the current line.

- `delete-char`    
    Delete the character under the cursor.

- `vi-delete` `[vi-delete-char]`     
    Delete the character under the cursor, without going past the end of the line.

- `delete-word`    
    Delete the current word from the cursor point up to the end of it.

- `delete-horizontal-whitespace`    
    Delete all spaces and tabs around point.

- `down-case-word`    
    Convert the current word to all lowercase and move past it.

- `vi-down-case`    
    Read a movement command from the keyboard, and convert all characters from the cursor position to the endpoint of 
    the movement to lowercase. If the movement command is vi-down-case, swap the case of all characters on the current line.

- `end-of-file` `(vi-eof-maybe)`    
    The character indicating end-of-file as set, for example, by “stty”.  
    If this character is read when there are no characters on the line, and point is at the beginning of
    the line, readline interprets it as the end of input and returns io.EOF.

- `kill-word`    
    Kill the current word.

- `shell-kill-word`    
    Kill the shell word behind point. Word boundaries are the same as those used by backward-word.
 
- `shell-backward-kill-word`    
    Kill the shell word behind point.

- `vi-insertion-mode` `(vi-editing-mode)` `[vi-insert]`    
    Enter insert mode.

- `vi-insert-beg` `[vi-insert-bol]`    
    Move to the first non-blank character on the line and enter insert mode.

- `kill-line`    
    Kill from the cursor to the end of the line. If already on the end of the line, kill the newline character.

- `vi-kill-line`    
    Kill from the cursor back to wherever insert mode was last entered.

- `vi-kill-eol`    
    Kill from the cursor to the end of the line.

- `kill-region`    
    Kill from the cursor to the mark.

- `kill-buffer`    
    Kill the entire buffer.

- `kill-whole-line`    
    Kill the current line.

- `vi-match` `[vi-match-bracket]`    
    Move to the bracket character (one of {}, () or []) that matches the one under the cursor. If the cursor is not on a bracket character, 
    move forward without going past the end of the line to find one, and then go to the matching bracket.

- `vi-change-case` `[vi-oper-swap-case]`    
    Read a movement command from the keyboard, and swap the case of all characters from the cursor position to the endpoint of the movement. 
    If the movement command is vi-oper-swap-case, swap the case of all characters on the current line.

- `overwrite-mode`    
    Toggle between overwrite mode and insert mode.

- `vi-put`    
    Combination of `vi-put-before` and `vi-put-after` (looking at the key triggering it).

- `vi-put-before`    
    Insert the contents of the kill buffer before the cursor.

- `vi-put-after`    
    Insert the contents of the kill buffer after the cursor.

- `quote-line`    
    Quote the current line; that is, put a ‘’’ character at the beginning and the end. 

- `quote-region`    
    Quote the region from the cursor to the mark.

- `quoted-insert`     
    Add the next character that you type to the line verbatim.
    This is how to insert characters like C-q, for example.

- `prefix-meta`    
    Metafy the next character typed.  ESC f is equivalent to Meta-f.

- `tab-insert`    
    Insert a tab character.

- `self-insert`    
    Insert the character typed.
    If the `output-meta` option is on and if the key is a metafied character, it's unmetafied and prefixed with ESC.    
 
- `vi-replace` `(vi-overstrike)`     
    Enter overwrite mode.

- `vi-change-char` `[vi-replace-chars]`    
    Replace the character under the cursor with a character read from the keyboard.

- `vi-subst` `[vi-substitute]`     
    Substitute the next character(s).

- `transpose-chars`    
    Drag the character before point forward over the character at point, moving point forward as well.  
    If point is at the end of the line, then this transposes the two characters before point.  
    Negative arguments have no effect.

- `transpose-words`    
    Drag the word before point past the word after point, moving point over that word as well.  
    If point is at the end of the line, this transposes the last two words on the line. 
    If a numeric argument is given, the word to transpose is chosen backward.

- `shell-transpose-words`    
    Drag the shell word before point past the shell word after point, moving point over that shell word as well. 
    If point is at the end of the line, this transposes the last two words on the line.
    If a numeric argument is given, the word to transpose is chosen backward.

- `do-lowercase-version`    
    If the metafied character x is uppercase, run the command that is bound to the corresponding metafied lowercase character.
    The behavior is undefined if x is already lowercase.

- `vi-up-case`    
    Read a movement command from the keyboard, and convert all characters from the cursor position to the endpoint of the movement to lowercase. 
    If the movement command is vi-up-case, swap the case of all characters on the current line.

- `up-case-word`    
    Convert the current word to all caps and move past it.

- `yank`    
    Insert the contents of the kill buffer at the cursor position.

- `yank-pop`    
    Rotate the kill ring, and yank the new top. Only works following yank or yank-pop.

- `vi-yank-to` `[vi-yank]`     
    Read a movement command from the keyboard, and copy the region from the cursor position to the endpoint 
    of the movement into the kill buffer. If the command is vi-yank, copy the current line.

- `vi-yank-whole-line`    
    Copy the current line into the kill buffer.

- `vi-yank-pop`    
    Rotate the kill ring, and yank the new top. Only works following yank or yank-pop.
   
- `keyword-increase`   
      Increments the current word under the cursor.
      The command supports digits arguments (vi iterations in Vim mode), and will perform the increment n-arg times (can be negative).
      The following word types can be incremented/decremented:  
        - Booleans: `true`|`false`, `t`|`f`, `on`|`off`, `yes`|`no`, `y`|`n`.  
        - Operators: `&&`|`||`, `++`|`--`, `==`|`!=`, `===`| `!==`, `+`| `-`, `-`| `*`, `*`| `/`, `/`| `+`, `and`| `or`.  
        - Hex digits `0xDe` => `0xdf`, `0xdE` => `0xDF`, `0xde0` => `0xddf`, `0xffffffffffffffff` => `0x0000000000000000`.  
        - Binary digits: `0b1` => `0b10`, `0B0` => `0B1`, etc.  
        - Integers. 

- `keyword-decrease`  
      Decrements the current word under the cursor.
      The command supports digits arguments (vi iterations in Vim mode), and will perform the increment n-arg times (can be negative).
      The following word types can be incremented/decremented:   
        - Booleans: `true`|`false`, `t`|`f`, `on`|`off`, `yes`|`no`, `y`|`n`.  
        - Operators: `&&`|`||`, `++`|`--`, `==`|`!=`, `===`| `!==`, `+`| `-`, `-`| `*`, `*`| `/`, `/`| `+`, `and`| `or`.  
        - Hex digits `0xDe` => `0xdf`, `0xdE` => `0xDF`, `0xde0` => `0xddf`, `0xffffffffffffffff` => `0x0000000000000000`.  
        - Binary digits: `0b1` => `0b10`, `0B0` => `0B1`, etc.  
        - Integers. 

----
## Arguments

- `digit-argument` `(vi-arg-digit)`  
    Start a new numeric argument, or add to the current one. 
    If the last command executed was a digit as part of an argument, continue the argument. 
    Otherwise and if the key triggering the command was 0, execute vi-beginning-of-line.

----
## Macros

- `start-kbd-macro`  
    Begin saving the characters typed into the current keyboard macro.

- `end-kbd-macro`  
    Stop saving the characters typed into the current keyboard macro and store the definition.

- `call-last-kbd-macro`  
    Re-execute the last keyboard macro defined, by making the
    characters in the macro appear as if typed at the keyboard.

- `print-last-kbd-macro`  
    Print the last keyboard macro defined in a format suitable for the inputrc file.

- `macro-toggle-record`  
    Either starts recording a macro (if not yet recording), or stops it.
    If the command is about to start recording a macro, it will read an additional argument key (must be a letter), 
    to be used as the macro "name", just like macro recording and use work in Vim.
    This command thus works "Vim-style", and should probably be used only when using Vim editing mode.
 
- `macro-run`            
    Reads a key from the keyboard, and runs the macro stored for this key identitier.
    This mimics the Vim-style or running macros. If no macro is recorded for this key,
    or if the key is invalid, nothing happens.

----
## History Control

- `beginning-of-buffer-or-history`  
    Move to the beginning of the buffer, or if already there, move to the first event in the history list.

- `beginning-of-line-hist`  
    Move to the beginning of the line. If already at the beginning of the buffer, move to the previous history line.

- `beginning-of-history`  
    Move to the first event in the history list.

- `down-line-or-history`  
    Move down a line in the buffer, or if already at the bottom line, move to the next event in the history list.

- `vi-down-line-or-history`  
    Move down a line in the buffer, or if already at the bottom line, move to the next event in the history list. 
    Then move to the first non-blank character on the line.

- `next-history` `[down-history]`  
    Move to the next event in the history list.

- `down-line-or-select`  
    If the cursor is on the last line of the buffer, start an incremental
    search forward on the history lines. Otherwise, move up a line in the buffer.

- `end-of-buffer-or-history`  
    Move to the end of the buffer, or if already there, move to the last event in the history list.

- `end-of-line-hist`  
    Move to the end of the line. If already at the end of the buffer, move to the next history line.

- `end-of-history`  
    Move to the last event in the history list. 

- `fetch-history`  
    With a numeric argument, fetch that entry from the history list and make it the current line.  
    Without an argument, move back to the first entry in the history list.

- `history-search-backward` `[history-beginning-search-backward]`  
    Search backward in the history for a line beginning with the current line up to the cursor. 
    This leaves the cursor in its original position.

- `reverse-search-history` `[history-incremental-search-backward]`  
    Search backward incrementally for a specified string. The search is case-insensitive if the search string does not have uppercase 
    letters and no numeric argument was given. The string may begin with ‘^’ to anchor the search to the beginning of the line. 
    A restricted set of editing functions is available in the mini-buffer. Keys are looked up in the special isearch keymap, 
    and if not found there in the main keymap (note that by default the isearch keymap is empty). 
    An interrupt signal, as defined by the `isearch-terminators` config option, will stop the search and go back to the original line. 

- `forward-search-history` `[history-incremental-search-forward]`  
    Search forward incrementally for a specified string. The search is case-insensitive if the search string does not have uppercase 
    letters and no numeric argument was given. The string may begin with ‘^’ to anchor the search to the beginning of the line. 
    The functions available in the mini-buffer are the same as for history-incremental-search-backward.

- `infer-next-history`  
    Search in the history list for a line matching the current one and fetch the event following it.

- `up-line-or-history`  
    Move up a line in the buffer, or if already at the top line, move to the previous event in the history list.

- `up-line-or-search`  
    If the cursor is on the first line of the buffer, start an incremental
    search backward on the history lines. Otherwise, move up a line in the buffer.
 
- `previous-history` `[up-history]`   
    Move to the previous event in the history list.

- `history-search-forward` `[history-beginning-search-forward]`  
    Search forward in the history for a line beginning with the current line up to the cursor. 
    This leaves the cursor in its original position.

- `non-incremental-reverse-search-history`  
    Proposes history lines matching the current input buffer as completions, in reverse order and autocomplete mode.

- `non-incremental-forward-search-history`  
    Proposes history lines matching the current input buffer as completions, in forward order and autocomplete mode.

- `history-substring-search-backward`  
    Search backward through the history for the string of characters between the start of the current line and the current cursor position.
    The search string may match anywhere in a history line. This is a non-incremental search.
    
- `history-substring-search-forward`  
    Search forward through the history for the string of characters between the start of the current line and the current cursor position.
    The search string may match anywhere in a history line. This is a non-incremental search.

- `vi-search`  
    Search forward or backward through the history for the string
    of characters between the start of the current line and the point.
    This is a non-incremental search.

- `vi-search-again`  
    Search again, through the history for the string of characters
    between the start of the current line and the point, using the
    same search string used by the previous search.
    This is a non-incremental search.
 
- `autosuggest-accept`  
    If a line is autosuggested for the current buffer, insert the suggestion and move the cursor at the end of it. 

- `autosuggest-execute`  
    If a line is autosuggested for the current buffer, insert the suggestion and execute the line. 

- `autosuggest-toggle`  
    Enable/disable line history autosuggestions. 

- `autosuggest-disable`  
    Disable line history autosuggestions. 

- `autosuggest-enable`  
    Enable line history autosuggestions. 

- `inline-suggest-accept`  
    If the application has set an inline suggestion (see the [Inline Suggestions](./INLINE-SUGGESTIONS.md) page) and it extends
    the current line with the cursor at its end, insert the whole suggestion and move the cursor to the end of it. 

- `inline-suggest-accept-word`  
    Same as `inline-suggest-accept`, but only accepts the next word of the inline suggestion (leading whitespace plus the
    following run of non-whitespace runes), leaving the rest as a suggestion. 

- `yank-last-arg` `(insert-last-argument) or (vi-yank-arg)`  
    Insert the last argument to the previous command (the last word of the previous history entry).  
    With a numeric argument, behave exactly like yank-nth-arg.  Successive calls to yank-last-arg move back through the history list,
    inserting the last word (or the word specified by the argument to the first call) of each line in turn.
    Any numeric argument supplied to these successive calls determines the direction to move through the history.
    A negative argument switches the direction through the history (back or forward).  
    The history expansion facilities are used to extract the last argument, as if the "!$" history expansion had been specified.

- `yank-nth-arg`  
    Insert the first argument to the previous command (usually the second word on the previous line) at point.  
    With an argument n, insert the nth word from the previous command (the words in the previous command begin with word 0).
    A negative argument inserts the nth word from the end of the previous command.  
    Once the argument n is computed, the argument is extracted as if the "!n" history expansion had been specified.

- `magic-space`  
    Perform history expansion on the current line and/or insert a space.
    If the current blank word under cursor starts with an exclamation mark, 
    the word up to the cursor is matched as a prefix against the history lines, 
    and the first match is inserted in place of it.

- `history-source-next`  
    If more than one source of command history is bound to the shell, cycle to the next one and use it 
    for all history search operations, movements across lines, their respective undo histories, etc.

- `history-source-prev`  
    If more than one source of command history is bound to the shell, cycle to the previous one and use it 
    for all history search operations, movements across lines, their respective undo histories, etc.

- `save-line`    
    Write the current line to the history if it is not empty (without executing it), and clear the line buffer.


----
## Completion

- `accept-and-menu-complete`  
    In a menu completion, insert the current completion into the buffer, and advance to the next possible completion.

- `complete`  
    Attempt completion on the current word.

- `delete-char-or-list`   
    Delete the character under the cursor. If the cursor is at the end of the line, list possible completions for the current word.

- `possible-completions` `[list-choices]`   
    List possible completions for the current word.

- `insert-completions`  
    Insert all completions for the current word into the line.

- `menu-complete`  
    Like complete-word, except that menu completion is used.

- `menu-incremental-search`  
    In a menu completion (wether a candidate is selected or not), start incremental-search (fuzzy search) on the results.
    Search backward incrementally for a specified string. The search is case-insensitive if the search string does not have uppercase 
    letters and no numeric argument was given. The string may begin with ‘^’ to anchor the search to the beginning of the line. 
    A restricted set of editing functions is available in the mini-buffer. Keys are looked up in the special isearch keymap, 
    On each change in the mini-buffer, any currently selected candidate is dropped from the line and the menu.
    An interrupt signal, as defined by the `isearch-terminators` config option, will stop the search and go back to the original line. 

- `menu-complete-backward` `[reverse-menu-complete]`  
    Perform menu completion, like menu-complete, except that if a menu completion 
    is already in progress, move to the previous completion rather than the next.

- `menu-complete-next-tag`  
    In a menu completion, if there are several tags of completions, go to the first result of the next tag.

- `menu-complete-prev-tag`  
    In a menu completion, if there are several tags of completions, go to the first result of the previous tag.

- `vi-registers-complete`  
    Open a completion menu (similar to `menu-complete`) with all currently populated Vim registers.


----
## Miscellaneous

- `abort`  
    Abort the current editing command.
    If one of the completion or non/incremental-search modes are active, only cancel them and nothing else.

- `accept-and-hold`  
    Accept the current input line (execute it) and keep it as the buffer on the next readline loop. 

- `accept-and-infer-next-history`  
    Execute the contents of the buffer. Then search the history list for a line matching the current one and push the event following onto the buffer stack.

- `accept-line`  
    Finish editing the buffer. Normally this causes the buffer to be executed as a shell command.

- `dump-functions`  
    Print all of the functions and their key bindings to the readline output stream.  
    If a numeric argument is supplied, the output is formatted in such a way that it
    can be made part of an inputrc file.

- `dump-macros`  
    Print all of the readline key sequences bound to macros and the strings they output.  
    If a numeric argument is supplied, the output is formatted in such a way that it
    can be made part of an inputrc file.

- `dump-variables`  
    Print all of the settable variables and their values to the readline output stream.  
    If a numeric argument is supplied, the output is formatted in such a way that it
    can be made part of an inputrc file.

- `insert-comment`  
    Without a numeric argument, the value of the readline comment-begin variable is inserted at the beginning of the
    current line.  If a numeric argument is supplied, this command acts as a toggle: if the characters at the beginning 
    of the line do not match the value of comment-begin, the value is inserted, otherwise the characters in comment-begin 
    are deleted from the beginning of the line.  
    In either case, the line is accepted as if a newline had been typed.  
    The default value of comment-begin makes the current line a shell comment.
    If a numeric argument causes the comment character to be removed, the line will be executed by the shell.
 
- `operate-and-get-next` `[accept-line-and-down-history]`  
    Execute the current line, and push the next history event on the buffer stack.

- `vi-movement-mode` `[vi-cmd-mode]`   
    Enter command mode; that is, select the ‘vicmd’ keymap. Yes, this is bound by default in emacs mode.

- `exchange-point-and-mark`  
    Exchange the cursor position (point) with the position of the mark. Unless a negative numeric argument is given, the region between point and mark 
    is activated so that it can be highlighted. If a zero numeric argument is given, the region is activated but point and mark are not swapped.

- `vi-set-buffer`  
    Specify a buffer to be used in the following command. See the [registers section](./VIM-MODE.md#Registers) in the Vim page.

- `vi-set-mark`  
    Set the specified mark at the cursor position.

- `set-mark` `[set-mark-command]`  
    Set the mark at the cursor position. If called with a negative numeric argument, do not set the mark but deactivate the region 
    so that it is no longer highlighted (it is still usable for other purposes). Otherwise the region is marked as active.

- `undo` `(vi-undo)`  
    Incrementally undo the last text modification. 
    Note that when invoked from vi command mode, the full prior change made in insert mode is reverted, the changes having been merged when command mode was selected.

- `redo` `(vi-undo)`  
    Incrementally redo undone text modifications.

- `revert-line`  
    Undo all changes made to this line.
    This is like executing the undo command enough times to return the line to its initial state.

- `re-read-init-file`  
    Read in the contents of the inputrc file, and incorporate any bindings or variable assignments found there.

- `vi-visual-mode` `[visual-mode]`  
    Toggle vim-style visual selection mode. 

- `vi-visual-line-mode` `[visual-line-mode]`  
    Toggle vim-style line-wise visual selection mode.

- `edit-command-line` `(vi-edit-command-line)`  
    Invoke an editor on the current command line.
    Readline attempts to invoke $VISUAL, $EDITOR, and emacs as the editor, in that order.

- `edit-and-execute-command` `(vi-edit-and-execute-command)`  
    Invoke an editor on the current command line, and execute the result as shell commands.
    Readline attempts to invoke $VISUAL, $EDITOR, and emacs as the editor, in that order.

- `bracketed-paste-begin`  
    Insert pasted text as literal content. When bracketed paste is enabled (see the `enable-bracketed-paste` option),
    terminals wrap pasted text between `ESC [200~` and `ESC [201~` markers; this command, bound to the start marker,
    reads everything up to the end marker and inserts it verbatim instead of interpreting control characters or key
    bindings contained in it. Carriage returns in the pasted text are normalised to newlines so multi-line pastes are
    stored and displayed correctly.  
    Bound by default to `\e[200~`; it only fires when `enable-bracketed-paste` is on.

- `skip-csi-sequence`  
    Read and discard the rest of a multi-byte CSI sequence, the kind terminals send for keys such as Home, End,
    Page-Up/Down or the function keys (an `ESC [` introducer, optional parameter/intermediate bytes, then a single final byte).  
    Bind it to `\e[` so that special keys without an explicit binding are swallowed instead of inserting their trailing
    bytes (e.g. `15~`, `Z`, `1;5D`) as stray characters into the line. Longer, explicitly bound sequences such as the
    arrow keys (`\e[A`) still take precedence; this only catches sequences nothing else claims.  
    Unbound by default (as in GNU readline); enable it from your inputrc with `"\e[": skip-csi-sequence`.


----
## Text Objects

Text objects are commands that can be used to select a block of text according to some criteria. 
They are a feature of the vim text editor and so are primarily intended for use with vi operators or from visual selection mode. 
However, they can also be used from vi-insert or emacs mode. Key bindings listed below apply to the viopp and visual keymaps.

- `select-a-blank-word` `(aW)`  
    Select a word including adjacent blanks, where a word is defined as a series of non-blank characters.

- `select-a-shell-word` `(aa)`  
    Select the current command argument applying the normal rules for quoting.

- `select-a-word` `(aw)`  
    Select a word including adjacent blanks, using the normal vi-style word definition.

- `select-in-blank-word` `(iW)`  
    Select a word, where a word is defined as a series of non-blank characters.

- `select-in-shell-word` `(ia)`  
    Select the current command argument applying the normal rules for quoting. 
    If the argument begins and ends with matching quote characters, these are not included in the selection.

- `select-in-word` `(iw)`  
    Select a word, using the normal vi-style word definition. 

- `select-keyword-next`  
    Considers the blank word under cursor, and tries a series of regular expressions on it
    to match various patterns: URL and their various subcomponents (host/path/params, etc).
    When one of the regular expressions succeeds, the match is visually selected,
    otherwise nothing is selected (if selection was active, it will stay the same)
    When repeatedly calling this function while in visual selection mode, the shell will
    cycle through either the current matcher's capturing subgroups (such as the parts of a URL),
    or cycle through the next matcher (for instance, attempting to grap an IP after trying URL).

- `select-keyword-prev`  
    Identical to select-keyword-prev, except that the matcher/subgroup cycling occurs backward.

- `vi-select-surround`  
    Read a key from the keyboard, and attempt to create a selection
    consisting of a pair of this character, if any such pair can be found.

- `vi-select-inside`  
    Read a key from the keyboard, and attempt to select a region surrounded by those keys.
    If the key triggering this command is 'i', the selection excludes the surrounding chars.
