package readline

import (
	"slices"
	"strings"

	"github.com/malivvan/readline/core"
)

func (rl *Shell) bracketedPasteBegin() {

	sequence := make([]byte, 0, 6)

	for {
		key, empty := core.PopKey(rl.Keys)
		if empty {
			core.WaitAvailableKeys(rl.Keys, rl.Config)

			if rl.Keys.IsEOF() || rl.Keys.ReadError() != nil {
				return
			}

			continue
		}

		sequence = append(sequence, key)

		if len(sequence) >= 6 && slices.Equal(sequence[len(sequence)-6:], []byte{'\x1b', '[', '2', '0', '1', '~'}) {
			break
		}
	}

	if len(sequence) > 6 {
		rl.insertPastedText(string(sequence[:len(sequence)-6]))
	}
}

func (rl *Shell) insertPastedText(text string) {

	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")

	if rl.PasteTransformer != nil {
		text = rl.PasteTransformer(text)
	}
	if text == "" {
		return
	}

	rl.cursor.InsertAt([]rune(text)...)
}

func (rl *Shell) skipCsiSequence() {
	rl.History.SkipSave()

	for {
		key, empty := core.PopKey(rl.Keys)
		if empty {
			return
		}

		if key < 0x20 || key >= 0x40 {
			return
		}
	}
}
