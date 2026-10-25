package strutil

func AdjustSurroundQuotes(dBpos, dEpos, sBpos, sEpos int) (mark, cpos int) {
	mark = -1
	cpos = -1

	if (sBpos == -1 || sEpos == -1) && (dBpos == -1 || dEpos == -1) {
		return
	}

	doubleFirstAndValid := (dBpos < sBpos &&
		dBpos >= 0 &&
		sBpos >= 0 &&
		dEpos > sEpos)

	singleFirstAndValid := (sBpos < dBpos &&
		sBpos >= 0 &&
		dBpos >= 0 &&
		sEpos > dEpos)

	if (sBpos == -1 || sEpos == -1) || doubleFirstAndValid {
		mark = dBpos
		cpos = dEpos
	} else if (dBpos == -1 || dEpos == -1) || singleFirstAndValid {
		mark = sBpos
		cpos = sEpos
	}

	return
}

func GetQuotedWordStart(line []rune) (unclosed bool, pos int) {
	var (
		inSingle, inDouble bool
		escape             bool
		spos, dpos         = -1, -1
	)

	for i, r := range line {
		if escape {
			escape = false
			continue
		}

		if r == '\\' {
			if !inSingle {
				escape = true
				continue
			}
		}

		switch r {
		case '"':
			if !inSingle {
				inDouble = !inDouble
				if inDouble {
					dpos = i
				}
			}
		case '\'':
			if !inDouble {
				inSingle = !inSingle
				if inSingle {
					spos = i
				}
			}
		}
	}

	if inDouble {
		return true, dpos
	}

	if inSingle {
		return true, spos
	}

	return false, -1
}

func IsBracket(char rune) bool {
	if char == '(' ||
		char == ')' ||
		char == '{' ||
		char == '}' ||
		char == '[' ||
		char == ']' {
		return true
	}

	return false
}

func IsSurround(bchar, echar rune) bool {
	switch bchar {
	case '{':
		return echar == '}'
	case '(':
		return echar == ')'
	case '[':
		return echar == ']'
	case '<':
		return echar == '>'
	case '"':
		return echar == '"'
	case '\'':
		return echar == '\''
	}

	return echar == bchar
}

func MatchSurround(r rune) (bchar, echar rune) {
	bchar = r
	echar = r

	switch bchar {
	case '{':
		echar = '}'
	case '(':
		echar = ')'
	case '[':
		echar = ']'
	case '<':
		echar = '>'
	case '}':
		bchar = '{'
		echar = '}'
	case ')':
		bchar = '('
		echar = ')'
	case ']':
		bchar = '['
		echar = ']'
	case '>':
		bchar = '<'
		echar = '>'
	case '"':
		bchar = '"'
		echar = '"'
	case '\'':
		bchar = '\''
		echar = '\''
	}

	return bchar, echar
}

func SurroundType(char rune) (surround, closer bool) {
	switch char {
	case '{':
		return true, false
	case '}':
		return true, true
	case '(':
		return true, false
	case ')':
		return true, true
	case '[':
		return true, false
	case ']':
		return true, true
	case '<':
		return true, false
	case '>':
		return true, true
	case '"':
		return true, true
	case '\'':
		return true, true
	}

	return false, false
}
