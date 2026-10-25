// Package inputrc parses .inputrc files and evaluates their conditional
// constructs, options and key bindings.
//
// The parser is shared with the legacy GNU readline syntax: it supports $if /
// $else / $endif blocks, $include, variable assignments and both command and
// macro binds. A parsed file is applied to a Config, the same object exposed by
// the shell as Shell.Config.
package inputrc
