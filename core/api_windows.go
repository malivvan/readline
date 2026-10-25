//go:build windows
// +build windows

package core

import (
	"os"
	"reflect"
	"syscall"
	"unsafe"

	"github.com/malivvan/readline/term"
)

var (
	kernel = NewKernel()
	stdin  = uintptr(syscall.Stdin)
)

type Kernel struct {
	SetConsoleCursorPosition,
	SetConsoleTextAttribute,
	FillConsoleOutputCharacterW,
	FillConsoleOutputAttribute,
	ReadConsoleInputW,
	GetConsoleScreenBufferInfo,
	GetConsoleCursorInfo,
	GetStdHandle CallFunc
}

type (
	short int16
	word  uint16
	dword uint32
	wchar uint16
)

type _COORD struct {
	x short
	y short
}

const (
	EVENT_KEY                = 0x0001
	EVENT_MOUSE              = 0x0002
	EVENT_WINDOW_BUFFER_SIZE = 0x0004
	EVENT_MENU               = 0x0008
	EVENT_FOCUS              = 0x0010
)

type _KEY_EVENT_RECORD struct {
	bKeyDown          int32
	wRepeatCount      word
	wVirtualKeyCode   word
	wVirtualScanCode  word
	unicodeChar       wchar
	dwControlKeyState dword
}

type _INPUT_RECORD struct {
	EventType word
	Padding   uint16
	Event     [16]byte
}

type _CONSOLE_SCREEN_BUFFER_INFO struct {
	dwSize              _COORD
	dwCursorPosition    _COORD
	wAttributes         word
	srWindow            _SMALL_RECT
	dwMaximumWindowSize _COORD
}

type _SMALL_RECT struct {
	left   short
	top    short
	right  short
	bottom short
}

type _FOCUS_EVENT_RECORD struct {
	bSetFocus bool
}

type CallFunc func(u ...uintptr) error

func NewKernel() *Kernel {
	k := &Kernel{}
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	v := reflect.ValueOf(k).Elem()
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		name := t.Field(i).Name
		f := kernel32.NewProc(name)
		v.Field(i).Set(reflect.ValueOf(k.Wrap(f)))
	}
	return k
}

func (k *Kernel) Wrap(p *syscall.LazyProc) CallFunc {
	return func(args ...uintptr) error {
		r0, _, e1 := syscall.SyscallN(p.Addr(), args...)

		if int(r0) == 0 {
			if e1 != 0 {
				return error(e1)
			}
			return syscall.EINVAL
		}
		return nil
	}
}

func (k *Keys) GetCursorPos() (x, y int) {

	term.FlushBuffer()

	out, ok := k.outputWriter().(*os.File)
	if !ok {
		return -1, -1
	}

	t := new(_CONSOLE_SCREEN_BUFFER_INFO)
	if err := kernel.GetConsoleScreenBufferInfo(
		uintptr(out.Fd()),
		uintptr(unsafe.Pointer(t)),
	); err != nil {
		return -1, -1
	}

	// Console coordinates are 0-based buffer coordinates. The display engine
	// expects 1-based coordinates relative to the visible window.
	x = int(t.dwCursorPosition.x-t.srWindow.left) + 1
	y = int(t.dwCursorPosition.y-t.srWindow.top) + 1

	return
}
