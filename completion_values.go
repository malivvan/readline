package readline

import (
	"fmt"

	"github.com/malivvan/readline/completion"
)

type Completion = completion.Candidate

type Completions struct {
	values   completion.RawValues
	messages completion.Messages
	noSpace  completion.SuffixMatcher
	usage    string
	listLong map[string]bool
	noSort   map[string]bool
	listSep  map[string]string
	pad      map[string]bool
	escapes  map[string]bool

	PREFIX string

	SUFFIX string
}

func CompleteMessage(msg string, args ...any) Completions {
	comps := Completions{}

	if len(args) > 0 {
		msg = fmt.Sprintf(msg, args...)
	}

	comps.messages.Add(msg)

	return comps
}

func CompleteRaw(values []Completion) Completions {
	return Completions{values: completion.RawValues(values)}
}

func CompleteStyledValues(values ...string) Completions {
	if length := len(values); length%2 != 0 {
		return CompleteMessage("invalid amount of arguments [CompleteStyledValues]: %v", length)
	}

	vals := make([]Completion, 0, len(values)/2)
	for i := 0; i < len(values); i += 2 {
		vals = append(vals, Completion{
			Value:       values[i],
			Display:     values[i],
			Description: "",
			Style:       values[i+1],
		})
	}

	return Completions{values: vals}
}

func CompleteStyledValuesDescribed(values ...string) Completions {
	if length := len(values); length%3 != 0 {
		return CompleteMessage("invalid amount of arguments [CompleteStyledValuesDescribed]: %v", length)
	}

	vals := make([]Completion, 0, len(values)/3)
	for i := 0; i < len(values); i += 3 {
		vals = append(vals, Completion{
			Value:       values[i],
			Display:     values[i],
			Description: values[i+1],
			Style:       values[i+2],
		})
	}

	return Completions{values: vals}
}

func CompleteValues(values ...string) Completions {
	vals := make([]Completion, 0, len(values))
	for _, val := range values {
		vals = append(vals, Completion{
			Value:       val,
			Display:     val,
			Description: "",
		})
	}

	return Completions{values: vals}
}

func CompleteValuesDescribed(values ...string) Completions {
	if length := len(values); length%2 != 0 {
		return CompleteMessage("invalid amount of arguments [CompleteValuesDescribed]: %v", length)
	}

	vals := make([]Completion, 0, len(values)/2)
	for i := 0; i < len(values); i += 2 {
		vals = append(vals, Completion{
			Value:       values[i],
			Display:     values[i],
			Description: values[i+1],
		})
	}

	return Completions{values: vals}
}

func (c Completions) DisplayList(tags ...string) Completions {
	if c.listLong == nil {
		c.listLong = make(map[string]bool)
	}

	if len(tags) == 0 {
		c.listLong["*"] = true
	}

	for _, tag := range tags {
		c.listLong[tag] = true
	}

	return c
}

func (c *Completions) EachValue(tagF func(comp Completion) Completion) {
	for index, v := range c.values {
		c.values[index] = tagF(v)
	}
}

func (c Completions) Filter(values []string) Completions {
	c.values = c.values.Filter(values...)
	return c
}

func (c Completions) JustifyDescriptions(tags ...string) Completions {
	if c.pad == nil {
		c.pad = make(map[string]bool)
	}

	if len(tags) == 0 {
		c.pad["*"] = true
	}

	for _, tag := range tags {
		c.pad[tag] = true
	}

	return c
}

func (c Completions) ListSeparator(seps ...string) Completions {
	if c.listSep == nil {
		c.listSep = make(map[string]string)
	}

	if length := len(seps); len(seps) > 1 && length%2 != 0 {
		return CompleteMessage("invalid amount of arguments (ListSeparator): %v", length)
	}

	if len(seps) == 1 {
		if len(c.listSep) == 0 {
			c.listSep["*"] = seps[0]
		} else {
			for tag := range c.listSep {
				c.listSep[tag] = seps[0]
			}
		}
	} else {
		for i := 0; i < len(seps); i += 2 {
			c.listSep[seps[i]] = seps[i+1]
		}
	}

	return c
}

func (c Completions) Merge(others ...Completions) Completions {
	uniqueRawValues := make(map[string]Completion)
	order := make([]string, 0)

	for _, other := range append([]Completions{c}, others...) {
		for _, comp := range other.values {
			if _, found := uniqueRawValues[comp.Value]; !found {
				order = append(order, comp.Value)
			}

			uniqueRawValues[comp.Value] = comp
		}
	}

	for _, other := range others {
		c.merge(other)
	}

	rawValues := make([]Completion, 0, len(uniqueRawValues))
	for _, value := range order {
		rawValues = append(rawValues, uniqueRawValues[value])
	}

	c.values = rawValues

	return c
}

func (c Completions) NoSort(tags ...string) Completions {
	if c.noSort == nil {
		c.noSort = make(map[string]bool)
	}

	if len(tags) == 0 {
		c.noSort["*"] = true
	}

	for _, tag := range tags {
		c.noSort[tag] = true
	}

	return c
}

func (c Completions) NoSpace(suffixes ...rune) Completions {
	if len(suffixes) == 0 {
		c.noSpace.Add('*')
	}

	c.noSpace.Add(suffixes...)

	return c
}

func (c Completions) Prefix(prefix string) Completions {
	for index, val := range c.values {
		c.values[index].Value = prefix + val.Value
	}

	return c
}

func (c Completions) PreserveEscapes(tags ...string) Completions {
	if c.escapes == nil {
		c.escapes = make(map[string]bool)
	}

	if len(tags) == 0 {
		c.escapes["*"] = true
	}

	for _, tag := range tags {
		c.escapes[tag] = true
	}

	return c
}

func (c Completions) Style(style string) Completions {
	return c.StyleF(func(_ string) string {
		return style
	})
}

func (c Completions) StyleF(f func(s string) string) Completions {
	for index, v := range c.values {
		c.values[index].Style = f(v.Value)
	}

	return c
}

func (c Completions) StyleR(style *string) Completions {
	if style != nil {
		return c.Style(*style)
	}

	return c
}

func (c Completions) Suffix(suffix string) Completions {
	for index, val := range c.values {
		c.values[index].Value = val.Value + suffix
	}

	return c
}

func (c Completions) Suppress(expr ...string) Completions {
	if err := c.messages.Suppress(expr...); err != nil {
		return CompleteMessage(err.Error())
	}

	return c
}

func (c Completions) Tag(tag string) Completions {
	return c.TagF(func(_ string) string {
		return tag
	})
}

func (c Completions) TagF(f func(value string) string) Completions {
	for index, v := range c.values {
		c.values[index].Tag = f(v.Value)
	}

	return c
}

func (c Completions) Usage(usage string, args ...any) Completions {
	return c.UsageF(func() string {
		return fmt.Sprintf(usage, args...)
	})
}

func (c Completions) UsageF(f func() string) Completions {
	if usage := f(); usage != "" {
		c.usage = usage
	}

	return c
}

func (c *Completions) convert() completion.Values {
	comps := completion.AddRaw(c.values)

	comps.Messages = c.messages
	comps.NoSpace = c.noSpace
	comps.Usage = c.usage
	comps.ListLong = c.listLong
	comps.NoSort = c.noSort
	comps.ListSep = c.listSep
	comps.Pad = c.pad
	comps.Escapes = c.escapes

	comps.PREFIX = c.PREFIX
	comps.SUFFIX = c.SUFFIX

	return comps
}

func (c *Completions) merge(other Completions) {
	if other.usage != "" {
		c.usage = other.usage
	}

	c.noSpace.Merge(other.noSpace)
	c.messages.Merge(other.messages)

	c.mergeTags(&c.listLong, other.listLong, true)
	c.mergeTags(&c.noSort, other.noSort, true)
	c.mergeTags(&c.pad, other.pad, true)
	c.mergeTags(&c.escapes, other.escapes, true)

	if len(other.listSep) > 0 {
		if c.listSep == nil {
			c.listSep = make(map[string]string, len(other.listSep))
		}

		for tag, sep := range other.listSep {
			if _, found := c.listSep[tag]; !found {
				c.listSep[tag] = sep
			}
		}
	}
}

// mergeTags copies the tags of another container into dst, allocating it when
// the other container has tags and dst does not. Values already set are kept.
func (c *Completions) mergeTags(dst *map[string]bool, other map[string]bool, value bool) {
	if len(other) == 0 {
		return
	}

	if *dst == nil {
		*dst = make(map[string]bool, len(other))
	}

	for tag := range other {
		if _, found := (*dst)[tag]; !found {
			(*dst)[tag] = value
		}
	}
}
