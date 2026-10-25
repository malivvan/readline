//go:build windows
// +build windows

package editor

import "errors"

func (reg *Buffers) EditBuffer(buf []rune, filename, filetype string, emacs bool) ([]rune, error) {
	return buf, errors.New("not currently supported on Windows")
}
