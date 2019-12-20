package parser

import (
	"fmt"
	"github.com/celskeggs/mediator/autocoder/dtype"
	"github.com/celskeggs/mediator/dream/tokenizer"
	"github.com/celskeggs/mediator/util"
)

func addVar(vars []DreamMakerTypedName, varType dtype.DType, varName string) []DreamMakerTypedName {
	result := make([]DreamMakerTypedName, len(vars)+1)
	copy(result, vars)
	result[len(vars)] = DreamMakerTypedName{
		Type: varType,
		Name: varName,
	}
	return result
}

func parseStatement(i *input, variables []DreamMakerTypedName) (DreamMakerStatement, error) {
	loc := i.Peek().Loc
	if i.Accept(tokenizer.TokKeywordIf) {
		if err := i.Expect(tokenizer.TokParenOpen); err != nil {
			return StatementNone(), err
		}
		condition, err := parseExpression(i, variables)
		if err != nil {
			return StatementNone(), err
		}
		if err := i.Expect(tokenizer.TokParenClose); err != nil {
			return StatementNone(), err
		}
		statements, err := parseStatementBlock(i, variables)
		if err != nil {
			return StatementNone(), err
		}
		return StatementIf(condition, statements, loc), nil
	} else if i.Accept(tokenizer.TokKeywordFor) {
		if err := i.Expect(tokenizer.TokParenOpen); err != nil {
			return StatementNone(), err
		}
		loc2 := i.Peek().Loc
		varpath, err := parseDeclPath(i)
		if err != nil {
			return StatementNone(), err
		}
		if !varpath.IsVarDef() {
			return StatementNone(), fmt.Errorf("path %v is not a variable definition at %v", varpath, loc2)
		}
		varTarget, varTypePath, varName := varpath.SplitDef()
		varType := dtype.FromPath(varTypePath)
		if !varTarget.IsEmpty() {
			return StatementNone(), fmt.Errorf("invalid prefix for 'var' in path %v at %v", varpath, loc2)
		}
		if i.Peek().TokenType == tokenizer.TokKeywordAs {
			return StatementNone(), fmt.Errorf("unsupported: keyword as in for loop at %v", i.Peek().Loc)
		}
		var inExpr DreamMakerExpression
		if i.Accept(tokenizer.TokKeywordIn) {
			if inExpr, err = parseExpression(i, variables); err != nil {
				return StatementNone(), err
			}
		}
		if err := i.Expect(tokenizer.TokParenClose); err != nil {
			return StatementNone(), err
		}
		body, err := parseStatementBlock(i, addVar(variables, varType, varName))
		if err != nil {
			return StatementNone(), err
		}
		return StatementForList(varType, varName, inExpr, body, loc), nil
	} else if i.Accept(tokenizer.TokKeywordReturn) {
		if err := i.Expect(tokenizer.TokNewline); err != nil {
			return StatementNone(), err
		}
		util.FIXME("support returning values")
		return StatementReturn(loc), nil
	} else if i.Accept(tokenizer.TokKeywordSet) {
		sym, err := i.ExpectParam(tokenizer.TokSymbol)
		if err != nil {
			return StatementNone(), err
		}
		setIn := i.Accept(tokenizer.TokKeywordIn)
		if !setIn {
			err = i.Expect(tokenizer.TokSetEqual)
			if err != nil {
				return StatementNone(), err
			}
		}
		expr, err := parseExpression(i, variables)
		if err != nil {
			return StatementNone(), err
		}
		if err := i.Expect(tokenizer.TokNewline); err != nil {
			return StatementNone(), err
		}
		if setIn {
			return StatementSetIn(sym.Str, expr, loc), nil
		} else {
			return StatementSetTo(sym.Str, expr, loc), nil
		}
	} else if i.Accept(tokenizer.TokKeywordDel) {
		expr, err := parseExpression(i, variables)
		if err != nil {
			return StatementNone(), err
		}
		if err := i.Expect(tokenizer.TokNewline); err != nil {
			return StatementNone(), err
		}
		return StatementDel(expr, loc), nil
	} else {
		leftHand, err := parseExpression(i, variables)
		if err != nil {
			return StatementNone(), err
		}
		loc2 := i.Peek().Loc
		if i.Accept(tokenizer.TokNewline) {
			if leftHand.Type != ExprTypeCall && leftHand.Type != ExprTypeNew {
				return StatementNone(), fmt.Errorf("single-expression statement %v instead of call at %v", leftHand, loc)
			}
			return StatementEvaluate(leftHand, loc), nil
		} else if i.Accept(tokenizer.TokLeftShift) {
			rightHand, err := parseExpression(i, variables)
			if err != nil {
				return StatementNone(), err
			}
			if err := i.Expect(tokenizer.TokNewline); err != nil {
				return StatementNone(), err
			}
			return StatementWrite(leftHand, rightHand, loc2), nil
		} else if i.Accept(tokenizer.TokSetEqual) {
			rightHand, err := parseExpression(i, variables)
			if err != nil {
				return StatementNone(), err
			}
			if err := i.Expect(tokenizer.TokNewline); err != nil {
				return StatementNone(), err
			}
			return StatementAssign(leftHand, rightHand, loc2), nil
		} else {
			return StatementNone(), fmt.Errorf("expected top-level operator at %v but got token %v (next afterwards is %v)", loc2, i.Peek(), i.LookAhead(1))
		}
	}
}

func parseStatementBlock(i *input, variables []DreamMakerTypedName) ([]DreamMakerStatement, error) {
	err := i.Expect(tokenizer.TokIndent)
	if err != nil {
		return nil, err
	}
	var statements []DreamMakerStatement
	for i.Peek().TokenType != tokenizer.TokUnindent {
		statement, err := parseStatement(i, variables)
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
