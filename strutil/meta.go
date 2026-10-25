package strutil

import (
	"github.com/malivvan/readline/inputrc"
)

func ConvertMeta(keys []rune) string {
	if len(keys) == 0 {
		return string(keys)
	}

	converted := make([]rune, 0, len(keys))

	for i := range keys {
		char := keys[i]

		if !inputrc.IsMeta(char) {
			converted = append(converted, char)
			continue
		}

		converted = append(converted, inputrc.Esc)
		converted = append(converted, inputrc.Demeta(char))
	}

	return string(converted)
}

func Quote(char rune) (res []rune, length int) {
	var inserted []rune

	if char == inputrc.Tab {
		inserted = append(inserted, char)
		return inserted, len(inserted)
	}

	switch {
	case inputrc.IsMeta(char):
		inserted = append(inserted, '^', '[')
		inserted = append(inserted, inputrc.Demeta(char))
	case inputrc.IsControl(char):
		inserted = append(inserted, '^')
		inserted = append(inserted, inputrc.Decontrol(char))
	default:
		inserted = []rune(inputrc.Unescape(string(char)))
	}

	return inserted, len(inserted)
}
