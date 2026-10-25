package completion

import (
	"testing"

	"github.com/malivvan/readline/core"
)

func TestAutopairInsertOrJump(t *testing.T) {
	tests := []struct {
		name       string
		line       string
		cursor     int
		key        rune
		wantLine   string
		wantSkip   bool
		wantCursor int
	}{
		{
			name:       "Empty line, insert quote",
			line:       "",
			cursor:     0,
			key:        '"',
			wantLine:   "\"",
			wantSkip:   false,
			wantCursor: 0,
		},
		{
			name:       "Inside quote, type closing quote",
			line:       "\"foo",
			cursor:     4,
			key:        '"',
			wantLine:   "\"foo",
			wantSkip:   false,
			wantCursor: 4,
		},
		{
			name:       "Balanced quotes, type new quote",
			line:       "\"foo\"",
			cursor:     5,
			key:        '"',
			wantLine:   "\"foo\"\"",
			wantSkip:   false,
			wantCursor: 5,
		},
		{
			name:       "Escaped quote inside double, type quote",
			line:       "\"foo \\\"",
			cursor:     7,
			key:        '"',
			wantLine:   "\"foo \\\"",
			wantSkip:   false,
			wantCursor: 7,
		},
		{
			name:       "Jump over closing quote",
			line:       "\"foo\"",
			cursor:     4,
			key:        '"',
			wantLine:   "\"foo\"",
			wantSkip:   true,
			wantCursor: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			line := core.Line([]rune(tt.line))
			cur := core.NewCursor(&line)
			cur.Set(tt.cursor)

			skip := AutopairInsertOrJump(tt.key, &line, cur)

			if skip != tt.wantSkip {
				t.Errorf("AutopairInsertOrJump() skip = %v, want %v", skip, tt.wantSkip)
			}

			if string(line) != tt.wantLine {
				t.Errorf("AutopairInsertOrJump() line = %q, want %q", string(line), tt.wantLine)
			}

			if cur.Pos() != tt.wantCursor {
				t.Errorf("AutopairInsertOrJump() cursor = %v, want %v", cur.Pos(), tt.wantCursor)
			}
		})
	}
}
