// Package term wraps terminal input/output: raw mode, window size queries,
// cursor movement and the ANSI sequences used by the renderer.
//
// The package writes to a process-wide output writer, set by the shell from its
// Stdout stream (see SetOutput), so that every component renders to the same
// place.
package term
