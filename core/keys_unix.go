//go:build unix

package core

import (
	"errors"
	"io"
	"os"
	"strconv"
	"sync"

	"golang.org/x/sys/unix"

	"github.com/malivvan/readline/term"
)

// keysOS holds the wake pipe used to interrupt a blocked read.
type keysOS struct {
	wakeMu    sync.Mutex
	wakeR     int
	wakeW     int
	wakeReady bool
}

// defaultInput is the reader used when no custom input is configured.
func defaultInput() io.Reader { return os.Stdin }

func (k *Keys) GetCursorPos() (x, y int) {
	fd, ok := k.stdinFd()
	if !ok || !term.IsTerminal(fd) {
		return -1, -1
	}

	// The query is written to the configured output stream. When that stream
	// is not a terminal (e.g. redirected to a file or a pipe), the terminal
	// never receives the query and therefore never answers it, so waiting for
	// the reply would block the shell forever. Disable probing in that case,
	// like the Windows implementation does for non-console output.
	output := k.outputWriter()
	outputFd, ok := output.(interface{ Fd() uintptr })
	if !ok || !term.IsTerminal(int(outputFd.Fd())) {
		return -1, -1
	}

	disable := func() (int, int) {
		_, _ = io.WriteString(k.stderrWriter(), "\r\ngetCursorPos() not supported by terminal emulator, disabling....\r\n")
		return -1, -1
	}

	var cursor []byte
	var match [][]string

	term.FlushBuffer()

	_, _ = io.WriteString(output, "\x1b[6n")

	for {
		switch {
		case k.waiting, k.reading:
			cursor = <-k.cursor
		default:
			buf := make([]byte, keyScanBufSize)

			read, err := k.stdin().Read(buf)
			if err != nil {
				return disable()
			}

			cursor = buf[:read]
		}

		if len(cursor) == 0 {
			return disable()
		}

		cursorSeq, remain := k.extractCursorPos(cursor)

		if len(remain) > 0 {
			k.mutex.RLock()
			k.buf = append(k.buf, remain...)
			k.mutex.RUnlock()
		}

		if len(cursorSeq) == 0 {
			continue
		}

		match = rxRcvCursorPos.FindAllStringSubmatch(string(cursorSeq), 1)

		break
	}

	y, err := strconv.Atoi(match[0][1])
	if err != nil {
		return disable()
	}

	x, err = strconv.Atoi(match[0][2])
	if err != nil {
		return disable()
	}

	return x, y
}

func (k *Keys) readInputFiltered() (keys []byte, err error) {

	buf := make([]byte, keyScanBufSize)

	read, err := k.readStdin(buf)
	if err != nil {
		return nil, err
	}

	cursor, keys := k.extractCursorPos(buf[:read])

	if len(cursor) > 0 {
		k.cursor <- cursor
	}

	return keys, nil
}

func (k *Keys) readStdin(buf []byte) (int, error) {
	k.wakeMu.Lock()
	ready := k.wakeReady
	wakeR := k.wakeR
	k.wakeMu.Unlock()

	fd, ok := k.stdinFd()
	if !ok || !ready {
		return k.stdin().Read(buf)
	}

	for {
		fds := []unix.PollFd{
			{Fd: int32(fd), Events: unix.POLLIN},
			{Fd: int32(wakeR), Events: unix.POLLIN},
		}

		if _, err := unix.Poll(fds, -1); err != nil {
			if errors.Is(err, unix.EINTR) {
				continue
			}

			return k.stdin().Read(buf)
		}

		if fds[0].Revents&(unix.POLLIN|unix.POLLHUP|unix.POLLERR) != 0 {
			return k.stdin().Read(buf)
		}

		if fds[1].Revents&unix.POLLIN != 0 {
			k.drainWake()
			return 0, errInputWake
		}
	}
}

func (k *Keys) InitWake() {
	k.CloseWake()

	var p [2]int
	if err := unix.Pipe(p[:]); err != nil {
		return
	}

	_ = unix.SetNonblock(p[0], true)
	_ = unix.SetNonblock(p[1], true)

	k.wakeMu.Lock()
	k.wakeR, k.wakeW = p[0], p[1]
	k.wakeReady = true
	k.wakeMu.Unlock()
}

func (k *Keys) CloseWake() {
	k.wakeMu.Lock()
	defer k.wakeMu.Unlock()

	if !k.wakeReady {
		return
	}

	_ = unix.Close(k.wakeR)
	_ = unix.Close(k.wakeW)
	k.wakeReady = false
}

func (k *Keys) RequestRefresh() {
	k.wakeMu.Lock()
	defer k.wakeMu.Unlock()

	if !k.wakeReady {
		return
	}

	_, _ = unix.Write(k.wakeW, []byte{0})
}

func (k *Keys) drainWake() {
	var scratch [16]byte

	for {
		n, err := unix.Read(k.wakeR, scratch[:])
		if n <= 0 || err != nil {
			return
		}
	}
}

func (k *Keys) stdinFd() (int, bool) {
	if f, ok := k.stdin().(interface{ Fd() uintptr }); ok {
		return int(f.Fd()), true
	}

	return 0, false
}
