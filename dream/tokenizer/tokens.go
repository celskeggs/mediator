package tokenizer

import "fmt"

type TokenType uint8

const (
	// we don't actually need TokNone; it's just useful to make sure that none of our tokens are equal to 0
	TokNone TokenType = iota

	// symbols
	TokSlash
	TokMinus
	TokSetEqual
	TokParenOpen
	TokParenClose
	TokComma
	TokDot
	TokDotDot
	TokColon
	TokSemicolon
	TokEquals
	TokNotEquals
	TokLessThan
	TokGreaterThan
	TokLessThanOrEquals
	TokGreaterThanOrEquals
	TokLeftShift
	TokRightShift
	TokNot

	// literals
	TokInteger
	TokSymbol
	TokResource
	TokStringStart
	TokStringEnd
	TokStringInsertStart
	TokStringInsertEnd
	TokStringLiteral

	// spacing
	TokNewline
	TokSpaces
	TokTabs
	TokIndent
	TokUnindent
)

type tokenMode uint8

const (
	modeUnknown tokenMode = iota
	modeNone
	modeInt
	modeStr
)

type Token struct {
	TokenType
	mode tokenMode
	Int  int64
	Str  string
}

func (t TokenType) token() Token {
	return Token{t, modeNone, 0, ""}
}

func (t TokenType) tokenInt(integer int64) Token {
	return Token{t, modeInt, integer, ""}
}

func (t TokenType) tokenStr(str string) Token {
	return Token{t, modeStr, 0, str}
}

func (t TokenType) String() string {
	switch t {
	case TokNone:
		return "TokNone"
	case TokSlash:
		return "TokSlash"
	case TokMinus:
		return "TokMinus"
	case TokSetEqual:
		return "TokSetEqual"
	case TokParenOpen:
		return "TokParenOpen"
	case TokParenClose:
		return "TokParenClose"
	case TokComma:
		return "TokComma"
	case TokDot:
		return "TokDot"
	case TokDotDot:
		return "TokDotDot"
	case TokColon:
		return "TokColon"
	case TokSemicolon:
		return "TokSemicolon"
	case TokEquals:
		return "TokEquals"
	case TokNotEquals:
		return "TokNotEquals"
	case TokLessThan:
		return "TokLessThan"
	case TokGreaterThan:
		return "TokGreaterThan"
	case TokLessThanOrEquals:
		return "TokLessThanOrEquals"
	case TokGreaterThanOrEquals:
		return "TokGreaterThanOrEquals"
	case TokLeftShift:
		return "TokLeftShift"
	case TokRightShift:
		return "TokRightShift"
	case TokNot:
		return "TokNot"
	case TokInteger:
		return "TokInteger"
	case TokSymbol:
		return "TokSymbol"
	case TokResource:
		return "TokResource"
	case TokStringStart:
		return "TokStringStart"
	case TokStringEnd:
		return "TokStringEnd"
	case TokStringInsertStart:
		return "TokStringInsertStart"
	case TokStringInsertEnd:
		return "TokStringInsertEnd"
	case TokStringLiteral:
		return "TokStringLiteral"
	case TokNewline:
		return "TokNewline"
	case TokSpaces:
		return "TokSpaces"
	case TokTabs:
		return "TokTabs"
	case TokIndent:
		return "TokIndent"
	case TokUnindent:
		return "TokUnindent"
	default:
		panic(fmt.Sprintf("unrecognized token: %d", t))
	}
}

func (t Token) String() string {
	switch t.mode {
	case modeNone:
		return fmt.Sprintf("%s()", t.TokenType.String())
	case modeInt:
		return fmt.Sprintf("%s(%d)", t.TokenType.String(), t.Int)
	case modeStr:
		return fmt.Sprintf("%s(%s)", t.TokenType.String(), t.Str)
	default:
		panic("unknown mode")
	}
}

func (t Token) IsNone() bool {
	return t.TokenType == TokNone
}

func NoToken() Token {
	return TokNone.token()
}
