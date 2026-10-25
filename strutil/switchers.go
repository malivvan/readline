package strutil

import (
	"fmt"
	"math"
	"math/big"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

type KeywordSwitcher func(word string, increase bool, times int) (done bool, switched string, bpos, epos int)

const (
	highestInt = 9223372036854775807
)

func AdjustNumberOperatorPos(cpos int, line []rune) int {
	word := lineSlice(line, cpos, 2)

	if match, _ := regexp.MatchString(`[+-][0-9]`, word); match {

		digit := regexp.MustCompile(`[^0-9]`)
		if cpos == 0 || digit.MatchString(string(line[cpos-1])) {
			cpos++
		}
	} else if match, _ := regexp.MatchString(`[+-][a-zA-Z]`, word); match {

		if cpos == 0 || line[cpos-1] == ' ' {
			cpos++
		}
	}

	return cpos
}

func KeywordSwitchers() []KeywordSwitcher {
	return []KeywordSwitcher{
		switchNumber,
		switchBoolean,
		switchWeekday,
		switchOperator,
	}
}

func lineSlice(line []rune, cpos, adjust int) (slice string) {
	switch {
	case cpos+adjust > len(line):
		slice = string(line[cpos:])
	case adjust < 0:
		if cpos+adjust < 0 {
			slice = string(line[:cpos])
		} else {
			slice = string(line[cpos+adjust : cpos])
		}
	default:
		slice = string(line[cpos : cpos+adjust])
	}

	return
}

func switchBinary(word string, inc int) (done bool, switched string, bpos, epos int) {
	binary := regexp.MustCompile(`[^0-9]?(0[bB][01]*)`)
	match := binary.FindString(word)

	if match == "" {
		return done, switched, bpos, epos
	}

	done = true

	number := match
	prefix := match[:2]
	binVal := number[len(prefix):]
	indexes := binary.FindStringIndex(word)
	mbegin, mend := indexes[0], indexes[1]
	bpos, epos = mbegin, mend

	num, err := strconv.ParseInt(binVal, 2, 64)
	if err != nil {
		done = false
		return done, switched, bpos, epos
	}

	max64Bit := big.NewInt(highestInt)
	bigNum := big.NewInt(num)
	bigInc := big.NewInt(int64(inc))
	sum := bigNum.Add(bigNum, bigInc)
	zero := big.NewInt(0)

	numBefore := num

	switch {
	case sum.Cmp(zero) < 0:
		offset := bigInc.Sub(max64Bit, sum.Abs(sum))
		if offset.IsInt64() {
			num = offset.Int64()
		} else {
			num = math.MaxInt64
		}
	case sum.CmpAbs(max64Bit) >= 0:
		offset := bigInc.Sub(sum, max64Bit)
		if offset.IsInt64() {
			num = offset.Int64()
		} else {
			num = int64(inc) - (num - numBefore)
		}
	default:
		num = sum.Int64()
	}

	binVal = fmt.Sprintf("%b", num)
	switched = prefix + binVal

	return done, switched, bpos, epos
}

func switchBoolean(word string, _ bool, _ int) (done bool, switched string, bpos, epos int) {
	epos = len(word)

	option := regexp.MustCompile(`(^[+-]{0,2})`)
	if match := option.FindString(word); match != "" {
		indexes := option.FindStringIndex(word)
		bpos = indexes[1]
		word = word[bpos:]
	}

	booleans := map[string]string{
		"true":  "false",
		"false": "true",
		"t":     "f",
		"f":     "t",
		"yes":   "no",
		"no":    "yes",
		"y":     "n",
		"n":     "y",
		"on":    "off",
		"off":   "on",
	}

	switched, done = booleans[strings.ToLower(word)]
	if !done {
		return done, switched, bpos, epos
	}

	done = true

	if match, _ := regexp.MatchString(`^[A-Z]+$`, word); match {
		switched = strings.ToLower(switched)
	} else if match, _ := regexp.MatchString(`^[A-Z]`, word); match {
		letter := switched[0]
		upper := unicode.ToUpper(rune(letter))
		switched = string(upper) + switched[1:]
	}

	return done, switched, bpos, epos
}

func switchDecimal(word string, inc int) (done bool, switched string, bpos, epos int) {
	decimal := regexp.MustCompile(`([-+]?[0-9]+)`)
	match := decimal.FindString(word)

	if match == "" {
		return
	}

	done = true

	indexes := decimal.FindStringIndex(word)
	mbegin, mend := indexes[0], indexes[1]
	bpos, epos = mbegin, mend

	num, _ := strconv.Atoi(match)
	num += inc

	switched = strconv.Itoa(num)

	if word[0] == '+' {
		switched = "+" + switched
	}

	if switched == "" {
		done = false
	}

	return
}

func switchHexa(word string, inc int) (done bool, switched string, bpos, epos int) {
	hexadecimal := regexp.MustCompile(`[^0-9]?(0[xX][0-9a-fA-F]*)`)
	match := hexadecimal.FindString(word)

	if match == "" {
		return done, switched, bpos, epos
	}

	done = true

	number := match
	prefix := match[:2]
	hexVal := number[len(prefix):]
	indexes := hexadecimal.FindStringIndex(word)
	mbegin, mend := indexes[0], indexes[1]
	bpos, epos = mbegin, mend

	if match, _ := regexp.MatchString(`[A-Z][0-9]*$`, number); !match {
		hexVal = strings.ToUpper(hexVal)
	}

	num, err := strconv.ParseInt(hexVal, 16, 64)
	if err != nil {
		done = false
		return done, switched, bpos, epos
	}

	max64Bit := big.NewInt(highestInt)
	bigNum := big.NewInt(num)
	bigInc := big.NewInt(int64(inc))
	sum := bigNum.Add(bigNum, bigInc)

	numBefore := num

	switch {
	case sum.Cmp(big.NewInt(0)) < 0:
		offset := bigInc.Sub(max64Bit, sum.Abs(sum))
		if offset.IsInt64() {
			num = offset.Int64()
		} else {
			num = math.MaxInt64
		}
	case sum.CmpAbs(max64Bit) >= 0:
		offset := bigInc.Sub(sum, max64Bit)
		if offset.IsInt64() {
			num = offset.Int64()
		} else {
			num = int64(inc) - (num - numBefore)
		}
	default:
		num = sum.Int64()
	}

	hexVal = fmt.Sprintf("%x", num)
	switched = prefix + hexVal

	return done, switched, bpos, epos
}

func switchNumber(word string, _ bool, times int) (done bool, switched string, bpos, epos int) {
	if done, switched, bpos, epos = switchHexa(word, times); done {
		return
	}

	if done, switched, bpos, epos = switchBinary(word, times); done {
		return
	}

	if done, switched, bpos, epos = switchDecimal(word, times); done {
		return
	}

	return
}

func switchOperator(word string, _ bool, _ int) (done bool, switched string, bpos, epos int) {
	epos = len(word)

	operators := map[string]string{
		"&&":  "||",
		"||":  "&&",
		"++":  "--",
		"--":  "++",
		"==":  "!=",
		"!=":  "==",
		"===": "!==",
		"!==": "===",
		"+":   "-",
		"-":   "*",
		"*":   "/",
		"/":   "+",
		"and": "or",
		"or":  "and",
	}

	switched, done = operators[strings.ToLower(word)]
	if !done {
		return done, switched, bpos, epos
	}

	done = true

	if match, _ := regexp.MatchString(`^[A-Z]+$`, word); match {
		switched = strings.ToLower(switched)
	} else if match, _ := regexp.MatchString(`^[A-Z]`, word); match {
		letter := switched[0]
		upper := unicode.ToUpper(rune(letter))
		switched = string(upper) + switched[1:]
	}

	return done, switched, bpos, epos
}

func switchWeekday(_ string, _ bool, _ int) (done bool, switched string, bpos, epos int) {
	return
}
