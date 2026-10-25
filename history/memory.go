package history

type Source interface {
	Write(line string) (int, error)

	GetLine(pos int) (string, error)

	Len() int

	Dump() interface{}
}

type memory struct {
	items []string
}

var standardSourceName = "default history"

func NewInMemoryHistory() Source {
	return new(memory)
}

func (h *memory) Dump() interface{} {
	return h.items
}

func (h *memory) GetLine(i int) (string, error) {
	if i < 0 {
		return "", errNegativeIndex
	}

	if i >= len(h.items) {
		return "", errOutOfRangeIndex
	}

	return h.items[i], nil
}

func (h *memory) Len() int {
	return len(h.items)
}

func (h *memory) Write(s string) (int, error) {
	h.items = append(h.items, s)
	return len(h.items), nil
}
