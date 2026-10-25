package readline

func (rl *Shell) ClearInlineSuggestion() {
	if rl == nil || rl.Display == nil {
		return
	}

	rl.Display.ClearInlineSuggestion()
}

func (rl *Shell) GetInlineSuggestion() string {
	if rl == nil || rl.Display == nil {
		return ""
	}

	return rl.Display.GetInlineSuggestion()
}

func (rl *Shell) SetInlineSuggestion(suggestion string) {
	if rl == nil || rl.Display == nil {
		return
	}

	rl.Display.SetInlineSuggestion(suggestion)
}

func (rl *Shell) acceptInlineSuggestion() bool {
	if rl == nil || rl.Display == nil {
		return false
	}

	return rl.Display.AcceptInline(rl.line, rl.cursor)
}

func (rl *Shell) inlineSuggestAccept() {
	rl.acceptInlineSuggestion()
}

func (rl *Shell) inlineSuggestAcceptWord() {
	if rl == nil || rl.Display == nil {
		return
	}

	rl.Display.AcceptInlineWord(rl.line, rl.cursor)
}
