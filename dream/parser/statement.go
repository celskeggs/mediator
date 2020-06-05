package parser

import (
	"fmt"
	"github.com/celskeggs/mediator/autocoder/dtype"
	"github.com/celskeggs/mediator/dream/ast"
	"github.com/celskeggs/mediator/dream/tokenizer"
	"github.com/celskeggs/mediator/util"
)

func parseStatement(i *input, scope *Scope) (ast.Statement, error) {
	loc := i.Peek().Loc
	if i.Accept(tokenizer.TokKeywordIf) {
		if err := i.Expect(tokenizer.TokParenOpen); err != nil {
			return ast.StatementNone(), err
		}
		condition, err := parseExpression(i, scope)
		if err != nil {
			return ast.StatementNone(), err
		}
		if err := i.Expect(tokenizer.TokParenClose); err != nil {
			return ast.StatementNone(), err
		}
		statements, err := parseStatementBlock(i, scope)
		if err != nil {
			return ast.StatementNone(), err
		}
		return ast.StatementIf(condition, statements, loc), nil
	} else if i.Accept(tokenizer.TokKeywordFor) {
		if err := i.Expect(tokenizer.TokParenOpen); err != nil {
			return ast.StatementNone(), err
		}
		loc2 := i.Peek().Loc
		varpath, err := parseDeclPath(i)
		if err != nil {
			return ast.StatementNone(), err
		}
		if !varpath.IsVarDef() {
			return ast.StatementNone(), fmt.Errorf("path %v is not a variable definition at %v", varpath, loc2)
		}
		varTarget, varTypePath, varName := varpath.SplitDef()
		varType := dtype.FromPath(varTypePath)
		if !varTarget.IsEmpty() {
			return ast.StatementNone(), fmt.Errorf("invalid prefix for 'var' in path %v at %v", varpath, loc2)
		}
		if i.Peek().TokenType == tokenizer.TokKeywordAs {
			return ast.StatementNone(), fmt.Errorf("unsupported: keyword as in for loop at %v", i.Peek().Loc)
		}
		var inExpr ast.Expression
		if i.Accept(tokenizer.TokKeywordIn) {
			if inExpr, err = parseExpression(i, scope); err != nil {
				return ast.StatementNone(), err
			}
		}
		if err := i.Expect(tokenizer.TokParenClose); err != nil {
			return ast.StatementNone(), err
		}
		scope.AddVar(varName)
		body, err := parseStatementBlock(i, scope)
		scope.RemoveVar(varName)
		if err != nil {
			return ast.StatementNone(), err
		}
		return ast.StatementForList(varType, varName, inExpr, body, loc), nil
	} else if i.Accept(tokenizer.TokKeywordReturn) {
		if err := i.Expect(tokenizer.TokNewline); err != nil {
			return ast.StatementNone(), err
		}
		util.FIXME("support returning values")
		return ast.StatementReturn(loc), nil
	} else if i.Accept(tokenizer.TokKeywordSet) {
		sym, err := i.ExpectParam(tokenizer.TokSymbol)
		if err != nil {
			return ast.StatementNone(), err
		}
		setIn := i.Accept(tokenizer.TokKeywordIn)
		if !setIn {
			err = i.Expect(tokenizer.TokSetEqual)
			if err != nil {
				return ast.StatementNone(), err
			}
		}
		expr, err := parseExpression(i, scope)
		if err != nil {
			return ast.StatementNone(), err
		}
		if err := i.Expect(tokenizer.TokNewline); err != nil {
			return ast.StatementNone(), err
		}
		if setIn {
			return ast.StatementSetIn(sym.Str, expr, loc), nil
		} else {
			return ast.StatementSetTo(sym.Str, expr, loc), nil
		}
	} else if i.Accept(tokenizer.TokKeywordDel) {
		expr, err := parseExpression(i, scope)
		if err != nil {
			return ast.StatementNone(), err
		}
		if err := i.Expect(tokenizer.TokNewline); err != nil {
			return ast.StatementNone(), err
		}
		return ast.StatementDel(expr, loc), nil
	} else {
		leftHand, err := parseExpression(i, scope)
		if err != nil {
			return ast.StatementNone(), err
		}
		loc2 := i.Peek().Loc
		if i.Accept(tokenizer.TokNewline) {
			if leftHand.Type != ast.ExprTypeCall && leftHand.Type != ast.ExprTypeNew {
				return ast.StatementNone(), fmt.Errorf("single-expression statement %v instead of call at %v", leftHand, loc)
			}
			return ast.StatementEvaluate(leftHand, loc), nil
		} else if i.Accept(tokenizer.TokLeftShift) {
			rightHand, err := parseExpression(i, scope)
			if err != nil {
				return ast.StatementNone(), err
			}
			if err := i.Expect(tokenizer.TokNewline); err != nil {
				return ast.StatementNone(), err
			}
			return ast.StatementWrite(leftHand, rightHand, loc2), nil
		} else if i.Accept(tokenizer.TokSetEqual) {
			rightHand, err := parseExpression(i, scope)
			if err != nil {
				return ast.StatementNone(), err
			}
			if err := i.Expect(tokenizer.TokNewline); err != nil {
				return ast.StatementNone(), err
			}
			return ast.StatementAssign(leftHand, rightHand, loc2), nil
		} else {
			return ast.StatementNone(), fmt.Errorf("expected top-level operator at %v but got token %v (next afterwards is %v)", loc2, i.Peek(), i.LookAhead(1))
		}
	}
}

func parseStatementBlock(i *input, scope *Scope) ([]ast.Statement, error) {
	err := i.Expect(tokenizer.TokIndent)
	if err != nil {
		return nil, err
	}
	var statements []ast.Statement
	for i.Peek().TokenType != tokenizer.TokUnindent {
		statement, err := parseStatement(i, scope)
		if err != nil {
			return nil, err
		}
		statements = append(statements, statement)
	}
	err = i.Expect(tokenizer.TokUnindent)
	if err != nil {
		return nil, err
	}
	return statements, nil
}
