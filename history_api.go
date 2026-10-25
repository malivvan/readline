package readline

import (
	"github.com/malivvan/readline/history"
)

type History = history.Source

var NewHistoryFromFile = history.NewSourceFromFile

var NewInMemoryHistory = history.NewInMemoryHistory
