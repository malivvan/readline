// Package keymap implements keymap selection and key dispatch.
//
// The engine holds one command per name, the bind tables of every keymap
// (emacs, emacs-standard, emacs-meta, emacs-ctlx, vi-insert, vi, vi-command,
// vi-move, vi-visual, vi-opp, isearch and menu-select), and tracks the active
// main and local keymaps. Incoming keys are matched against the current
// keymaps, and the matching command is executed by the editor loop.
//
// Applications add their own commands with Register and bind them through the
// inputrc Config exposed by the shell.
package keymap
