//go:build !windows
// +build !windows

package editor

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

var (
	ErrNoTempDirectory = errors.New("could not identify the temp directory on this system")

	ErrWrite = errors.New("failed to write buffer to file")

	ErrCreate = errors.New("failed to create buffer file")

	ErrOpen = errors.New("failed to open buffer file")

	ErrRemove = errors.New("failed to remove buffer file")

	ErrRead = errors.New("failed to read buffer file")
)

func fetchSystemEditor(emacsDefault bool) (editor string) {
	editor = os.Getenv("VISUAL")
	if editor == "" {
		editor = os.Getenv("EDITOR")
	}

	if editor == "" {
		if emacsDefault {
			return "emacs"
		}
		return "vi"
	}

	return
}

func installTempFile(name string) ([]byte, error) {
	file, err := os.Open(name)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrOpen, err.Error())
	}

	buf, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrRead, err.Error())
	}

	if len(buf) > 0 && buf[len(buf)-1] == '\n' {
		buf = buf[:len(buf)-1]
	}

	if len(buf) > 0 && buf[len(buf)-1] == '\r' {
		buf = buf[:len(buf)-1]
	}

	if len(buf) > 0 && buf[len(buf)-1] == '\n' {
		buf = buf[:len(buf)-1]
	}

	if len(buf) > 0 && buf[len(buf)-1] == '\r' {
		buf = buf[:len(buf)-1]
	}

	if err = os.Remove(name); err != nil {
		return nil, fmt.Errorf("%w: %s", ErrRemove, err.Error())
	}

	return buf, nil
}

func persistToFile(buf []byte, filename string) (string, error) {
	var path string

	tmp := os.TempDir()
	if tmp == "" {
		return "", ErrNoTempDirectory
	}

	if filename == "" {
		fileID := strconv.Itoa(time.Now().Nanosecond()) + ":" + string(buf)

		h := sha256.New()

		_, err := h.Write([]byte(fileID))
		if err != nil {
			return "", err
		}

		name := "readline-" + hex.EncodeToString(h.Sum(nil)) + "-" + strconv.Itoa(os.Getpid())
		path = filepath.Join(tmp, name)
	} else {

		path = filepath.Join(tmp, filename)
	}

	file, err := os.Create(path)
	if err != nil {
		return "", fmt.Errorf("%w: %s", ErrCreate, err.Error())
	}

	defer func() { _ = file.Close() }()

	_, err = file.Write(buf)
	if err != nil {
		return "", fmt.Errorf("%w: %s", ErrWrite, err.Error())
	}

	return path, nil
}
