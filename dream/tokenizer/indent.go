package tokenizer

import "github.com/pkg/errors"

type indentState struct {
	IsTabs bool
	Stops  []int
	Output chan <-Token
}

func (i *indentState) Clear() error {
	for n := 0; n < len(i.Stops); n++ {
		i.Output <- TokUnindent.token()
	}
	i.Stops = nil
	return nil
}

func (i *indentState) setCount(indent int) error {
	if indent == 0 {
		// should have called Clear instead
		panic("cannot have zero indent")
	}
	if len(i.Stops) == 0 || indent > i.Stops[len(i.Stops)-1] {
		i.Stops = append(i.Stops, indent)
		i.Output <- TokIndent.token()
	} else {
		for len(i.Stops) > 0 && indent < i.Stops[len(i.Stops)-1] {
			i.Stops = i.Stops[:len(i.Stops)-1]
			if len(i.Stops) == 0 {
				return errors.New("too much indent to run out of indent")
			}
			if indent > i.Stops[len(i.Stops)-1] {
				return errors.New("did not remove enough indentation levels at once")
			}
			i.Output <- TokUnindent.token()
		}
	}
	return nil
}

func (i *indentState) SetTabs(tabs int) error {
	if !i.IsTabs && len(i.Stops) > 0 {
		return errors.New("mixing tabs and spaces")
	}
	i.IsTabs = true
	return i.setCount(tabs)
}

func (i *indentState) SetSpaces(spaces int) error {
	if i.IsTabs && len(i.Stops) > 0 {
		return errors.New("mixing tabs and spaces")
	}
	i.IsTabs = false
	return i.setCount(spaces)
}

func (i *indentState) UpdateForToken(t Token) error {
	if t.TokenType == TokNewline {
		return i.Clear()
	} else if t.TokenType == TokSpaces {
		return i.SetSpaces(int(t.Int))
	} else if t.TokenType == TokTabs {
		return i.SetTabs(int(t.Int))
	} else {
		panic("not a spacing token")
	}
}

func isSpacingToken(t Token) bool {
	return t.TokenType == TokNewline || t.TokenType == TokSpaces || t.TokenType == TokTabs
}

func ProcessIndentation(input <-chan Token, output chan<- Token) error {
	defer func() {
		close(output)
		for range input {
			// drain the rest of the input
		}
	}()
	lastSpacingToken := TokNone.token()
	indentState := indentState{
		Output: output,
	}
	for token := range input {
		if isSpacingToken(token) {
			lastSpacingToken = token
		} else if token.TokenType == TokIndent || token.TokenType == TokUnindent {
			panic("should never receive Indent/Unindent in ProcessIndentation")
		} else {
			if lastSpacingToken.TokenType != TokNone {
				output <- TokNewline.token()
				err := indentState.UpdateForToken(lastSpacingToken)
				if err != nil {
					return err
				}
			}
			lastSpacingToken = TokNone.token()
			output <- token
		}
	}
	return indentState.Clear()
}
