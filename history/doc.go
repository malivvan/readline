// Package history implements the history sources of the shell.
//
// Sources holds an ordered list of named sources: a fresh shell starts with a
// single in-memory source, and applications can add file-backed sources (see
// NewSourceFromFile) or their own implementation of the Source interface. The
// shell can cycle through sources, and every history command (navigation,
// search, completion) operates on the current one.
package history
