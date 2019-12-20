package parser

import (
	"fmt"
	"github.com/celskeggs/mediator/autocoder/dtype"
	"github.com/celskeggs/mediator/dream/declpath"
	"github.com/celskeggs/mediator/dream/path"
	"github.com/celskeggs/mediator/dream/tokenizer"
)

func parseFunctionArguments(i *input) ([]DreamMakerTypedName, error) {
	if err := i.Expect(tokenizer.TokParenOpen); err != nil {
		return nil, err
	}
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
			Type: dtype.FromPath(typePath),
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

func parseFunctionBody(i *input, srcType path.TypePath, arguments []DreamMakerTypedName) ([]DreamMakerStatement, error) {
	variables := make([]DreamMakerTypedName, len(arguments))
	copy(variables, arguments)
	variables = append(variables,
		DreamMakerTypedName{
			Type: dtype.Path(srcType),
			Name: "src",
		},
		DreamMakerTypedName{
			Type: dtype.ConstPath("/mob"),
			Name: "usr",
		},
	)
	return parseStatementBlock(i, variables)
}

func parseBlock(i *input, basePath declpath.DeclPath) ([]DreamMakerDefinition, error) {
	if i.Accept(tokenizer.TokNewline) {
		return nil, nil
	}
	loc := i.Peek().Loc
	relPath, err := parseDeclPath(i)
	if err != nil {
		return nil, err
	}
	fullPath, ok := basePath.Join(relPath)
	if !ok {
		return nil, fmt.Errorf("cannot join paths %v and %v at %v", basePath, relPath, loc)
	}
	if fullPath.IsVarDef() {
		varTarget, varType, varName := fullPath.SplitDef()
		if i.Accept(tokenizer.TokSetEqual) {
			// no variables because there's no function context during initializations
			expr, err := parseExpression(i, nil)
			if err != nil {
				return nil, err
			}
			return []DreamMakerDefinition{
				DefVarDef(varTarget, varType, varName, loc),
				DefAssign(varTarget, varName, expr, loc),
			}, nil
		} else if i.Accept(tokenizer.TokNewline) {
			return []DreamMakerDefinition{
				DefVarDef(varTarget, varType, varName, loc),
			}, nil
		} else {
			return nil, fmt.Errorf("expected valid start-var token, not %s at %v", i.Peek().String(), i.Peek().Loc)
		}
	} else if fullPath.IsProcDef() || fullPath.IsVerbDef() {
		procTarget, _, procName := fullPath.SplitDef()
		args, err := parseFunctionArguments(i)
		if err != nil {
			return nil, err
		}
		body, err := parseFunctionBody(i, procTarget, args)
		if err != nil {
			return nil, err
		}
		if len(procTarget.Segments) == 0 {
			return nil, fmt.Errorf("cannot declare function on root at %v", loc)
		}
		if fullPath.IsVerbDef() {
			return []DreamMakerDefinition{
				DefVerbDecl(procTarget, procName, loc),
				DefImplement(procTarget, procName, args, body, loc),
			}, nil
		} else {
			return []DreamMakerDefinition{
				DefProcDecl(procTarget, procName, loc),
				DefImplement(procTarget, procName, args, body, loc),
			}, nil
		}
	} else if !fullPath.IsPlain() {
		// incomplete definition, because we're not a plain path but not a full decl, so we need to get one more entry
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
	// in this case, we just have a plain path, so we can accept a lot of things
	plainPath := fullPath.Unwrap()
	if i.Accept(tokenizer.TokSetEqual) {
		// no variables because there's no function context during initializations
		expr, err := parseExpression(i, nil)
		if err != nil {
			return nil, err
		}
		typePath, variable, err := plainPath.SplitLast()
		if err != nil {
			return nil, err
		}
		if len(typePath.Segments) == 0 {
			return nil, fmt.Errorf("cannot assign variable on root at %v", loc)
		}
		return []DreamMakerDefinition{
			DefAssign(typePath, variable, expr, loc),
		}, nil
	} else if i.Peek().TokenType == tokenizer.TokParenOpen {
		args, err := parseFunctionArguments(i)
		if err != nil {
			return nil, err
		}
		typePath, function, err := plainPath.SplitLast()
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
			DefDefine(plainPath, loc),
		}, nil
	} else if i.Accept(tokenizer.TokIndent) {
		defs := []DreamMakerDefinition{
			DefDefine(plainPath, loc),
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
		defs, err := parseBlock(i, declpath.Root())
		if err != nil {
			return nil, err
		}
		allDefs = append(allDefs, defs...)
	}
	return &DreamMakerFile{
		Definitions: allDefs,
	}, nil
}
