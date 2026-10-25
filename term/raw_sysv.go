//go:build solaris || aix
// +build solaris aix

package term

import (
	"syscall"

	"golang.org/x/sys/unix"
)

type State struct {
	state *unix.Termios
}

func IsTerminal(fd int) bool {
	_, err := unix.IoctlGetTermios(fd, unix.TCGETS)
	return err == nil
}

func MakeRaw(fd int) (*State, error) {
	oldTermiosPtr, err := unix.IoctlGetTermios(fd, unix.TCGETS)
	if err != nil {
		return nil, err
	}
	oldTermios := *oldTermiosPtr

	newTermios := oldTermios
	newTermios.Iflag &^= syscall.IGNBRK | syscall.BRKINT | syscall.PARMRK | syscall.ISTRIP | syscall.INLCR | syscall.IGNCR | syscall.ICRNL | syscall.IXON

	newTermios.Lflag &^= syscall.ECHO | syscall.ECHONL | syscall.ICANON | syscall.ISIG | syscall.IEXTEN
	newTermios.Cflag &^= syscall.CSIZE | syscall.PARENB
	newTermios.Cflag |= syscall.CS8
	newTermios.Cc[unix.VMIN] = 1
	newTermios.Cc[unix.VTIME] = 0

	if err := unix.IoctlSetTermios(fd, unix.TCSETS, &newTermios); err != nil {
		return nil, err
	}

	return &State{
		state: oldTermiosPtr,
	}, nil
}

func Restore(fd int, oldState *State) error {
	return unix.IoctlSetTermios(fd, unix.TCSETS, oldState.state)
}

func GetState(fd int) (*State, error) {
	oldTermiosPtr, err := unix.IoctlGetTermios(fd, unix.TCGETS)
	if err != nil {
		return nil, err
	}

	return &State{
		state: oldTermiosPtr,
	}, nil
}

func GetSize(fd int) (width, height int, err error) {
	ws, err := unix.IoctlGetWinsize(fd, unix.TIOCGWINSZ)
	if err != nil {
		return 0, 0, err
	}
	return int(ws.Col), int(ws.Row), nil
}
