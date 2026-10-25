package inputrc

//go:generate go run gen.go

import (
	"bytes"
	"errors"
	"io"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
)

type Error string

func Parse(r io.Reader, h Handler, opts ...Option) error {
	return New(opts...).Parse(r, h)
}

func ParseBytes(buf []byte, h Handler, opts ...Option) error {
	return New(opts...).Parse(bytes.NewReader(buf), h)
}

func ParseFile(name string, h Handler, opts ...Option) error {
	f, err := os.Open(name)
	if err != nil {
		return err
	}

	defer func() { _ = f.Close() }()

	return New(append(opts, WithName(name))...).Parse(f, h)
}

func UserDefault(u *user.User, cfg *Config, opts ...Option) error {

	var files []string
	if name := os.Getenv("INPUTRC"); name != "" {
		files = append(files, name)
	}

	if u != nil {
		name := ".inputrc"
		if runtime.GOOS == "windows" {
			name = "_inputrc"
		}

		files = append(files, filepath.Join(u.HomeDir, name))
	}

	if runtime.GOOS != "windows" {
		files = append(files, "/etc/inputrc")
	}

	for _, name := range files {
		buf, err := cfg.ReadFile(name)

		switch {
		case err != nil && errors.Is(err, os.ErrNotExist):
			continue
		case err != nil:
			return err
		}

		return ParseBytes(buf, cfg, append(opts, WithName(name))...)
	}

	return nil
}

func (err Error) Error() string {
	return string(err)
}
