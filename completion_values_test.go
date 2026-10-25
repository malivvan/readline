package readline

import "testing"

func TestCompletionsMerge(t *testing.T) {
	merged := CompleteValues("alpha", "beta").Merge(CompleteValues("beta", "gamma"))

	if got, want := len(merged.values), 3; got != want {
		t.Fatalf("merged value count = %d, want %d", got, want)
	}

	want := []string{"alpha", "beta", "gamma"}
	for index, value := range merged.values {
		if value.Value != want[index] {
			t.Errorf("merged value %d = %q, want %q", index, value.Value, want[index])
		}
	}
}

func TestCompletionsMergeNilMaps(t *testing.T) {
	// A zero container has no maps at all; merging containers that carry tags
	// must allocate them instead of panicking on a nil map write.
	merged := Completions{}.Merge(
		CompleteValues("alpha").DisplayList().NoSort().JustifyDescriptions().PreserveEscapes(),
		CompleteValues("beta").ListSeparator("|"),
	)

	if !merged.listLong["*"] {
		t.Error("DisplayList tag was not merged")
	}

	if !merged.noSort["*"] {
		t.Error("NoSort tag was not merged")
	}

	if !merged.pad["*"] {
		t.Error("JustifyDescriptions tag was not merged")
	}

	if !merged.escapes["*"] {
		t.Error("PreserveEscapes tag was not merged")
	}

	if merged.listSep["*"] != "|" {
		t.Errorf("ListSeparator = %q, want %q", merged.listSep["*"], "|")
	}
}
