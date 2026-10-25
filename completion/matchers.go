package completion

import (
	"sort"
	"strings"
)

type SuffixMatcher struct {
	string
	pos int
}

type byRune []rune

func (sm *SuffixMatcher) Add(suffixes ...rune) {
	if strings.Contains(sm.string, "*") || strings.Contains(string(suffixes), "*") {
		sm.string = "*"

		return
	}

	unique := []rune(sm.string)

	for _, r := range suffixes {
		if !strings.Contains(sm.string, string(r)) {
			unique = append(unique, r)
		}
	}

	sort.Sort(byRune(unique))
	sm.string = string(unique)
}

func (sm SuffixMatcher) Matches(s string) bool {
	for _, r := range sm.string {
		if r == '*' || strings.HasSuffix(s, string(r)) {
			return true
		}
	}

	return false
}

func (sm *SuffixMatcher) Merge(other SuffixMatcher) {
	for _, r := range other.string {
		sm.Add(r)
	}
}

func (r byRune) Len() int { return len(r) }

func (r byRune) Less(i, j int) bool { return r[i] < r[j] }

func (r byRune) Swap(i, j int) { r[i], r[j] = r[j], r[i] }
