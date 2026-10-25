// Package color holds the ANSI SGR sequences used by the interface, along with
// helpers to format, strip, measure and truncate colored strings.
//
// The package variables are plain strings, so they can be embedded in any text
// rendered by the shell. DisableEffects empties them all for terminals without
// escape support.
package color
