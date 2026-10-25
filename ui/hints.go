package ui

import (
	"strings"
	"sync"

	"github.com/malivvan/readline/color"
	"github.com/malivvan/readline/core"
	"github.com/malivvan/readline/strutil"
	"github.com/malivvan/readline/term"
)

type Hint struct {
	mu         sync.RWMutex
	text       []rune
	persistent []rune
	provided   []rune
	transient  []rune
	cleanup    bool
	temp       bool
	set        bool

	provider func(line []rune, cursor int) []rune

	refresh func()
}

func CoordinatesHint(hint *Hint) int {
	hint.mu.RLock()
	defer hint.mu.RUnlock()

	width := term.GetWidth()
	if width <= 0 {
		width = 80
	}

	usedY := 0

	for _, lane := range [][]rune{hint.persistent, hint.provided, hint.transient, hint.text} {
		if len(lane) == 0 {
			continue
		}

		for _, sub := range strings.Split(string(lane), "\n") {
			length := strutil.RealLength(sub)

			rows := length / width
			if length%width != 0 || rows == 0 {
				rows++
			}

			usedY += rows
		}
	}

	return usedY
}

func DisplayHint(hint *Hint) {
	hint.mu.Lock()
	defer hint.mu.Unlock()

	if hint.temp && hint.set {
		hint.set = false
	} else if hint.temp {
		hint.text = make([]rune, 0)
		hint.temp = false
		hint.set = false
	}

	if hint.empty() {
		if hint.cleanup {
			term.WriteString(term.ClearLineAfter)
		}

		hint.cleanup = false

		return
	}

	text := hint.renderLocked()

	if strutil.RealLength(text) == 0 {
		return
	}

	text += term.ClearLineAfter + color.Reset

	if len(text) > 0 {
		term.WriteString(text)
	}
}

func NewHint(keys *core.Keys) *Hint {
	hint := &Hint{}
	if keys != nil {
		hint.refresh = keys.RequestRefresh
	}

	return hint
}

func (h *Hint) ClearTransient() {
	h.mu.Lock()
	h.cleanup = h.cleanup || len(h.transient) > 0
	h.transient = make([]rune, 0)
	refresh := h.refresh
	h.mu.Unlock()

	if refresh != nil {
		refresh()
	}
}

func (h *Hint) Len() int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	return len(h.text)
}

func (h *Hint) Persist(hint string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.persistent = []rune(hint)
}

func (h *Hint) Reset() {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.text = make([]rune, 0)
	h.temp = false
	h.set = false
}

func (h *Hint) ResetPersist() {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.cleanup = len(h.persistent) > 0
	h.persistent = make([]rune, 0)
}

func (h *Hint) Set(hint string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.text = []rune(hint)
	h.set = true
}

func (h *Hint) SetProvided(hint string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.provided = []rune(hint)
}

func (h *Hint) SetProvider(provider func(line []rune, cursor int) []rune) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.provider = provider

	if provider == nil {
		h.provided = make([]rune, 0)
	}
}

func (h *Hint) SetRefreshFunc(refresh func()) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.refresh = refresh
}

func (h *Hint) SetTemporary(hint string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.text = []rune(hint)
	h.set = true
	h.temp = true
}

func (h *Hint) SetTransient(hint string) {
	h.mu.Lock()
	h.transient = []rune(hint)
	refresh := h.refresh
	h.mu.Unlock()

	if refresh != nil {
		refresh()
	}
}

func (h *Hint) Text() string {
	h.mu.RLock()
	defer h.mu.RUnlock()

	return string(h.text)
}

func (h *Hint) UpdateProvided(line []rune, cursor int) {
	h.mu.RLock()
	provider := h.provider
	h.mu.RUnlock()

	if provider == nil {
		return
	}

	h.SetProvided(string(provider(line, cursor)))
}

func (h *Hint) empty() bool {
	return len(h.text) == 0 &&
		len(h.persistent) == 0 &&
		len(h.provided) == 0 &&
		len(h.transient) == 0
}

func (h *Hint) renderLocked() (text string) {

	for _, lane := range [][]rune{h.persistent, h.provided, h.transient, h.text} {
		if len(lane) > 0 {
			text += string(lane) + term.NewlineReturn
		}
	}

	if strutil.RealLength(text) == 0 {
		return
	}

	text = strings.ReplaceAll(text, term.NewlineReturn, term.ClearLineAfter+term.NewlineReturn)

	return text
}
