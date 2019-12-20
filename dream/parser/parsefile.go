package parser

import (
	"fmt"
	"github.com/celskeggs/mediator/autocoder/dtype"
	"github.com/celskeggs/mediator/dream/ast"
	"github.com/celskeggs/mediator/dream/declpath"
	"github.com/celskeggs/mediator/dream/path"
	"github.com/celskeggs/mediator/dream/tokenizer"
)

func parseFunctionArguments(i *input) ([]ast.TypedName, error) {
	if err := i.Expect(tokenizer.TokParenOpen); err != nil {
		return nil, err
	}
	if i.Accept(tokenizer.TokParenClose) {
		return nil, nil
	}
	var args []ast.TypedName
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
		args = append(args, ast.TypedName{
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

func parseFunctionBody(i *input, srcType path.TypePath, arguments []ast.TypedName) ([]ast.Statement, error) {
	variables := make([]ast.TypedName, len(arguments))
	copy(variables, arguments)
	variables = append(variables,
		ast.TypedName{
			Type: dtype.Path(srcType),
			Name: "src",
		},
		ast.TypedName{
			Type: dtype.ConstPath("/mob"),
			Name: "usr",
		},
	)
	return parseStatementBlock(i, variables)
}

func parseBlock(i *input, basePath declpath.DeclPath) ([]ast.Definition, error) {
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
			return []ast.Definition{
				ast.DefVarDef(varTarget, varType, varName, loc),
				ast.DefAssign(varTarget, varName, expr, loc),
			}, nil
		} else if i.Accept(tokenizer.TokNewline) {
			return []ast.Definition{
				ast.DefVarDef(varTarget, varType, varName, loc),
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
			return []ast.Definition{
				ast.DefVerbDecl(procTarget, procName, loc),
				ast.DefImplement(procTarget, procName, args, body, loc),
			}, nil
		} else {
			return []ast.Definition{
				ast.DefProcDecl(procTarget, procName, loc),
				ast.DefImplement(procTarget, procName, args, body, loc),
			}, nil
		}
	} else if !fullPath.IsPlain() {
		// incomplete definition, because we're not a plain path but not a full decl, so we need to get one more entry
		if i.Accept(tokenizer.TokNewline) {
			// nothing to define
			return nil, nil
		} else if i.Accept(tokenizer.TokIndent) {
			var defs []ast.Definition
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
		return []ast.Definition{
			ast.DefAssign(typePath, variable, expr, loc),
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
		return []ast.Definition{
			ast.DefImplement(typePath, function, args, body, loc),
		}, nil
	} else if i.Accept(tokenizer.TokNewline) {
		return []ast.Definition{
			ast.DefDefine(plainPath, loc),
		}, nil
	} else if i.Accept(tokenizer.TokIndent) {
		defs := []ast.Definition{
			ast.DefDefine(plainPath, loc),
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

func parseFile(i *input) (*ast.File, error) {
	var allDefs []ast.Definition
	for i.HasNext() {
		defs, err := parseBlock(i, declpath.Root())
		if err != nil {
			return nil, err
		}
		allDefs = append(allDefs, defs...)
	}
	return &ast.File{
		Definitions: allDefs,
	}, nil
}
