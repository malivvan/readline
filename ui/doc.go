// Package ui implements the two informational areas of the interface: the
// prompt handlers and the hint area.
//
// Prompt lets applications register handlers for the primary, secondary, right,
// transient and tooltip prompts; each handler is called whenever the
// corresponding prompt is redisplayed.
//
// Hint is a set of independent lanes rendered top to bottom: a persistent
// status lane, a provider lane recomputed from the current line, a transient
// (asynchronous) lane, and the completion lane fed by usage and error
// messages.
package ui
