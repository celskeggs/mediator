package tokenizer

import (
	"fmt"
	"github.com/celskeggs/mediator/util"
	"strconv"
	"unicode"
)

const NoChar rune = 0

type scan struct {
	Channel <-chan RuneLoc
	Stashed rune
	Loc     SourceLocation
}

func (s *scan) HasNext() bool {
	for s.Stashed == NoChar {
		next, ok := <-s.Channel
		if !ok {
			return false
		}
		s.Stashed = next.Rune
		s.Loc = next.Loc
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

func (s *scan) StringChunk(terminator rune) (string, SourceLocation, error) {
	var runes []rune
	loc := s.Loc
	for {
		ch := s.Take()
		if ch == terminator || ch == '[' {
			s.Untake(ch)
			return string(runes), loc, nil
		}
		if ch == NoChar {
			return "", loc, fmt.Errorf("unterminated string chunk at %v", loc)
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
			return fmt.Errorf("unexpected end of input at %v", s.Loc)
		}
		switch {
		case ch == '/':
			if s.Accept('*') {
				// multi-line comment
				if !s.ConsumeRemainderOfComment() {
					return fmt.Errorf("unterminated */ at %v", s.Loc)
				}
			} else if s.Accept('/') {
				s.ConsumeUntil('\n')
				// we don't care if we run out of characters, because we'll just treat that as a final "end of line"
				s.Untake('\n')
			} else {
				output <- TokSlash.token(s.Loc)
			}
		case ch == '"':
			output <- TokStringStart.token(s.Loc)
			for !s.Accept('"') {
				chunk, loc, err := s.StringChunk('"')
				if err != nil {
					return err
				}
				if chunk != "" {
					output <- TokStringLiteral.tokenStr(chunk, loc)
				}
				if s.Accept('[') {
					output <- TokStringInsertStart.token(s.Loc)
					util.NiceToHave("fix handling of [ and ] within []")
					err := tokenizeInternal(s, output, ']')
					if err != nil {
						return err
					}
					if !s.Accept(']') {
						panic("should only have gotten here if we hit a ']'")
					}
					output <- TokStringInsertEnd.token(s.Loc)
				}
			}
			output <- TokStringEnd.token(s.Loc)
		case ch == '\'':
			chunk, loc, err := s.StringChunk('\'')
			if err != nil {
				return err
			}
			if !s.Accept('\'') {
				return fmt.Errorf("expected resource literal to be ended with a single quote at %v", s.Loc)
			}
			output <- TokResource.tokenStr(chunk, loc)
		case ch == '\r':
			// ignore \r
		case ch == '\n':
			if s.Accept(' ') {
				loc := s.Loc
				spaces := 1 + s.AcceptCount(' ')
				output <- TokSpaces.tokenInt(spaces, loc)
			} else if s.Accept('\t') {
				loc := s.Loc
				spaces := 1 + s.AcceptCount('\t')
				output <- TokTabs.tokenInt(spaces, loc)
			} else {
				output <- TokNewline.token(s.Loc)
			}
		case ch == ' ':
			// ignore spaces inside lines
		case ch == '\t':
			// ignore tabs inside lines
		case ch == '(':
			output <- TokParenOpen.token(s.Loc)
		case ch == ')':
			output <- TokParenClose.token(s.Loc)
		case ch == ',':
			output <- TokComma.token(s.Loc)
		case ch == ':':
			output <- TokColon.token(s.Loc)
		case ch == ';':
			output <- TokSemicolon.token(s.Loc)
		case ch == '.':
			loc := s.Loc
			if s.Accept('.') {
				output <- TokDotDot.token(loc)
			} else {
				output <- TokDot.token(loc)
			}
		case ch == '!':
			loc := s.Loc
			if s.Accept('=') {
				output <- TokNotEquals.token(loc)
			} else {
				output <- TokNot.token(loc)
			}
		case ch == '=':
			loc := s.Loc
			if s.Accept('=') {
				output <- TokEquals.token(loc)
			} else {
				output <- TokSetEqual.token(loc)
			}
		case ch == '<':
			loc := s.Loc
			if s.Accept('<') {
				output <- TokLeftShift.token(loc)
			} else if s.Accept('=') {
				output <- TokLessThanOrEquals.token(loc)
			} else {
				output <- TokLessThan.token(loc)
			}
		case ch == '>':
			loc := s.Loc
			if s.Accept('>') {
				output <- TokRightShift.token(loc)
			} else if s.Accept('=') {
				output <- TokGreaterThanOrEquals.token(loc)
			} else {
				output <- TokGreaterThan.token(loc)
			}
		case ch == '-':
			loc := s.Loc
			if isDigit(s.Peek()) {
				integer, err := s.Integer(true)
				if err != nil {
					return err
				}
				output <- TokInteger.tokenInt(integer, loc)
			} else {
				output <- TokMinus.token(loc)
			}
		case ch == '#':
			loc := s.Loc
			sym := s.AllMatching(IsValidInIdentifier)
			switch sym {
			case "include":
				output <- TokPreprocessorInclude.token(loc)
			case "define":
				output <- TokPreprocessorDefine.token(loc)
			default:
				return fmt.Errorf("unknown preprocessor declaration #%s at %v", sym, loc)
			}
		case isDigit(ch):
			loc := s.Loc
			s.Untake(ch)
			integer, err := s.Integer(false)
			if err != nil {
				return err
			}
			output <- TokInteger.tokenInt(integer, loc)
		case IsValidInIdentifier(ch):
			loc := s.Loc
			s.Untake(ch)
			sym := s.AllMatching(IsValidInIdentifier)
			switch sym {
			case "if":
				output <- TokKeywordIf.token(loc)
			case "return":
				output <- TokKeywordReturn.token(loc)
			case "set":
				output <- TokKeywordSet.token(loc)
			case "in":
				output <- TokKeywordIn.token(loc)
			case "new":
				output <- TokKeywordNew.token(loc)
			case "del":
				output <- TokKeywordDel.token(loc)
			case "for":
				output <- TokKeywordFor.token(loc)
			case "as":
				output <- TokKeywordAs.token(loc)
			case "var":
				output <- TokKeywordVar.token(loc)
			case "proc":
				output <- TokKeywordProc.token(loc)
			case "verb":
				output <- TokKeywordVerb.token(loc)
			default:
				output <- TokSymbol.tokenStr(sym, loc)
			}
		default:
			return fmt.Errorf("unexpected character: '%s' at %v", string([]rune{ch}), s.Loc)
		}
	}
}

func Tokenize(input <-chan RuneLoc, output chan<- Token) error {
	// we start with a newline so that indentation handling does the right thing with the first line
	scanner := &scan{input, '\n', SourceLocation{"", 0, 0}}
	defer func() {
		close(output)
		for range input {
			// discard the rest of the input
		}
	}()
	return tokenizeInternal(scanner, output, NoChar)
}
