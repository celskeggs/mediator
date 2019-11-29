package parser

import (
	"fmt"
	"github.com/celskeggs/mediator/dream/path"
	"github.com/celskeggs/mediator/dream/tokenizer"
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
	return tpath, nil
}

func parseExpression(i *input) (DreamMakerExpression, error) {
	if i.Accept(tokenizer.TokStringStart) {
		var subexpressions []DreamMakerExpression
		for !i.Accept(tokenizer.TokStringEnd) {
			if i.Accept(tokenizer.TokStringInsertStart) {
				expr, err := parseExpression(i)
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
				subexpressions = append(subexpressions, ExprStringLiteral(tok.Str))
			}
		}
		if len(subexpressions) == 0 {
			return ExprStringLiteral(""), nil
		} else if len(subexpressions) == 1 {
			return subexpressions[0], nil
		} else {
			return ExprStringConcat(subexpressions), nil
		}
	} else if tok, ok := i.AcceptParam(tokenizer.TokInteger); ok {
		return ExprIntegerLiteral(tok.Int), nil
	} else if tok, ok := i.AcceptParam(tokenizer.TokResource); ok {
		return ExprResourceLiteral(tok.Str), nil
	} else if i.Peek().TokenType == tokenizer.TokSlash {
		tpath, err := parsePath(i)
		if err != nil {
			return ExprNone(), err
		}
		return ExprPathLiteral(tpath), nil
	} else {
		return ExprNone(), fmt.Errorf("invalid token %v when parsing expression at %v", i.Peek(), i.Peek().Loc)
	}
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
	if i.Accept(tokenizer.TokSetEqual) {
		expr, err := parseExpression(i)
		if err != nil {
			return nil, err
		}
		typePath, variable, err := fullPath.SplitLast()
		if err != nil {
			return nil, err
		}
		return []DreamMakerDefinition{
			DefAssign(typePath, variable, expr, loc),
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
