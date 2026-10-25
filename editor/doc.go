// Package editor implements the editor buffers used by Vi mode, and the
// ability to hand the current input line over to $EDITOR/$VISUAL.
//
// Buffers stores the numbered and named Vi registers (the kill ring lives in
// the numbered ones) and provides completion for both, so that "vi-registers-
// complete" can list them.
package editor
