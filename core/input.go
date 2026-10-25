package core

import (
	"errors"
	"io"
	"os"
	"regexp"
	"sync"

	"github.com/malivvan/readline/inputrc"
	"github.com/malivvan/readline/strutil"
	"github.com/rivo/uniseg"
)

type Keys struct {
	buf       []byte
	matched   []rune
	macroKeys []rune
	mustWait  bool
	waiting   bool
	reading   bool
	keysOnce  chan []byte
	cursor    chan []byte

	keysOS

	input  io.Reader
	output io.Writer
	stderr io.Writer

	eof     bool
	readErr error
	cfg     *inputrc.Config
	mutex   sync.RWMutex
}

const (
	keyScanBufSize = 1024
)

// SetInput sets the reader that keys are read from. A nil reader restores the
// default (os.Stdin). It must be called before Readline starts.
func (k *Keys) SetInput(input io.Reader) {
	k.input = input
}

// SetOutput sets the writer used for terminal queries issued while reading input.
// A nil writer restores the default (os.Stdout).
func (k *Keys) SetOutput(output io.Writer) {
	k.output = output
}

// SetErrorOutput sets the writer used for input-related diagnostics.
// A nil writer restores the default (os.Stderr).
func (k *Keys) SetErrorOutput(stderr io.Writer) {
	k.stderr = stderr
}

// stdin returns the reader keys are read from.
func (k *Keys) stdin() io.Reader {
	if k.input != nil {
		return k.input
	}

	return defaultInput()
}

// outputWriter returns the writer used for terminal queries.
func (k *Keys) outputWriter() io.Writer {
	if k.output != nil {
		return k.output
	}

	return os.Stdout
}

// stderrWriter returns the writer used for input diagnostics.
func (k *Keys) stderrWriter() io.Writer {
	if k.stderr != nil {
		return k.stderr
	}

	return os.Stderr
}

var errInputWake = errors.New("readline: input wake")

var rxRcvCursorPos = regexp.MustCompile(`\x1b\[([0-9]+);([0-9]+)R`)

func FlushUsed(keys *Keys) {
	keys.mutex.Lock()
	keys.matched = nil
	defer keys.mutex.Unlock()
}

func MacroKeys(keys *Keys) []rune {
	if keys.mustWait {
		return nil
	}

	return keys.matched
}

func MatchedKeys(keys *Keys, matched []byte, args ...byte) {
	if len(matched) > 0 {
		keys.matched = []rune(string(matched))
	}

	if len(args) > 0 {
		keys.buf = append(args, keys.buf...)
	}

	keys.mustWait = false
}

func MatchedPrefix(keys *Keys, prefix ...byte) {
	if len(prefix) == 0 {
		return
	}

	keys.mutex.Lock()
	defer keys.mutex.Unlock()

	keys.mustWait = len(keys.buf) == 0
	keys.buf = append(prefix, keys.buf...)
	keys.matched = []rune(string(prefix))
}

func PeekChar(keys *Keys) (char []byte, empty bool) {
	switch {
	case len(keys.buf) > 0:

		char, _, _, _ = uniseg.FirstGraphemeCluster(keys.buf, -1)
		return char, false

	case len(keys.macroKeys) > 0:

		return []byte(string(keys.macroKeys[0])), false

	default:
		return nil, true
	}
}

func PeekKey(keys *Keys) (key byte, empty bool) {
	switch {
	case len(keys.buf) > 0:
		key = keys.buf[0]
	case len(keys.macroKeys) > 0:
		key = byte(keys.macroKeys[0])
	default:
		return byte(0), true
	}

	return key, false
}

func PopChar(keys *Keys) (char []byte, empty bool) {
	switch {
	case len(keys.buf) > 0:
		char, keys.buf, _, _ = uniseg.FirstGraphemeCluster(keys.buf, -1)
	case len(keys.macroKeys) > 0:
		char = []byte(string(keys.macroKeys[0]))
		keys.macroKeys = keys.macroKeys[1:]
	default:
		return nil, true
	}

	return char, false
}

func PopForce(keys *Keys) (key byte, empty bool) {
	switch {
	case len(keys.buf) > 0:
		key = keys.buf[0]
		keys.buf = keys.buf[1:]
	case len(keys.macroKeys) > 0:
		key = byte(keys.macroKeys[0])
		keys.macroKeys = keys.macroKeys[1:]
	default:
		return byte(0), true
	}

	keys.mustWait = false

	return key, false
}

func PopKey(keys *Keys) (key byte, empty bool) {
	switch {
	case len(keys.buf) > 0:
		key = keys.buf[0]
		keys.buf = keys.buf[1:]
	case len(keys.macroKeys) > 0:
		key = byte(keys.macroKeys[0])
		keys.macroKeys = keys.macroKeys[1:]
	default:
		return byte(0), true
	}

	return key, false
}

func WaitAvailableKeys(keys *Keys, cfg *inputrc.Config) {
	keys.cfg = cfg

	if len(keys.buf) > 0 && !keys.mustWait {
		return
	}

	if len(keys.macroKeys) > 0 {
		return
	}

	keys.mutex.Lock()
	keys.waiting = true
	keys.cursor = make(chan []byte)
	keys.mutex.Unlock()

	defer func() {
		keys.mutex.Lock()
		keys.waiting = false
		keys.mutex.Unlock()
	}()

	for {

		keyBuf, err := keys.readInputFiltered()
		if err != nil {

			switch {
			case errors.Is(err, errInputWake):
			case errors.Is(err, io.EOF):
				keys.mutex.Lock()
				keys.eof = true
				keys.mutex.Unlock()
			default:
				keys.mutex.Lock()
				if keys.readErr == nil {
					keys.readErr = err
				}
				keys.mutex.Unlock()
			}

			return
		}

		if len(keyBuf) == 0 {
			continue
		}

		switch {
		case keys.reading:
			keys.keysOnce <- keyBuf
			continue

		default:

			if keys.cfg != nil && keys.cfg.GetBool("convert-meta") {
				keyBuf = []byte(strutil.ConvertMeta([]rune(string(keyBuf))))
			}

			keys.mutex.RLock()
			keys.buf = append(keys.buf, keyBuf...)
			keys.mutex.RUnlock()
		}

		return
	}
}

func (k *Keys) Caller() (keys []rune) {
	return k.matched
}

func (k *Keys) Empty() bool {
	k.mutex.RLock()
	defer k.mutex.RUnlock()

	return len(k.buf) == 0 && len(k.macroKeys) == 0
}

func (k *Keys) Feed(begin bool, keys ...rune) {
	if len(keys) == 0 {
		return
	}

	keyBuf := []rune(string(keys))

	k.mutex.Lock()
	defer k.mutex.Unlock()

	if begin {
		k.macroKeys = append(keyBuf, k.macroKeys...)
	} else {
		k.macroKeys = append(k.macroKeys, keyBuf...)
	}
}

func (k *Keys) IsEOF() bool {
	k.mutex.RLock()
	defer k.mutex.RUnlock()

	return k.eof
}

func (k *Keys) Pop() (key byte, empty bool) {
	switch {
	case len(k.buf) > 0:
		key = k.buf[0]
		k.buf = k.buf[1:]
	case len(k.macroKeys) > 0:
		key = byte(k.macroKeys[0])
		k.macroKeys = k.macroKeys[1:]
	default:
		return byte(0), true
	}

	k.matched = append(k.matched, rune(key))

	return key, false
}

func (k *Keys) ReadError() error {
	k.mutex.RLock()
	defer k.mutex.RUnlock()

	return k.readErr
}

func (k *Keys) ReadKey() (key rune, isAbort bool) {
	k.mutex.RLock()
	k.keysOnce = make(chan []byte)
	k.reading = true
	k.mutex.RUnlock()

	defer func() {
		k.mutex.RLock()
		k.reading = false
		k.mutex.RUnlock()
	}()

	switch {
	case len(k.macroKeys) > 0:
		key = k.macroKeys[0]
		k.macroKeys = k.macroKeys[1:]

	case k.waiting:
		buf := <-k.keysOnce
		if len(buf) == 0 {
			return rune(0), true
		}

		key = []rune(string(buf))[0]
	default:

		var buf []byte
		for len(buf) == 0 {
			b, err := k.readInputFiltered()
			if err != nil && !errors.Is(err, errInputWake) {
				break
			}

			buf = b
		}

		if len(buf) == 0 {
			return rune(0), true
		}

		key = []rune(string(buf))[0]
	}

	k.matched = append(k.matched, key)

	return key, key == inputrc.Esc
}

func (k *Keys) extractCursorPos(keys []byte) (cursor, remain []byte) {
	if !rxRcvCursorPos.Match(keys) {
		return cursor, keys
	}

	allCursors := rxRcvCursorPos.FindAll(keys, -1)
	cursor = allCursors[len(allCursors)-1]
	remain = rxRcvCursorPos.ReplaceAll(keys, nil)

	return
}
