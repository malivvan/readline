//go:build windows
// +build windows

package term

import (
	"os"

	"golang.org/x/sys/windows"
)

type State struct {
	mode uint32
}

func IsTerminal(fd int) bool {
	var st uint32
	err := windows.GetConsoleMode(windows.Handle(fd), &st)
	return err == nil
}

func MakeRaw(fd int) (*State, error) {
	handle := windows.Handle(fd)

	var st uint32
	if err := windows.GetConsoleMode(handle, &st); err != nil {
		return nil, err
	}

	// Disable line editing, echo and Ctrl+C processing, and keep QuickEdit
	// from freezing console output while the line editor is running.
	raw := st &^ (windows.ENABLE_ECHO_INPUT | windows.ENABLE_PROCESSED_INPUT | windows.ENABLE_LINE_INPUT)
	raw &^= windows.ENABLE_QUICK_EDIT_MODE
	raw |= windows.ENABLE_EXTENDED_FLAGS

	if err := windows.SetConsoleMode(handle, raw); err != nil {
		return nil, err
	}

	state := &State{mode: st}
	enableVirtualTerminal()

	return state, nil
}

// enableVirtualTerminal makes sure the rendering stream interprets ANSI escape
// sequences. Windows 11 terminals enable this by default, but Windows 10
// consoles do not, and the line editor relies on ANSI rendering. The mode is
// left enabled: rendering also happens outside of Readline.
func enableVirtualTerminal() {
	out, ok := Output().(*os.File)
	if !ok {
		return
	}

	handle := windows.Handle(out.Fd())

	var mode uint32
	if err := windows.GetConsoleMode(handle, &mode); err != nil {
		return
	}

	if mode&windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING == 0 {
		_ = windows.SetConsoleMode(handle, mode|windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING)
	}
}

func GetState(fd int) (*State, error) {
	var st uint32
	if err := windows.GetConsoleMode(windows.Handle(fd), &st); err != nil {
		return nil, err
	}
	return &State{mode: st}, nil
}

func Restore(fd int, state *State) error {
	if state == nil {
		return nil
	}

	return windows.SetConsoleMode(windows.Handle(fd), state.mode)
}

func GetSize(fd int) (width, height int, err error) {
	var info windows.ConsoleScreenBufferInfo
	if err := windows.GetConsoleScreenBufferInfo(windows.Handle(fd), &info); err != nil {
		return 0, 0, err
	}

	// Terminal dimensions are the dimensions of the visible window, not of the
	// screen buffer: the buffer usually keeps a scrollback much taller than the
	// visible window. Both rectangle edges are inclusive.
	width = int(info.Window.Right-info.Window.Left) + 1
	height = int(info.Window.Bottom-info.Window.Top) + 1

	return width, height, nil
}
