package tokenizer

import (
	"github.com/pkg/errors"
	"fmt"
	"unicode"
	"strconv"
	"github.com/celskeggs/mediator/util"
)

const NoChar rune = 0

type scan struct {
	Channel <-chan rune
	Stashed rune
}

func (s *scan) HasNext() bool {
	for s.Stashed == NoChar {
		next, ok := <-s.Channel
		if !ok {
			return false
		}
		s.Stashed = next
	}
	return true
}

func (s *scan) Untake(r rune) {
	if s.Stashed != NoChar {
		panic("untake when character already for taking")
	}
	s.Stashed = r
}

func (s *scan) Peek() rune {
	if !s.HasNext() {
		return NoChar
	}
	return s.Stashed
}

func (s *scan) Consume() {
	if !s.HasNext() {
		panic("cannot consume: no character to consume")
	}
	s.Stashed = NoChar
}

func (s *scan) Take() rune {
	r := s.Peek()
	if r != NoChar {
		s.Consume()
	}
	return r
}

func (s *scan) Accept(r rune) bool {
	if s.Peek() == r {
		s.Consume()
		return true
	}
	return false
}

func (s *scan) ConsumeUntil(r rune) bool {
	for {
		ch := s.Take()
		if ch == r {
			return true
		} else if ch == NoChar {
			return false
		}
	}
}

func runeArrayEq(a []rune, b []rune) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func (s *scan) ConsumeRemainderOfComment() bool {
	depth := 1
	for depth > 0 {
		if s.Accept('/') && s.Accept('*') {
			depth += 1
		} else if s.Accept('*') && s.Accept('/') {
			depth -= 1
		} else if !s.HasNext() {
			return false
		} else {
			s.Consume()
		}
	}
	return true
}

func (s *scan) AcceptCount(r rune) (count int64) {
	for s.Accept(r) {
		count += 1
	}
	return count
}

func (s *scan) AllMatching(predicate func(rune) bool) string {
	var runes []rune
	for s.HasNext() && predicate(s.Peek()) {
		runes = append(runes, s.Take())
	}
	return string(runes)
}

func isDigit(r rune) bool {
	return '0' <= r && r <= '9'
}

func (s *scan) Integer(negative bool) (int64, error) {
	var prefix string
	// we do this so that we don't need to deal with overflow errors (by passing them off to ParseInt)
	if negative {
		prefix = "-"
	}
	strInt := s.AllMatching(isDigit)
	if len(strInt) == 0 {
		panic("Integer() expects that at least one digit comes next")
	}
	return strconv.ParseInt(prefix+strInt, 10, 64)
}

func (s *scan) StringChunk(terminator rune) (string, error) {
	var runes []rune
	for {
		ch := s.Take()
		if ch == terminator || ch == '[' {
			s.Untake(ch)
			return string(runes), nil
		}
		if ch == NoChar {
			return "", errors.New("unterminated string chunk")
		}
		if ch == '\\' {
			if ch != terminator && ch != '\\' {
				util.FIXME("handle text macros here")
				panic("unimplemented: text macros")
			}
			ch = s.Take()
		}
		runes = append(runes, ch)
	}
}

func IsValidInIdentifier(r rune) bool {
	// for now we're just borrowing Go's rules
	return unicode.IsLetter(r) || r == '_' || unicode.IsNumber(r) || unicode.IsDigit(r)
}

func tokenizeInternal(s *scan, output chan<- Token, terminator rune) error {
	for {
		ch := s.Take()
		if ch == terminator {
			// does nothing if terminator was NoChar
			s.Untake(ch)
			return nil
		}
		if ch == NoChar {
			return errors.New("unexpected end of input")
		}
		switch {
		case ch == '/':
			if s.Accept('*') {
				// multi-line comment
				if !s.ConsumeRemainderOfComment() {
					return errors.New("unterminated */")
				}
			} else if s.Accept('/') {
				s.ConsumeUntil('\n')
				// we don't care if we run out of characters, because we'll just treat that as a final "end of line"
				s.Untake('\n')
			} else {
				output <- TokSlash.token()
			}
		case ch == '"':
			output <- TokStringStart.token()
			for !s.Accept('"') {
				chunk, err := s.StringChunk('"')
				if err != nil {
					return err
				}
				if chunk != "" {
					output <- TokStringLiteral.tokenStr(chunk)
				}
				if s.Accept('[') {
					output <- TokStringInsertStart.token()
					util.NiceToHave("fix handling of [ and ] within []")
					err := tokenizeInternal(s, output, ']')
					if err != nil {
						return err
					}
					if !s.Accept(']') {
						panic("should only have gotten here if we hit a ']'")
					}
					output <- TokStringInsertEnd.token()
				}
			}
			output <- TokStringEnd.token()
		case ch == '\'':
			chunk, err := s.StringChunk('\'')
			if err != nil {
				return err
			}
			if !s.Accept('\'') {
				return errors.New("expected resource literal to be ended with a single quote")
			}
			output <- TokResource.tokenStr(chunk)
		case ch == '\r':
			// ignore \r
		case ch == '\n':
			if s.Accept(' ') {
				spaces := 1 + s.AcceptCount(' ')
				output <- TokSpaces.tokenInt(spaces)
			} else if s.Accept('\t') {
				spaces := 1 + s.AcceptCount('\t')
				output <- TokTabs.tokenInt(spaces)
			} else {
				output <- TokNewline.token()
			}
		case ch == ' ':
			// ignore spaces inside lines
		case ch == '\t':
			// ignore tabs inside lines
		case ch == '(':
			output <- TokParenOpen.token()
		case ch == ')':
			output <- TokParenClose.token()
		case ch == ',':
			output <- TokComma.token()
		case ch == ':':
			output <- TokColon.token()
		case ch == ';':
			output <- TokSemicolon.token()
		case ch == '.':
			if s.Accept('.') {
				output <- TokDotDot.token()
			} else {
				output <- TokDot.token()
			}
		case ch == '!':
			if s.Accept('=') {
				output <- TokNotEquals.token()
			} else {
				output <- TokNot.token()
			}
		case ch == '=':
			if s.Accept('=') {
				output <- TokEquals.token()
			} else {
				output <- TokSetEqual.token()
			}
		case ch == '<':
			if s.Accept('<') {
				output <- TokLeftShift.token()
			} else if s.Accept('=') {
				output <- TokLessThanOrEquals.token()
			} else {
				output <- TokLessThan.token()
			}
		case ch == '>':
			if s.Accept('>') {
				output <- TokRightShift.token()
			} else if s.Accept('=') {
				output <- TokGreaterThanOrEquals.token()
			} else {
				output <- TokGreaterThan.token()
			}
		case ch == '-':
			if isDigit(s.Peek()) {
				integer, err := s.Integer(true)
				if err != nil {
					return err
				}
				output <- TokInteger.tokenInt(integer)
			} else {
				output <- TokMinus.token()
			}
		case isDigit(ch):
			s.Untake(ch)
			integer, err := s.Integer(false)
			if err != nil {
				return err
			}
			output <- TokInteger.tokenInt(integer)
		case IsValidInIdentifier(ch):
			s.Untake(ch)
			output <- TokSymbol.tokenStr(s.AllMatching(IsValidInIdentifier))
		default:
			return fmt.Errorf("unexpected character: '%s'", string([]rune{ch}))
		}
	}
}

func Tokenize(input <-chan rune, output chan<- Token) error {
	// we start with a newline so that indentation handling does the right thing with the first line
	scanner := &scan{input, '\n'}
	defer func() {
		close(output)
		for range input {
			// discard the rest of the input
		}
	}()
	return tokenizeInternal(scanner, output, NoChar)
}
