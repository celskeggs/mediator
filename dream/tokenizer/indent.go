package tokenizer

import (
	"fmt"
)

type indentState struct {
	IsTabs bool
	Stops  []int
	Output chan<- Token
}

func (i *indentState) Clear(loc SourceLocation) error {
	i.Output <- TokNewline.token(loc)
	for n := 0; n < len(i.Stops); n++ {
		i.Output <- TokUnindent.token(loc)
	}
	i.Stops = nil
	return nil
}

func (i *indentState) setCount(indent int, loc SourceLocation) error {
	if indent == 0 {
		// should have called Clear instead
		panic("cannot have zero indent")
	}
	if len(i.Stops) == 0 || indent > i.Stops[len(i.Stops)-1] {
		i.Stops = append(i.Stops, indent)
		i.Output <- TokIndent.token(loc)
	} else {
		i.Output <- TokNewline.token(loc)
		for len(i.Stops) > 0 && indent < i.Stops[len(i.Stops)-1] {
			i.Stops = i.Stops[:len(i.Stops)-1]
			if len(i.Stops) == 0 {
				return fmt.Errorf("too much indent to run out of indent at %v", loc)
			}
			if indent > i.Stops[len(i.Stops)-1] {
				return fmt.Errorf("did not remove enough indentation levels at once at %v", loc)
			}
			i.Output <- TokUnindent.token(loc)
		}
	}
	return nil
}

func (i *indentState) SetTabs(tabs int, loc SourceLocation) error {
	if !i.IsTabs && len(i.Stops) > 0 {
		return fmt.Errorf("mixing tabs and spaces at %v", loc)
	}
	i.IsTabs = true
	return i.setCount(tabs, loc)
}

func (i *indentState) SetSpaces(spaces int, loc SourceLocation) error {
	if i.IsTabs && len(i.Stops) > 0 {
		return fmt.Errorf("mixing tabs and spaces at %v", loc)
	}
	i.IsTabs = false
	return i.setCount(spaces, loc)
}

func (i *indentState) UpdateForToken(t Token) error {
	if t.TokenType == TokNewline {
		return i.Clear(t.Loc)
	} else if t.TokenType == TokSpaces {
		return i.SetSpaces(int(t.Int), t.Loc)
	} else if t.TokenType == TokTabs {
		return i.SetTabs(int(t.Int), t.Loc)
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
	lastSpacingToken := NoToken()
	indentState := indentState{
		Output: output,
	}
	for token := range input {
		if isSpacingToken(token) {
			lastSpacingToken = token
		} else if token.TokenType == TokIndent || token.TokenType == TokUnindent {
			panic("should never receive Indent/Unindent in ProcessIndentation")
		} else {
			if !lastSpacingToken.IsNone() {
				err := indentState.UpdateForToken(lastSpacingToken)
				if err != nil {
					return err
				}
			}
			lastSpacingToken = NoToken()
			output <- token
		}
	}
	return indentState.Clear(lastSpacingToken.Loc)
}
