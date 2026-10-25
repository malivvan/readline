> **Runnable example:** [`cmd/basics`](../cmd/basics)


## Current shell libraries

There are several readline libraries available out there, some of them widely used 
and having passed the test of time, which is currently not the case of this one.
However, most of those libraries are somewhat restrained in their feature set, and
lack some of the features that are now commonly found in modern shells.

Some might lack an advanced completion system, hint/UI interfaces, cross-platform
powerful prompt support, or even might have taken architectural decisions preventing
them to fully and correctly implement Emacs/Vim editing modes.

## Original readline compatiblity

The library is also based on the universal `.inputrc` configuration file, used by the `readline` library.
Therefore, all readline commands and configuration options are supported. These have been augmented
with additional commands and options (Vim commands & modes, completion/history control, display settings).

## Aims and choices of this package

### Architecture & inspirations
This readline aims to solve most of the shortcomings found in other projects, providing a shell 
that offers a powerful and architecturally proven "keymap, widget-based" model, along with a modern 
user interface with support for modern prompt engines, an advanced completion system and support for
multiple history sources.

The shell from which readline draws its architectural choices, keymap model, and UI features
is the [Z-Shell](https://zsh.sourceforge.io/Doc/Release/The-Z-Shell-Manual.html#The-Z-Shell-Manual). While shell usage is also a personal choice, ZSH offers quite an advanced 
(maybe too much) UI and editing model. Its architecture is, in any case, quite well devised.

### Simpler APIs, same user experience
Where ZSH would go too far as far as most users are concerned with, this readline library
offers simplified APIs, while preserving the same versatility and user experience whenever
and wherever possible.

A good example of this is the completion system: the ZSH completion engine is just wizardry,
and no one sane would ever think of checking how deep and evolved it is. Most of it is either
overkill, and maintenance of completion scripts for a given program always incurs technical 
debt which, if not paid down, results in useless scripts. 
Additionally, people should spend the least amount of time writing stuff related to completion.
However, its a very powerful completion engine, including in its almost default setup.

The readline library offers a quite small but very powerful API for providing completions
to the shell. This API is outright copied after the [carapace library API](https://rsteube.github.io/carapace/carapace.html),
which is implemented by [`completion_values.go`](../completion_values.go). Its internals
are, however, mostly modeled after ZSH completion principles. The net result is a very
powerful completion user experience, with a very simple API to feed it with.

Other elements of the readline shell are also made in this optic: simplify the API for developers,
allowing them to concentrate on their application core, while preserving power and versatility of use.

