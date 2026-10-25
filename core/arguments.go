package core

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/malivvan/readline/color"
)

type Iterations struct {
	times   string
	active  bool
	pending bool
}

func ResetPostRunIterations(iter *Iterations) (hint string) {
	if iter.pending {
		hint = color.Dim + fmt.Sprintf("(arg: %s)", iter.times)
		iter.pending = false

		return
	}

	iter.active = false

	return
}

func (i *Iterations) Add(times string) {
	if times == "" {
		return
	}

	if _, err := strconv.Atoi(times); err != nil && times != "-" {
		return
	}

	i.active = true
	i.pending = true

	switch {
	case times == "-":
		i.times = times + i.times
	case strings.HasPrefix(times, "-"):
		i.times = "-" + i.times + strings.TrimPrefix(times, "-")
	default:
		i.times += times
	}
}

func (i *Iterations) Get() int {
	times, err := strconv.Atoi(i.times)

	switch {
	case err != nil && strings.HasPrefix(i.times, "-"):
		times = -1
	case err != nil && times == 0:
		times = 1
	case times == 0 && strings.HasPrefix(i.times, "-"):
		times = -1
	}

	if times == 0 {
		times++
	}

	i.times = ""

	return times
}

func (i *Iterations) IsPending() bool {
	return i.pending
}

func (i *Iterations) IsSet() bool {
	return i.active
}

func (i *Iterations) Reset() {
	i.times = ""
	i.active = false
	i.pending = false
}
