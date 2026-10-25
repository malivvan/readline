//go:build windows
// +build windows

package core

import (
	"bytes"
	"testing"
)

func TestGetCursorPosNonConsoleWriter(t *testing.T) {
	keys := &Keys{}

	var out bytes.Buffer

	keys.SetOutput(&out)

	x, y := keys.GetCursorPos()
	if x != -1 || y != -1 {
		t.Fatalf("GetCursorPos() = (%d, %d), want (-1, -1) for a non-console writer", x, y)
	}
}
