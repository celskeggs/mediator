package parser

import (
	"fmt"
	"github.com/celskeggs/mediator/dream/ast"
	"github.com/celskeggs/mediator/dream/tokenizer"
	"github.com/celskeggs/mediator/util"
	"strings"
)

func parseExpression0(i *input, scope *Scope) (ast.Expression, error) {
	loc := i.Peek().Loc
	if i.Accept(tokenizer.TokStringStart) {
		var subexpressions []ast.Expression
		capitalize := true
		for !i.Accept(tokenizer.TokStringEnd) {
			partLoc := i.Peek().Loc
			if i.Accept(tokenizer.TokStringInsertStart) {
				expr, err := parseExpression(i, scope)
				if err != nil {
					return ast.ExprNone(), err
				}
				macro := "the"
				if capitalize {
					macro = "The"
				}
				subexpressions = append(subexpressions, ast.ExprStringMacro(macro, expr, partLoc))
				if err := i.Expect(tokenizer.TokStringInsertEnd); err != nil {
					return ast.ExprNone(), err
				}
				capitalize = false
			} else {
				tok, err := i.ExpectParam(tokenizer.TokStringLiteral)
				if err != nil {
					return ast.ExprNone(), err
				}
				subexpressions = append(subexpressions, ast.ExprStringLiteral(tok.Str, tok.Loc))
				capitalize = strings.HasSuffix(strings.TrimSpace(tok.Str), ".")
			}
		}
		if len(subexpressions) == 0 {
			return ast.ExprStringLiteral("", loc), nil
		} else if len(subexpressions) == 1 {
			return subexpressions[0], nil
		} else {
			return ast.ExprStringConcat(subexpressions, loc), nil
		}
	} else if ok := i.Accept(tokenizer.TokKeywordNew); ok {
		typepath, err := parsePath(i)
		if err != nil {
			return ast.ExprNone(), err
		}
		keywords, exprs, err := parseExpressionArguments(i, scope)
		if err != nil {
			return ast.ExprNone(), err
		}
		return ast.ExprNew(typepath, keywords, exprs, loc), nil
	} else if tok, ok := i.AcceptParam(tokenizer.TokInteger); ok {
		return ast.ExprIntegerLiteral(tok.Int, loc), nil
	} else if tok, ok := i.AcceptParam(tokenizer.TokResource); ok {
		return ast.ExprResourceLiteral(tok.Str, loc), nil
	} else if i.Peek().TokenType == tokenizer.TokSlash {
		tpath, err := parsePath(i)
		if err != nil {
			return ast.ExprNone(), err
		}
		return ast.ExprPathLiteral(tpath, loc), nil
	} else if tok, ok := i.AcceptParam(tokenizer.TokSymbol); ok {
		if scope.HasVar(tok.Str) {
			return ast.ExprGetLocal(tok.Str, loc), nil
		} else {
			return ast.ExprGetNonLocal(tok.Str, loc), nil
		}
	} else if i.Accept(tokenizer.TokDot) {
		util.FIXME("support .()")
		return ast.ExprGetLocal(".", loc), nil
	} else if i.Accept(tokenizer.TokDotDot) {
		return ast.ExprGetNonLocal("..", loc), nil
	} else {
		return ast.ExprNone(), fmt.Errorf("invalid token %v when parsing expression at %v", i.Peek(), loc)
	}
}

func parseExpressionArguments(i *input, scope *Scope) (keywords []string, expressions []ast.Expression, err error) {
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
		expr, err := parseExpression(i, scope)
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

func parseExpression1(i *input, scope *Scope) (ast.Expression, error) {
	expr, err := parseExpression0(i, scope)
	if err != nil {
		return ast.ExprNone(), err
	}
	for {
		loc := i.Peek().Loc
		if i.Peek().TokenType == tokenizer.TokParenOpen {
			keywords, exprs, err := parseExpressionArguments(i, scope)
			if err != nil {
				return ast.ExprNone(), err
			}
			expr = ast.ExprCall(expr, keywords, exprs, loc)
		} else if i.Accept(tokenizer.TokDot) {
			field, err := i.ExpectParam(tokenizer.TokSymbol)
			if err != nil {
				return ast.ExprNone(), err
			}
			expr = ast.ExprGetField(expr, field.Str, field.Loc)
		} else {
			return expr, nil
		}
	}
}

func parseExpression(i *input, scope *Scope) (ast.Expression, error) {
	loc := i.Peek().Loc
	if i.Accept(tokenizer.TokNot) {
		expr, err := parseExpression1(i, scope)
		if err != nil {
			return ast.ExprNone(), err
		}
		return ast.ExprBooleanNot(expr, loc), nil
	}
	return parseExpression1(i, scope)
}
