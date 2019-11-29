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
	loc := i.Peek().Loc
	if i.Peek().TokenType == tokenizer.TokParenOpen {
		keywords, exprs, err := parseExpressionArguments(i, variables)
		if err != nil {
			return ExprNone(), err
		}
		return ExprCall(expr, keywords, exprs, loc), nil
	}
	return expr, nil
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
		println("started parsing if block")
		statements, err := parseStatementBlock(i, variables)
		if err != nil {
			return StatementNone(), err
		}
		println("finished parsing if block")
		return StatementIf(condition, statements, loc), nil
	} else if i.Accept(tokenizer.TokKeywordReturn) {
		if err := i.Expect(tokenizer.TokNewline); err != nil {
			return StatementNone(), err
		}
		util.FIXME("support returning values")
		return StatementReturn(loc), nil
	} else {
		leftHand, err := parseExpression(i, variables)
		if err != nil {
			return StatementNone(), err
		}
		loc := i.Peek().Loc
		err = i.Expect(tokenizer.TokLeftShift)
		if err != nil {
			return StatementNone(), err
		}
		rightHand, err := parseExpression(i, variables)
		if err != nil {
			return StatementNone(), err
		}
		if err := i.Expect(tokenizer.TokNewline); err != nil {
			return StatementNone(), err
		}
		return StatementWrite(leftHand, rightHand, loc), nil
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
		})
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
