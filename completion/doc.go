// Package completion implements the completion engine: it turns the values
// returned by an application completer into candidate groups, renders them,
// and inserts the selection into the input buffer.
//
// Applications normally interact with it through the root readline package:
// the shell's Completer field returns a readline.Completions, which is a
// wrapper around Values. This package exposes the lower-level pieces used by
// the editor: the engine, candidate values and their rendering, autopairing,
// and the search modes (menu-select and incremental search).
package completion
