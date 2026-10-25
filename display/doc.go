// Package display renders the shell interface: the prompts, the input line and
// its highlights, the hint area, the completion menu, ghost text, and the
// transient prompt.
//
// The engine never writes to the terminal from a background goroutine: workers
// only update their own state and wake the editor, and the actual repaint
// happens on the goroutine running the Readline loop.
package display
