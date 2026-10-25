//go:build !windows && !plan9

package editor

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"

	"github.com/kballard/go-shellquote"
)

var ErrStart = errors.New("failed to start editor")

func (reg *Buffers) EditBuffer(buf []rune, filename, filetype string, emacs bool) ([]rune, error) {
	name, err := persistToFile([]byte(string(buf)), filename)
	if err != nil {
		return buf, err
	}

	editor, args, err := systemEditorCommand(fetchSystemEditor(emacs))
	if err != nil {
		return buf, err
	}

	if filetype != "" && supportsFileType(editor) {
		args = append(args, "-c", "set filetype="+filetype)
	}

	args = append(args, name)

	cmd := exec.CommandContext(context.Background(), editor, args...)

	cmd.Stdin = reg.input()
	cmd.Stdout = reg.output()
	cmd.Stderr = reg.errorOutput()

	if err = cmd.Start(); err != nil {
		return buf, fmt.Errorf("%w: %s", ErrStart, err.Error())
	}

	if err = cmd.Wait(); err != nil {
		return buf, fmt.Errorf("%w: %s", ErrStart, err.Error())
	}

	b, err := installTempFile(name)

	return []rune(string(b)), err
}

// systemEditorCommand splits the configured editor into the binary and its
// arguments. VISUAL/EDITOR may legitimately contain flags (e.g. "code -w" or
// "open -W -n"); passing the whole value to exec.Command would fail with
// "executable file not found".
func systemEditorCommand(editor string) (string, []string, error) {
	fields, err := shellquote.Split(editor)
	if err != nil {
		return "", nil, fmt.Errorf("%w: %s", ErrStart, err.Error())
	}

	if len(fields) == 0 {
		return "", nil, fmt.Errorf("%w: empty editor command", ErrStart)
	}

	return fields[0], append([]string(nil), fields[1:]...), nil
}

// supportsFileType reports whether the editor understands vim's -c option used
// to set the buffer filetype. Editors like emacs and nano use different flags,
// so the option is only added for the vi family.
func supportsFileType(editor string) bool {
	switch filepath.Base(editor) {
	case "vi", "vim", "nvim", "view", "vimdiff", "gvim", "mvim":
		return true
	default:
		return false
	}
}
