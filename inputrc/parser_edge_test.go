package inputrc

import "testing"

func TestParseSetWithoutValue(t *testing.T) {
	// A "set" directive with no value must be reported as a parse error at
	// most; it must never panic while reading the (empty) value.
	for _, line := range []string{"set   \n", "set \t\n", "set  # comment\n"} {
		cfg := NewDefaultConfig()

		_ = ParseBytes([]byte(line), cfg, WithStrict(false))
	}
}
