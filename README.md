<div align="center">
  <br> <h1> Readline </h1>
</div>

<!-- Badges -->
<p align="center">
  <a href="https://github.com/malivvan/readline/actions/workflows/go.yml">
    <img src="https://github.com/malivvan/readline/actions/workflows/go.yml/badge.svg?branch=master"
      alt="Github Actions (workflows)" />
  </a>

  <a href="https://github.com/malivvan/readline">
    <img src="https://img.shields.io/github/go-mod/go-version/malivvan/readline.svg"
      alt="Go module version" />
  </a>

  <a href="https://pkg.go.dev/github.com/malivvan/readline">
    <img src="https://img.shields.io/badge/godoc-reference-blue.svg"
      alt="GoDoc reference" />
  </a>

  <a href="https://goreportcard.com/report/github.com/malivvan/readline">
    <img src="https://goreportcard.com/badge/github.com/malivvan/readline"
      alt="Go Report Card" />
  </a>

  <a href="https://codecov.io/gh/malivvan/readline">
    <img src="https://codecov.io/gh/malivvan/readline/branch/main/graph/badge.svg"
      alt="codecov" />
  </a>

  <a href="https://opensource.org/licenses/BSD-3-Clause">
    <img src="https://img.shields.io/badge/License-BSD_3--Clause-blue.svg"
      alt="License: BSD-3" />
  </a>
</p>

This library is a modern, pure Go `readline` shell implementation, with full `.inputrc` and legacy
readline command/option support, and extended with various commands, options and tools commonly
found in modern shells. Its architecture and completion system is heavily inspired from Z-Shell.

## Features

### Core

- Pure Go, almost-only standard library
- Cross-platform (Linux / MacOS / Windows)
- Full `.inputrc` parsing, with options and the implemented legacy commands
- Extensive test suite and almost full coverage of core code
- [Extended list](docs/KEYMAPS-&-COMMANDS.md) of additional commands/options (edition/completion/history)
- Complete [multiline edition/movement support](docs/MULTILINE.md)
- Command-line edition in `$EDITOR`/`$VISUAL` support
- [Programmable API](docs/PROGRAMMABLE-COMMANDS.md), with failure-safe access to core components
- Support for an [arbitrary number of history sources](docs/HISTORY-SOURCES.md)

### Emacs / Standard

- Native Emacs commands
- Emacs-style [macro engine](docs/MACROS.md#emacs-mode) (not working across multiple calls)
- Keywords [switching](docs/KEYMAPS-&-COMMANDS.md#modifying-text) (operators, booleans, hex/binary/digit) with iterations
- Command/mode cursor status indicator
- Complete undo/redo history
- Command status/arg/iterations hint display

### Vim

- Near-native Vim mode
- Vim [text objects](docs/KEYMAPS-&-COMMANDS.md#text-objects) (code blocks, words/blank/shellwords)
- Extended surround select/change/add functionality, with highlighting
- Vim Visual/Operator pending mode & cursor styles indications
- Vim Insert and Replace (once/many)
- All Vim registers, with completion support
- [Vim-style](docs/MACROS.md#vim-mode) macro recording (`q<a>`) and invocation (`@<a>`)

### Interface

- Support for PS1/PS2/RPROMPT/transient/tooltip [prompts](docs/PROMPTS.md) (compatible with [oh-my-posh](https://github.com/JanDeDobbeleer/oh-my-posh))
- Extended completion system, [keymap-based and configurable](docs/KEYMAPS-&-COMMANDS.md#completion), easy to populate & use
- Multiple completion display styles, with color support.
- Completion & History incremental search system & highlighting (fuzzy-search).
- Automatic & context-aware suffix removal for efficient flags/path/list completion.
- Optional asynchronous autocomplete
- Builtin & programmable [syntax highlighting](docs/SYNTAX-HIGHLIGHTING.md)

## Documentation

Every feature of the library has a small runnable example in [`cmd/`](cmd/README.md). To get a grasp
of the functionality provided by readline and its default configuration, build and start one of them:

```console
go run ./cmd/basics
```

The documentation is available in the [repository docs](docs/HOME.md), for both users and developers.

## Installation

```console
go get github.com/malivvan/readline
```

```go
package main

import (
	"fmt"

	"github.com/malivvan/readline"
)

func main() {
	rl := readline.NewShell()
	for {
		line, err := rl.Readline()
		if err != nil {
			return
		}
		fmt.Printf("accepted: %s\n", line)
	}
}
```

## Showcases

<details>
  <summary>- Emacs edition</summary>
 <dd><em>(This extract is quite a pity, because its author is not using Emacs and does not know many of its shortcuts)</em></dd>
<img src="https://github.com/malivvan/readline/blob/assets/emacs.gif"/>
</details>
<details>
  <summary>- Vim edition</summary>
<img src="https://github.com/malivvan/readline/blob/assets/vim.gif"/>
</details>
<details>
  <summary>- Undo/redo line history </summary>
<img src="https://github.com/malivvan/readline/blob/assets/undo.gif"/>
</details>
<details>
  <summary>- Keyword switching & selection </summary>
 <dd><em>Switching various keywords</em></dd>
<img src="https://github.com/malivvan/readline/blob/assets/switch-keywords.gif"/>
 <dd><em>Using regexp-based selection to grab parts of words (here, URL components)</em></dd>
<img src="https://github.com/malivvan/readline/blob/assets/select-keywords.gif"/>
</details>
<details>
  <summary>- Vim selection & movements (basic) </summary>
<img src="https://github.com/malivvan/readline/blob/assets/vim-selection.gif"/>
</details>
<details>
  <summary>- Vim surround (selection and change) </summary>
 <dd><em>Selecting/adding/changing surround regions</em></dd>
<img src="https://github.com/malivvan/readline/blob/assets/vim-surround.gif"/>
 <dd><em>Surround and change in shellwords, matching brackets, etc.</em></dd>
<img src="https://github.com/malivvan/readline/blob/assets/vim-surround-2.gif"/>
</details>
<details>
  <summary>- Vim registers (with completion) </summary>
<img src="https://github.com/malivvan/readline/blob/assets/registers.gif"/>
</details>
<details>
  <summary>- History movements/completion/use/search </summary>
 <dd><em>History movement, completion and some other other widgets</em></dd>
<img src="https://github.com/malivvan/readline/blob/assets/history.gif"/>
</details>
<details>
  <summary>- Completion </summary>
 <dd><em>Classic mode & incremental search mode</em></dd>
<img src="https://github.com/malivvan/readline/blob/assets/completion.gif"/>
 <dd><em>Suffix-autoremoval </em></dd>
<img src="https://github.com/malivvan/readline/blob/assets/suffix-autoremoval.gif"/>
</details>
<details>
  <summary>- Prompts </summary>
<img src="https://github.com/malivvan/readline/blob/assets/prompts.gif"/>
</details>
<details>
  <summary>- Logging </summary>
<img src="https://github.com/malivvan/readline/blob/assets/logging.gif"/>
</details>
<details>
  <summary>- Inputrc init file reload </summary>
<img src="https://github.com/malivvan/readline/blob/assets/config-reload.gif"/>
</details>
<details>
  <summary>- Multiline edition </summary>
<img src="https://github.com/malivvan/readline/blob/assets/multiline.gif"/>
</details>
<details>
  <summary>- Macros </summary>
 <dd><em>Emacs</em></dd>
<img src="https://github.com/malivvan/readline/blob/assets/emacs-macros.gif"/>
 <dd><em>Vim</em></dd>
<img src="https://github.com/malivvan/readline/blob/assets/vim-macros.gif"/>
</details>

## Status

**1.0.0** — feature-complete, with a solid test suite.

Key dispatch/flushing and meta-key handling follow 30 years of legacy readline
behaviour across a wide range of terminal emulators; if you hit an issue, please
open an issue or a PR.

## Credits

- @kenshaw for his `.inputrc` parsing package, which brings much wider compatibility to this library.
- `chzyer/readline` for the Windows I/O code and everything related.
- Some of the Vim code is inspired or translated from [zsh-vi-mode](https://github.com/jeffreytse/zsh-vi-mode).
- [lmorg/readline](https://github.com/lmorg/readline), for the line tokenizers.
