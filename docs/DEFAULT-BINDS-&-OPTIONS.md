> **Runnable example:** [`cmd/keymaps`](../cmd/keymaps)


This page contains a list of global options and binds for all keymaps.
Some keymaps come with an additional explanations on their role.
Note that the binds below are not in an `.inputrc`-compliant notation, 
for the sake of conciseness and because the list below is the default.

Note that the default options/binds is different from the ones 
of the legacy readline library, as functionality has been enhanced.

## Options

Some global options have not been implemented: some because they were not useful anymore (some completion ones),
out of the library scope (other completion ones), or not yet faced by the library author. 
Some of those might be implemented in the future if the need arises.

Notation: <set> `<option-name>` `<value>` \[NI/S] (\[NI]: not implemented, \[S]: specific to this library).

set `active-region-start-color` `""`     
set `autocomplete` `off` \[S]     
set `autopairs` `off` \[S]     
set `bell-style` `none` \[NI]     
set `bind-tty-special-chars` `on`     
set `blink-matching-paren` `off`     
set `byte-oriented` `off`     
set `colored-completion-prefix` `off`     
set `colored-stats` `off` \[NI]     
set `comment-begin` `#`     
set `completion-description-style` `""` \[S]
set `completion-display-width` `-1`     
set `completion-ignore-case` `off`     
set `completion-list-separator` `--` \[S]
set `completion-map-case` `off`     
set `completion-prefix-display-length` `0`     
set `completion-query-items` `100`     
set `completion-selection-style` `""` \[S]    
set `convert-meta` `off`     
set `cursor-<keymap-name>` `default` \[S]
set `cursor-position-probe` `on` \[S]
set `disable-completion` `off`     
set `echo-control-characters` `on`     
set `editing-mode` `emacs`     
set `emacs-mode-string` `@`     
set `enable-bracketed-paste` `on` \[NI]     
set `enable-keypad` `off` \[NI]     
set `enable-meta-key` `off` \[NI]     
set `expand-tilde` `off` \[NI]     
set `history-autosuggest` `off` \[S]     
set `history-preserve-point` `off`     
set `history-size` `0`     
set `horizontal-scroll-mode` `off` \[NI]     
set `input-meta` `off`     
set `keymap` `emacs`     
set `keyseq-timeout` `50`     
set `mark-directories` `on` \[NI]     
set `mark-modified-lines` `off` \[NI]     
set `mark-symlinked-directories` `on` \[NI]     
set `match-hidden-files` `on` \[NI]     
set `menu-complete-display-prefix` `off`     
set `meta-flag` `off`     
set `output-meta` `Off`     
set `page-completions` `on` \[NI]     
set `prefer-visible-bell` `on` \[NI]     
set `print-completions-horizontally` `off`     
set `transient-prompt` `off` \[S]     
set `revert-all-at-newline` `off`     
set `show-all-if-ambiguous` `off` \[NI]     
set `show-all-if-unmodified` `off` \[NI]     
set `show-mode-in-prompt` `off`     
set `skip-completed-text` `off`     
set `usage-hint-always` `off` \[S]     
set `vi-cmd-mode-string` `(cmd)`     
set `vi-ins-mode-string` `(ins)`     
set `visible-stats` `off` \[NI]     
     

## Emacs 

### Emacs standard
The emacs keymap is a combination of both `emacs-ctlx` and `emacs-meta` keymaps below,
but additionally includes/overwrites some of the binds by the list below:

"C-@"  set-mark          
"C-A"  beginning-of-line     
"C-B"  backward-char     
"C-D"  delete-char     
"C-E"  end-of-line     
"C-F"  forward-char     
"C-G"  abort     
"C-H"  backward-delete-char     
"C-I"  complete     
"C-J"  accept-line     
"C-K"  kill-line     
"C-L"  clear-screen     
"C-M"  accept-line     
"C-N"  next-history     
"C-P"  previous-history     
"C-Q"  quoted-insert     
"C-R"  reverse-search-history     
"C-S"  forward-search-history     
"C-T"  transpose-chars     
"C-U"  unix-line-discard     
"C-V"  quoted-insert     
"C-W"  unix-word-rubout     
"C-Y"  yank     
"C-]"  character-search     
"C-_"  undo     
" " to "/"  self-insert     
"0"  to "9"  self-insert     
":"  to "~"  self-insert     
"C-?"  backward-delete-char     
     
### Emacs Meta     

"M-C-G"  abort     
"M-C-H"  backward-kill-word     
"M-C-I"  tab-insert     
"M-C-J"  vi-editing-mode     
"M-C-L"  clear-display     
"M-C-M"  vi-editing-mode     
"M-C-R"  revert-line     
"M-C-Y"  yank-nth-arg     
"M-C-\["  complete     
"M-C-]"  character-search-backward     
"M-C-^"  copy-backward-word     
"M-space"  set-mark     
"M-#"  insert-comment     
"M-&"  tilde-expand     
"M-*"  insert-completions     
"M--"  digit-argument     
"M-."  yank-last-arg     
"M-'"  quote-line     
"M-0"  digit-argument     
"M-1"  digit-argument     
"M-2"  digit-argument     
"M-3"  digit-argument     
"M-4"  digit-argument     
"M-5"  digit-argument     
"M-6"  digit-argument     
"M-7"  digit-argument     
"M-8"  digit-argument     
"M-9"  digit-argument     
"M-<"  beginning-of-buffer-or-history     
"M-="  possible-completions     
"M->"  end-of-buffer-or-history     
"M-?"  possible-completions     
"M-B"  backward-word     
"M-C"  capitalize-word     
"M-D"  kill-word     
"M-F"  forward-word     
"M-L"  down-case-word     
"M-M"   copy-prev-shell-word     
"M-N"  history-search-forward     
"M-P"  history-search-backward     
"M-R"  revert-line     
"M-T"  transpose-words     
"M-U"  up-case-word     
"M-W"   kill-region     
"M-Y"  yank-pop     
"M-|"   vi-column     
"M-\"  delete-horizontal-whitespace     
"M-~"  tilde-expand     
"M-C-?"  backward-kill-word     
"M-_"  yank-last-arg     
"M-[A"  up-line-or-search     
"M-[B"  down-line-or-select     
     
### Emacs CtrlX      

"C-XC-G"  abort     
"C-XC-R"  re-read-init-file     
"C-XC-U"  undo     
"C-XC-X"  exchange-point-and-mark     
"C-Xr"  reverse-search-history     
"C-Xs"  forward-search-history     
"C-Xu"  undo     
"C-X("  start-kbd-macro     
"C-X)"  end-kbd-macro     
"C-XE"  call-last-kbd-macro     
"C-XC-?"  backward-kill-line     
"C-xC-b"  vi-match     
"C-xC-e"  edit-command-line     
"C-xC-n"  infer-next-history     
"C-xC-o"  overwrite-mode     


## Vim

The various Vim keymaps have been significantly enhanced, either with new binds,
new keymaps (Operator pending), or better in-code functionality.

### Insert

"C-A"   beginning-of-line    
"C-B"   backward-char    
"C-D"  vi-eof-maybe    
"C-F"   forward-char    
"C-H"  backward-delete-char    
"C-I"  complete    
"C-J"  accept-line    
"C-K"   kill-line    
"C-L"   clear-screen    
"C-M"  accept-line    
"C-N"   down-line-or-history    
"C-O"   operate-and-get-next    
"C-P"   up-line-or-history    
"C-Q"   accept-and-infer-next-history    
"C-R"  reverse-search-history    
"C-S"  forward-search-history    
"C-T"  transpose-chars    
"C-U"  unix-line-discard    
"C-V"  quoted-insert    
"C-W"  unix-word-rubout    
"C-Y"  yank    
"C-["  vi-movement-mode    
"C-_"  undo    
" " to "\~"  self-insert    
"C-?"  backward-delete-char    
"M-q"   macro-toggle-record    
"M-@"   macro-run    
"M-'"   vi-registers-complete    
"M-[3~" delete-char    
"M-[H"  beginning-of-line    
"M-[F"  end-of-line    
"M-[A"  up-line-or-search    
"M-[B"  down-line-or-search    
    
### Command    
    
"C-A"  keyword-increase    
"C-D"  vi-eof-maybe    
"C-E"  emacs-editing-mode    
"C-G"  abort    
"C-H"  backward-char    
"C-J"  accept-line    
"C-K"  kill-line    
"C-L"  clear-screen    
"C-M"  accept-line    
"C-N"  next-history    
"C-P"  previous-history    
"C-Q"  quoted-insert    
"C-R"  reverse-search-history    
"C-S"  forward-search-history    
"C-T"  transpose-chars    
"C-U"  unix-line-discard    
"C-V"  quoted-insert    
"C-W"  unix-word-rubout    
"C-X"  keyword-decrease    
"C-Y"  yank    
"C-\_"  vi-undo    
"M-<"  beginning-of-buffer-or-history    
"M->"  end-of-buffer-or-history    
"M-'"  quote-line    
"\M-[3~"   delete-char    
"\M-[6~"   down-line-or-history    
"\M-[5~"   up-line-or-history    
"\M-[H"    beginning-of-line    
"\M-[F"    end-of-line    
"\M-[A"    history-search-backward    
"\M-[B"    menu-complete    
"\M-[C"    vi-forward-char    
"\M-[D"    vi-backward-char    
"\M-[3;5~" kill-word    
"\M-[1;5C" forward-word    
"\M-[1;5D" backward-word    
" "  vi-forward-char    
"#"  insert-comment    
"$"  vi-end-of-line    
"%"  vi-match    
"&"  vi-tilde-expand    
"*"  vi-complete    
"+"  next-history    
","  vi-char-search    
"-"  previous-history    
"."  vi-redo    
"/"  vi-search    
"0"  beginning-of-line    
"1" to "9"  vi-arg-digit    
";"  vi-char-search    
"="  vi-complete    
"?"  vi-search    
"A"  vi-append-eol    
"B"  vi-backward-bigword    
"C"  vi-change-to    
"D"  vi-delete-to    
"E"  vi-end-bigword    
"F"  vi-char-search    
"G"  vi-fetch-history    
"I"  vi-insert-beg    
"N"  vi-search-again    
"O"  vi-open-line-above    
"P"  vi-put    
"R"  vi-replace    
"S"  vi-subst    
"T"  vi-char-search    
"U"  revert-line    
"V"  vi-visual-line-mode    
"W"  vi-forward-bigword    
"X"  vi-backward-delete-char    
"Y"  vi-yank-to    
"\\"  vi-complete    
"^"  vi-first-print    
"""  vi-set-buffer    
"_"  vi-yank-arg    
"`"  vi-goto-mark    
"a"  vi-append-mode    
"b"  vi-prev-word    
"c"  vi-change-to    
"d"  vi-delete-to    
"e"  vi-end-word    
"f"  vi-char-search    
"gg"  beginning-of-buffer-or-history    
"ge"  vi-backward-end-word    
"gE"  vi-backward-end-bigword    
"gu"  vi-down-case    
"gU"  vi-up-case    
"h"  backward-char    
"i"  vi-insertion-mode    
"j"  down-line-or-history    
"k"  up-line-or-history    
"l"  forward-char    
"m"  vi-set-mark    
"n"  vi-search-again    
"o"  vi-open-line-below    
"p"  vi-put    
"q"  macro-toggle-record    
"r"  vi-change-char    
"s"  vi-subst    
"t"  vi-char-search    
"u"  vi-undo    
"v"  vi-visual-mode    
"w"  vi-next-word    
"x"  vi-delete    
"y"  vi-yank-to    
"|"  vi-column    
"~"  vi-change-case    
"@"  macro-run    

### Visual 

The `vi-visual` keymap enables classic Vim-style visual selection,
either in visual line mode or in normal visual mode. 
This keymap is a local keymap: all movement commands from the
`vi-command` keymap also work while in this mode.

"M-]" vi-movement-mode    
"aW"  select-a-blank-word    
"aa"  select-a-shell-word    
"aw"  select-a-word    
"iW"  select-in-blank-word    
"ia"  select-in-shell-word    
"iw"  select-in-word    
"a"   vi-select-inside    
"c"   vi-change-to    
"d"   vi-delete-to    
"i"   vi-select-inside    
"j"   next-screen-line    
"k"   previous-screen-line    
"s"   vi-subst    
"S"   vi-add-surround    
"u"   vi-down-case    
"v"   vi-edit-command-line    
"x"   vi-delete-to    
"y"   vi-yank-to    
"~"   vi-change-case    

### Operator Pending

The `operator-pending` is the lesser known Vim keymap. It is used
for more complex actions often starting with yank/delete/change
operations, and involving a text object. A few examples are:
- `diW` to delete the bigword under (inside) cursor.
- `ya"` to yank everything around two surrounding double quotes.
- `cs"'` to change surrounding double quotes into single ones.
 
This keymap is also a local keymap: all movement commands from the
`vi-command` keymap also work while in this mode.

"M-]" vi-movement-mode    
"a"   vi-select-inside    
"aW"  select-a-blank-word    
"aa"  select-a-shell-word    
"aw"  select-a-word    
"i"   vi-select-inside    
"iW"  select-in-blank-word    
"ia"  select-in-shell-word    
"iw"  select-in-word    
"s"   vi-select-surround    
"j"   next-screen-line 
"k"   up-line    


## Completion & Isearch

### Menu-Select

The `menu-select` is only enabled when completions have been requested by the user. 
It is by default populated with keys for reverse cycling, tag-cycling, and arrow 
keys for directional cycling,

"TAB"    menu-complete    
"Shift+Tab"    menu-complete-backward    
"C-@"    accept-and-menu-complete    
"C-F"    menu-incremental-search    
"C-N"    menu-complete    
"C-P"    menu-complete-backward    
"\e[A"    menu-complete-backward    
"\e[B"    menu-complete    
"\e[C"    menu-complete    
"\e[D"    menu-complete-backward    
"\e[1;5A" menu-complete-prev-tag    
"\e[1;5B" menu-complete-next-tag    

### Incremental Search

The `isearch` keymap is a special, empty keymap: when incrementally-searching
in the history lines or available completions, only a subset of the `emacs`/`vi-insert`
commands are available, and will directly be looked into their corresponding keymap.
Any other command will exit the incremental search.

Available commands in the `isearch` mode are:    

**Edition**    
`abort`    
`backward-delete-char`    
`backward-kill-word`    
`backward-kill-line`    
`unix-line-discard`    
`unix-word-rubout`    
`vi-unix-word-rubout`    
`clear-screen`    
`clear-display`    
`magic-space`    
`vi-movement-mode`    
`yank`    
`self-insert`    

**History**   
`accept-and-infer-next-history`    
`accept-line`    
`accept-and-hold`    
`operate-and-get-next`    
`history-incremental-search-forward`    
`history-incremental-search-backward`    
`forward-search-history`    
`reverse-search-history`    
`history-search-forward`    
`history-search-backward`    
`history-substring-search-forward`    
`history-substring-search-backward`    
`incremental-forward-search-history`    
`incremental-reverse-search-history`    

### Non-incremental Search

Some history search commands are non-incremental: they do not refresh completions
or the input line buffer (with a match) on each new character typed in the minibuffer.
They instead wait for the Enter key to be pressed to operate the search and insert.
Non-incremental search is not in itself a keymap, but works on the same principle as `isearch`. 

The subset of allowed commands is even smaller:     
`abort`    
`accept-line`    
`backward-delete-char`    
`backward-kill-word`    
`backward-kill-line`    
`unix-line-discard`    
`unix-word-rubout`    
`vi-unix-word-rubout`    
`self-insert`    
