package parser

import (
	"fmt"
	"github.com/celskeggs/mediator/dream/tokenizer"
	"strings"
)

func parseExpression0(i *input, variables []DreamMakerTypedName) (DreamMakerExpression, error) {
	loc := i.Peek().Loc
	if i.Accept(tokenizer.TokStringStart) {
		var subexpressions []DreamMakerExpression
		capitalize := true
		for !i.Accept(tokenizer.TokStringEnd) {
			partLoc := i.Peek().Loc
			if i.Accept(tokenizer.TokStringInsertStart) {
				expr, err := parseExpression(i, variables)
				if err != nil {
					return ExprNone(), err
				}
				macro := "the"
				if capitalize {
					macro = "The"
				}
				subexpressions = append(subexpressions, ExprStringMacro(macro, expr, partLoc))
				if err := i.Expect(tokenizer.TokStringInsertEnd); err != nil {
					return ExprNone(), err
				}
				capitalize = false
			} else {
				tok, err := i.ExpectParam(tokenizer.TokStringLiteral)
				if err != nil {
					return ExprNone(), err
				}
				subexpressions = append(subexpressions, ExprStringLiteral(tok.Str, tok.Loc))
				capitalize = strings.HasSuffix(strings.TrimSpace(tok.Str), ".")
			}
		}
		if len(subexpressions) == 0 {
			return ExprStringLiteral("", loc), nil
		} else if len(subexpressions) == 1 {
			return subexpressions[0], nil
		} else {
			return ExprStringConcat(subexpressions, loc), nil
		}
	} else if ok := i.Accept(tokenizer.TokKeywordNew); ok {
		typepath, err := parsePath(i)
		if err != nil {
			return ExprNone(), err
		}
		keywords, exprs, err := parseExpressionArguments(i, variables)
		if err != nil {
			return ExprNone(), err
		}
		return ExprNew(typepath, keywords, exprs, loc), nil
	} else if tok, ok := i.AcceptParam(tokenizer.TokInteger); ok {
		return ExprIntegerLiteral(tok.Int, loc), nil
	} else if tok, ok := i.AcceptParam(tokenizer.TokResource); ok {
		return ExprResourceLiteral(tok.Str, loc), nil
	} else if i.Peek().TokenType == tokenizer.TokSlash {
		tpath, err := parsePath(i)
		if err != nil {
			return ExprNone(), err
		}
		return ExprPathLiteral(tpath, loc), nil
	} else if tok, ok := i.AcceptParam(tokenizer.TokSymbol); ok {
		for _, param := range variables {
			if tok.Str == param.Name {
				return ExprGetLocal(tok.Str, loc), nil
			}
		}
		return ExprGetNonLocal(tok.Str, loc), nil
	} else {
		return ExprNone(), fmt.Errorf("invalid token %v when parsing expression at %v", i.Peek(), loc)
	}
}

func parseExpressionArguments(i *input, variables []DreamMakerTypedName) (keywords []string, expressions []DreamMakerExpression, err error) {
	if err := i.Expect(tokenizer.TokParenOpen); err != nil {
		return nil, nil, err
	}
	if i.Accept(tokenizer.TokParenClose) {
		return nil, nil, nil
	}
	for {
		var keyword string
		if i.LookAhead(1).TokenType == tokenizer.TokSetEqual {
			tok, err := i.ExpectParam(tokenizer.TokSymbol)
			if err != nil {
				return nil, nil, err
			}
			keyword = tok.Str
			if !i.Accept(tokenizer.TokSetEqual) {
				panic("should have been no way for the next token to not be TokSetEqual")
			}
		}
		expr, err := parseExpression(i, variables)
		if err != nil {
			return nil, nil, err
		}
		keywords = append(keywords, keyword)
		expressions = append(expressions, expr)
		if !i.Accept(tokenizer.TokComma) {
			break
		}
	}
	if err := i.Expect(tokenizer.TokParenClose); err != nil {
		return nil, nil, err
	}
	return keywords, expressions, nil
}

func parseExpression1(i *input, variables []DreamMakerTypedName) (DreamMakerExpression, error) {
	expr, err := parseExpression0(i, variables)
	if err != nil {
		return ExprNone(), err
	}
	for {
		loc := i.Peek().Loc
		if i.Peek().TokenType == tokenizer.TokParenOpen {
			keywords, exprs, err := parseExpressionArguments(i, variables)
			if err != nil {
				return ExprNone(), err
			}
			expr = ExprCall(expr, keywords, exprs, loc)
		} else if i.Accept(tokenizer.TokDot) {
			field, err := i.ExpectParam(tokenizer.TokSymbol)
			if err != nil {
				return ExprNone(), err
			}
			expr = ExprGetField(expr, field.Str, field.Loc)
		} else {
			return expr, nil
		}
	}
}

func parseExpression(i *input, variables []DreamMakerTypedName) (DreamMakerExpression, error) {
	loc := i.Peek().Loc
	if i.Accept(tokenizer.TokNot) {
		expr, err := parseExpression1(i, variables)
		if err != nil {
			return ExprNone(), err
		}
		return ExprBooleanNot(expr, loc), nil
	}
	return parseExpression1(i, variables)
}
