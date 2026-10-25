// Package core holds the editor primitives shared by every other package: the
// input line, the cursor, text regions and selections, numeric arguments,
// tokenizers, and the key reader.
//
// The types are safe by construction: line and cursor mutations are clamped so
// that positions can never become invalid, and selections are expressed with
// line positions. They are meant to be driven by commands (see keymap and the
// root readline package), not to be used concurrently.
package core
