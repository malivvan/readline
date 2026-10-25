//go:build windows
// +build windows

package core

import (
	"io"
	"os"
	"unsafe"

	"github.com/malivvan/readline/inputrc"
	"github.com/malivvan/readline/term"
)

// keysOS holds the channel the console reader signals terminal resizes on.
type keysOS struct {
	resize chan bool
}

const (
	VK_CANCEL   = 0x03
	VK_BACK     = 0x08
	VK_TAB      = 0x09
	VK_RETURN   = 0x0D
	VK_SHIFT    = 0x10
	VK_CONTROL  = 0x11
	VK_MENU     = 0x12
	VK_ESCAPE   = 0x1B
	VK_LEFT     = 0x25
	VK_UP       = 0x26
	VK_RIGHT    = 0x27
	VK_DOWN     = 0x28
	VK_DELETE   = 0x2E
	VK_LSHIFT   = 0xA0
	VK_RSHIFT   = 0xA1
	VK_LCONTROL = 0xA2
	VK_RCONTROL = 0xA3
	VK_SNAPSHOT = 0x2C
	VK_INSERT   = 0x2D
	VK_HOME     = 0x24
	VK_END      = 0x23
	VK_PRIOR    = 0x21
	VK_NEXT     = 0x22
)

const (
	WINDOWS_RESIZE = 0x07
)

const (
	charTab       = 9
	charCtrlH     = 8
	charBackspace = 127
)

// defaultInput is the reader used when no custom input is configured.
func defaultInput() io.Reader { return os.Stdin }

// consoleReader translates Windows console input records into the ANSI
// sequences understood by the keymaps. It is shared by all Keys reading from
// the process console.
var consoleReader = makeRawReader()

// inputReader returns the reader keys should be read from. When the configured
// input is the process console, console input records are translated to ANSI
// sequences (arrow keys, Home/End, ...); any other stream is read as is.
func (k *Keys) inputReader() io.Reader {
	reader := k.stdin()

	if file, ok := reader.(*os.File); ok && file == os.Stdin && term.IsTerminal(int(file.Fd())) {
		return consoleReader
	}

	return reader
}

func GetTerminalResize(keys *Keys) <-chan bool {
	keys.resize = make(chan bool, 1)

	return keys.resize
}

func (k *Keys) readInputFiltered() (keys []byte, err error) {
	for {

		buf := make([]byte, keyScanBufSize)

		read, err := k.inputReader().Read(buf)
		if err != nil {

			return keys, err
		}

		input := buf[:read]

		if len(input) == 1 && input[0] == WINDOWS_RESIZE {
			if k.resize != nil {
				select {
				case k.resize <- true:
				default:
				}
			}

			continue
		}

		cursor, keys := k.extractCursorPos(input)

		if len(cursor) > 0 {
			k.cursor <- cursor
		}

		return keys, nil
	}
}

func (k *Keys) InitWake() {}

func (k *Keys) CloseWake() {}

func (k *Keys) RequestRefresh() {}

type rawReader struct {
	ctrlKey  bool
	altKey   bool
	shiftKey bool
}

func makeRawReader() *rawReader {
	r := new(rawReader)
	return r
}

func (r *rawReader) Read(buf []byte) (int, error) {
	ir := new(_INPUT_RECORD)
	var read int
	var err error

next:

	err = kernel.ReadConsoleInputW(stdin,
		uintptr(unsafe.Pointer(ir)),
		1,
		uintptr(unsafe.Pointer(&read)),
	)
	if err != nil {
		return 0, err
	}

	if ir.EventType == EVENT_FOCUS {
		ker := (*_FOCUS_EVENT_RECORD)(unsafe.Pointer(&ir.Event[0]))

		if !ker.bSetFocus && r.altKey {
			r.altKey = false
			goto next
		}
	}

	if ir.EventType == EVENT_WINDOW_BUFFER_SIZE {
		return r.write(buf, WINDOWS_RESIZE)
	}

	if ir.EventType != EVENT_KEY {
		goto next
	}

	ker := (*_KEY_EVENT_RECORD)(unsafe.Pointer(&ir.Event[0]))
	if ker.bKeyDown == 0 {
		if r.ctrlKey || r.altKey || r.shiftKey {
			switch ker.wVirtualKeyCode {
			case VK_RCONTROL, VK_LCONTROL, VK_CONTROL:
				r.ctrlKey = false
			case VK_MENU:
				r.altKey = false
			case VK_SHIFT, VK_LSHIFT, VK_RSHIFT:
				r.shiftKey = false
			}
		}
		goto next
	}

	if ker.unicodeChar == 0 {
		if modifiers, target := r.translateSeq(ker); target != 0 {
			return r.writeEsc(buf, append(modifiers, target)...)
		}
		goto next
	}

	char := rune(ker.unicodeChar)

	switch {
	case r.shiftKey && char == charTab:
		return r.writeEsc(buf, 91, 90)
	case r.ctrlKey && char == charBackspace:
		char = charCtrlH
	case !r.ctrlKey && char == charCtrlH:
		char = charBackspace
	case r.ctrlKey:
		char = inputrc.Encontrol(char)
	case r.altKey:
		char = inputrc.Enmeta(char)
	}

	return r.write(buf, char)
}

func (r *rawReader) Close() error {
	return nil
}

func (r *rawReader) writeEsc(b []byte, char ...rune) (int, error) {
	b[0] = byte(inputrc.Esc)
	n := copy(b[1:], []byte(string(char)))
	return n + 1, nil
}

func (r *rawReader) write(b []byte, char ...rune) (int, error) {
	n := copy(b, []byte(string(char)))
	return n, nil
}

func (r *rawReader) translateSeq(ker *_KEY_EVENT_RECORD) (modifiers []rune, target rune) {

	modifiers = append(modifiers, 91)

	switch ker.wVirtualKeyCode {
	case VK_RCONTROL, VK_LCONTROL, VK_CONTROL:
		r.ctrlKey = true
	case VK_MENU:
		r.altKey = true
	case VK_SHIFT, VK_LSHIFT, VK_RSHIFT:
		r.shiftKey = true
	}

	switch {
	case r.ctrlKey:
		modifiers = append(modifiers, 49, 59, 53)
	case r.altKey:
		modifiers = append(modifiers, 49, 59, 51)
	case r.shiftKey:
		modifiers = append(modifiers, 49, 59, 50)
	}

	changeModifiers := func(swap rune, pos int) {
		if len(modifiers) > pos-1 && pos > 0 {
			modifiers[pos] = swap
		} else {
			modifiers = append(modifiers, swap)
		}
	}

	switch ker.wVirtualKeyCode {

	case VK_LEFT:
		target = 68
	case VK_RIGHT:
		target = 67
	case VK_UP:
		target = 65
	case VK_DOWN:
		target = 66
	case VK_HOME:
		target = 72
	case VK_END:
		target = 70

	case VK_SNAPSHOT:
	case VK_INSERT:
		changeModifiers(50, 2)
		target = 126
	case VK_DELETE:
		changeModifiers(51, 2)
		target = 126
	case VK_PRIOR:
		changeModifiers(53, 2)
		target = 126
	case VK_NEXT:
		changeModifiers(54, 2)
		target = 126
	}

	return
}
