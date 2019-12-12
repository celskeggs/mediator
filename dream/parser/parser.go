package parser

import (
	"fmt"
	"github.com/celskeggs/mediator/dream/path"
	"github.com/celskeggs/mediator/dream/tokenizer"
	"github.com/celskeggs/mediator/util"
)

func parsePath(i *input) (path.TypePath, error) {
	tpath := path.Empty()
	i.AcceptAll(tokenizer.TokNewline)
	if i.Accept(tokenizer.TokSlash) {
		tpath = path.Root()
	}
	if i.Peek().TokenType != tokenizer.TokSymbol {
		return path.Empty(), i.ErrorExpect(tokenizer.TokSymbol)
	}
	for {
		tok, ok := i.AcceptParam(tokenizer.TokSymbol)
		if !ok {
			break
		}
		tpath = tpath.Add(tok.Str)
		if !i.Accept(tokenizer.TokSlash) {
			break
		}
	}
	if tpath.IsEmpty() {
		return path.Empty(), fmt.Errorf("expected a path at %v", i.Peek().Loc)
	}
	return tpath, nil
}

func parseExpression0(i *input, variables []DreamMakerTypedName) (DreamMakerExpression, error) {
	loc := i.Peek().Loc
	if i.Accept(tokenizer.TokStringStart) {
		var subexpressions []DreamMakerExpression
		for !i.Accept(tokenizer.TokStringEnd) {
			if i.Accept(tokenizer.TokStringInsertStart) {
				expr, err := parseExpression(i, variables)
				if err != nil {
					return ExprNone(), err
				}
				subexpressions = append(subexpressions, expr)
				if err := i.Expect(tokenizer.TokStringInsertEnd); err != nil {
					return ExprNone(), err
				}
			} else {
				tok, err := i.ExpectParam(tokenizer.TokStringLiteral)
				if err != nil {
					return ExprNone(), err
				}
				subexpressions = append(subexpressions, ExprStringLiteral(tok.Str, tok.Loc))
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
			expr = ExprGetField(expr, field.Str, loc)
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

func parseFunctionArguments(i *input) ([]DreamMakerTypedName, error) {
	if i.Accept(tokenizer.TokParenClose) {
		return nil, nil
	}
	var args []DreamMakerTypedName
	for {
		loc := i.Peek().Loc
		declPath, err := parsePath(i)
		if err != nil {
			return nil, err
		}
		if declPath.IsAbsolute {
			return nil, fmt.Errorf("invalid use of absolute path at %v", loc)
		}
		// these paths are specified without a leading slash, but are actually absolute
		declPath = path.Root().Join(declPath)

		typePath, varName, err := declPath.SplitLast()
		if err != nil {
			return nil, err
		}
		args = append(args, DreamMakerTypedName{
			Type: typePath,
			Name: varName,
		})
		if !i.Accept(tokenizer.TokComma) {
			break
		}
	}
	err := i.Expect(tokenizer.TokParenClose)
	if err != nil {
		return nil, err
	}
	return args, nil
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
		typepath, err := parsePath(i)
		if err != nil {
			return StatementNone(), err
		}
		if typepath.IsAbsolute {
			return StatementNone(), fmt.Errorf("unexpected absolute path at %v", loc2)
		}
		if !typepath.StartsWith("var") {
			return StatementNone(), fmt.Errorf("unexpected non-var path at %v", loc2)
		}
		_, rest, err := typepath.SplitFirst()
		if err != nil {
			return StatementNone(), err
		}
		typepath, varname, err := rest.SplitLast()
		if err != nil {
			return StatementNone(), err
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
		body, err := parseStatementBlock(i, variables)
		if err != nil {
			return StatementNone(), err
		}
		return StatementForList(typepath, varname, inExpr, body, loc), nil
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

func parseFunctionBody(i *input, srcType path.TypePath, arguments []DreamMakerTypedName) ([]DreamMakerStatement, error) {
	variables := make([]DreamMakerTypedName, len(arguments))
	copy(variables, arguments)
	variables = append(variables,
		DreamMakerTypedName{
			Type: srcType,
			Name: "src",
		},
		DreamMakerTypedName{
			Type: path.ConstTypePath("/mob"),
			Name: "usr",
		},
	)
	return parseStatementBlock(i, variables)
}

func parseBlock(i *input, basePath path.TypePath) ([]DreamMakerDefinition, error) {
	if i.Accept(tokenizer.TokNewline) {
		return nil, nil
	}
	loc := i.Peek().Loc
	relPath, err := parsePath(i)
	if err != nil {
		return nil, err
	}
	fullPath := basePath.Join(relPath)
	err = fullPath.CheckKeywords()
	if err != nil {
		return nil, err
	}
	if fullPath.IsVarDef() {
		varTarget, varName := fullPath.SplitVarDef()
		if i.Accept(tokenizer.TokSetEqual) {
			// no variables because there's no function context during initializations
			expr, err := parseExpression(i, nil)
			if err != nil {
				return nil, err
			}
			return []DreamMakerDefinition{
				DefVarDef(varTarget, varName, loc),
				DefAssign(varTarget, varName, expr, loc),
			}, nil
		} else if i.Accept(tokenizer.TokNewline) {
			return []DreamMakerDefinition{
				DefVarDef(varTarget, varName, loc),
			}, nil
		} else {
			return nil, fmt.Errorf("expected valid start-var token, not %s at %v", i.Peek().String(), i.Peek().Loc)
		}
	} else if fullPath.EndsWith("var") {
		if i.Accept(tokenizer.TokNewline) {
			// nothing to define
			return nil, nil
		} else if i.Accept(tokenizer.TokIndent) {
			var defs []DreamMakerDefinition
			for !i.Accept(tokenizer.TokUnindent) {
				block, err := parseBlock(i, fullPath)
				if err != nil {
					return nil, err
				}
				defs = append(defs, block...)
			}
			return defs, nil
		} else {
			return nil, fmt.Errorf("expected valid start-of-var-block token, not %s at %v", i.Peek().String(), i.Peek().Loc)
		}
	}
	if i.Accept(tokenizer.TokSetEqual) {
		// no variables because there's no function context during initializations
		expr, err := parseExpression(i, nil)
		if err != nil {
			return nil, err
		}
		typePath, variable, err := fullPath.SplitLast()
		if err != nil {
			return nil, err
		}
		if len(typePath.Segments) == 0 {
			return nil, fmt.Errorf("cannot assign variable on root at %v", loc)
		}
		return []DreamMakerDefinition{
			DefAssign(typePath, variable, expr, loc),
		}, nil
	} else if i.Accept(tokenizer.TokParenOpen) {
		args, err := parseFunctionArguments(i)
		if err != nil {
			return nil, err
		}
		typePath, function, err := fullPath.SplitLast()
		if err != nil {
			return nil, err
		}
		body, err := parseFunctionBody(i, typePath, args)
		if err != nil {
			return nil, err
		}
		if len(typePath.Segments) == 0 {
			return nil, fmt.Errorf("cannot implement function on root at %v", loc)
		}
		return []DreamMakerDefinition{
			DefImplement(typePath, function, args, body, loc),
		}, nil
	} else if i.Accept(tokenizer.TokNewline) {
		return []DreamMakerDefinition{
			DefDefine(fullPath, loc),
		}, nil
	} else if i.Accept(tokenizer.TokIndent) {
		defs := []DreamMakerDefinition{
			DefDefine(fullPath, loc),
		}
		for !i.Accept(tokenizer.TokUnindent) {
			block, err := parseBlock(i, fullPath)
			if err != nil {
				return nil, err
			}
			defs = append(defs, block...)
		}
		return defs, nil
	} else {
		return nil, fmt.Errorf("expected valid start-of-block token, not %s at %v", i.Peek().String(), i.Peek().Loc)
	}
}

func parseFile(i *input) (*DreamMakerFile, error) {
	var allDefs []DreamMakerDefinition
	for i.HasNext() {
		defs, err := parseBlock(i, path.Root())
		if err != nil {
			return nil, err
		}
		allDefs = append(allDefs, defs...)
	}
	return &DreamMakerFile{
		Definitions: allDefs,
	}, nil
}
