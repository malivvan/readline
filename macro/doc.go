// Package macro implements keyboard macro recording and replay.
//
// Keys are recorded into escaped form (see inputrc.EscapeMacro) and replayed by
// feeding them back to the key reader. The Emacs commands start-kbd-macro,
// end-kbd-macro and call-last-kbd-macro drive the engine, and Vi mode records
// into registers with q<register>/@<register>.
package macro
