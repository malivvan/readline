package strutil

import (
	"testing"
)

func TestGetQuotedWordStart(t *testing.T) {
	tests := []struct {
		name         string
		line         string
		wantUnclosed bool
		wantPos      int
	}{
		{
			name:         "Empty",
			line:         "",
			wantUnclosed: false,
			wantPos:      -1,
		},
		{
			name:         "Single word",
			line:         "word",
			wantUnclosed: false,
			wantPos:      -1,
		},
		{
			name:         "Unclosed double",
			line:         "\"word",
			wantUnclosed: true,
			wantPos:      0,
		},
		{
			name:         "Closed double",
			line:         "\"word\"",
			wantUnclosed: false,
			wantPos:      -1,
		},
		{
			name:         "Unclosed single",
			line:         "'word",
			wantUnclosed: true,
			wantPos:      0,
		},
		{
			name:         "Escaped quote in double",
			line:         "\"word \\\"",
			wantUnclosed: true,
			wantPos:      0,
		},
		{
			name:         "Escaped quote in single (literal)",
			line:         "'word \\'",
			wantUnclosed: false,
			wantPos:      -1,
		},
		{
			name:         "Nested quotes (single in double)",
			line:         "\"'\"",
			wantUnclosed: false,
			wantPos:      -1,
		},
		{
			name:         "Nested quotes (double in single)",
			line:         "'\"'",
			wantUnclosed: false,
			wantPos:      -1,
		},
		{
			name:         "Balanced nested",
			line:         "\"'hello'\"",
			wantUnclosed: false,
			wantPos:      -1,
		},
		{
			name:         "Multiple words unclosed",
			line:         "hello \"world",
			wantUnclosed: true,
			wantPos:      6,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			unclosed, pos := GetQuotedWordStart([]rune(tt.line))
			if unclosed != tt.wantUnclosed {
				t.Errorf("GetQuotedWordStart() unclosed = %v, want %v", unclosed, tt.wantUnclosed)
			}

			if unclosed && pos != tt.wantPos {
				t.Errorf("GetQuotedWordStart() pos = %v, want %v", pos, tt.wantPos)
			}
		})
	}
}

func TestSurroundType(t *testing.T) {
	tests := []struct {
		char     rune
		surround bool
		closer   bool
	}{
		{'{', true, false},
		{'}', true, true},
		{'(', true, false},
		{')', true, true},
		{'[', true, false},
		{']', true, true},
		{'<', true, false},
		{'>', true, true},
		{'"', true, true},
		{'\'', true, true},
		{'a', false, false},
	}

	for _, test := range tests {
		surround, closer := SurroundType(test.char)

		if surround != test.surround || closer != test.closer {
			t.Errorf("SurroundType(%q) = (%v, %v), want (%v, %v)", test.char, surround, closer, test.surround, test.closer)
		}
	}
}
