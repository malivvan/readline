package completion

import (
	"sort"
	"strings"
)

type Candidate struct {
	Value       string
	Display     string
	Description string
	Style       string
	Tag         string

	displayLen int
	descLen    int
}

type Completer func() Values

type RawValues []Candidate

type Values struct {
	values   RawValues
	Messages Messages
	NoSpace  SuffixMatcher
	Usage    string
	ListLong map[string]bool
	NoSort   map[string]bool
	ListSep  map[string]string
	Pad      map[string]bool
	Escapes  map[string]bool

	PREFIX string

	SUFFIX string
}

func AddRaw(values []Candidate) Values {
	return Values{
		values:   RawValues(values),
		ListLong: make(map[string]bool),
		NoSort:   make(map[string]bool),
		ListSep:  make(map[string]string),
		Pad:      make(map[string]bool),
	}
}

func (c RawValues) EachTag(tagF func(tag string, values RawValues)) {
	tags := make([]string, 0)
	tagGroups := make(map[string]RawValues)

	for _, val := range c {
		if _, exists := tagGroups[val.Tag]; !exists {
			tagGroups[val.Tag] = make(RawValues, 0)

			tags = append(tags, val.Tag)
		}

		tagGroups[val.Tag] = append(tagGroups[val.Tag], val)
	}

	for _, tag := range tags {
		tagF(tag, tagGroups[tag])
	}
}

func (c RawValues) Filter(values ...string) RawValues {
	toremove := make(map[string]bool)
	for _, v := range values {
		toremove[v] = true
	}

	filtered := make([]Candidate, 0)

	for _, rawValue := range c {
		if _, ok := toremove[rawValue.Value]; !ok {
			filtered = append(filtered, rawValue)
		}
	}

	return filtered
}

func (c RawValues) FilterPrefix(prefix string, matchCase bool) RawValues {
	if prefix == "" {
		return c
	}

	filtered := make(RawValues, 0)

	if !matchCase {
		prefix = strings.ToLower(prefix)
	}

	for _, raw := range c {
		val := raw.Value

		if !matchCase {
			val = strings.ToLower(val)
		}

		if strings.HasPrefix(val, prefix) {
			filtered = append(filtered, raw)
		}
	}

	return filtered
}

func (c RawValues) Len() int { return len(c) }

func (c RawValues) Less(i, j int) bool {
	return strings.ToLower(c[i].Value) < strings.ToLower(c[j].Value)
}

func (c RawValues) Swap(i, j int) { c[i], c[j] = c[j], c[i] }

func (c RawValues) sortStable() {
	type keyed struct {
		key   string
		value Candidate
	}

	pairs := make([]keyed, len(c))
	for i, val := range c {
		pairs[i] = keyed{key: strings.ToLower(val.Value), value: val}
	}

	sort.SliceStable(pairs, func(i, j int) bool {
		return pairs[i].key < pairs[j].key
	})

	for i := range pairs {
		c[i] = pairs[i].value
	}
}

func (c *Values) Merge(other Values) {
	if other.Usage != "" {
		c.Usage = other.Usage
	}

	c.NoSpace.Merge(other.NoSpace)
	c.Messages.Merge(other.Messages)

	if len(other.ListLong) > 0 && c.ListLong == nil {
		c.ListLong = make(map[string]bool, len(other.ListLong))
	}

	for tag := range other.ListLong {
		if _, found := c.ListLong[tag]; !found {
			c.ListLong[tag] = true
		}
	}
}
