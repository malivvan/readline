package strutil

import (
	"bytes"
	"errors"
	"regexp"
	"strings"
	"unicode/utf8"
)

var NewlineMatcher = regexp.MustCompile(`\r\n`)

var (
	errUnterminatedSingleQuote = errors.New("unterminated single-quoted string")
	errUnterminatedDoubleQuote = errors.New("unterminated double-quoted string")
	errUnterminatedEscape      = errors.New("unterminated backslash-escape")
)

var (
	divideChars       = " \n\t"
	singleChar        = '\''
	doubleChar        = '"'
	escapeChar        = '\\'
	doubleEscapeChars = "$`\"\n\\"
)

func Split(input string) (words []string, err error) {
	var buf bytes.Buffer
	words = make([]string, 0)

	for len(input) > 0 {

		c, l := utf8.DecodeRuneInString(input)
		if strings.ContainsRune(divideChars, c) {
			input = input[l:]
			continue
		} else if c == escapeChar {

			next := input[l:]
			if len(next) == 0 {
				err = errUnterminatedEscape
				return words, err
			}

			c2, l2 := utf8.DecodeRuneInString(next)
			if c2 == '\n' {
				input = next[l2:]
				continue
			}
		}

		var word string

		word, input, err = divideWord(input, &buf)
		if err != nil {
			return words, err
		}

		words = append(words, word)
	}

	return words, err
}

func divideWord(input string, buf *bytes.Buffer) (word string, remainder string, err error) {
	buf.Reset()

raw:
	{
		cur := input
		for len(cur) > 0 {
			c, l := utf8.DecodeRuneInString(cur)
			cur = cur[l:]
			switch {
			case c == singleChar:
				buf.WriteString(input[0 : len(input)-len(cur)-l])
				input = cur
				goto single
			case c == doubleChar:
				buf.WriteString(input[0 : len(input)-len(cur)-l])
				input = cur
				goto double
			case c == escapeChar:
				buf.WriteString(input[0 : len(input)-len(cur)-l])
				input = cur
				goto escape
			case strings.ContainsRune(divideChars, c):
				buf.WriteString(input[0 : len(input)-len(cur)-l])
				return buf.String(), cur, nil
			}
		}
		if len(input) > 0 {
			buf.WriteString(input)
			input = ""
		}
		goto done
	}

escape:
	{
		if len(input) == 0 {
			return "", "", errUnterminatedEscape
		}
		c, l := utf8.DecodeRuneInString(input)

		if c != '\n' {
			buf.WriteString(input[:l])
		}
		input = input[l:]
	}

	goto raw

single:
	{
		i := strings.IndexRune(input, singleChar)
		if i == -1 {
			return "", "", errUnterminatedSingleQuote
		}
		buf.WriteString(input[0:i])
		input = input[i+1:]
		goto raw
	}

double:
	{
		cur := input
		for len(cur) > 0 {
			c, l := utf8.DecodeRuneInString(cur)
			cur = cur[l:]
			switch c {
			case doubleChar:
				buf.WriteString(input[0 : len(input)-len(cur)-l])
				input = cur
				goto raw
			case escapeChar:

				c2, l2 := utf8.DecodeRuneInString(cur)
				cur = cur[l2:]
				if strings.ContainsRune(doubleEscapeChars, c2) {
					buf.WriteString(input[0 : len(input)-len(cur)-l-l2])

					if c2 != '\n' {
						buf.WriteRune(c2)
					}
					input = cur
				}
			}
		}
		return "", "", errUnterminatedDoubleQuote
	}

done:
	return buf.String(), input, nil
}
