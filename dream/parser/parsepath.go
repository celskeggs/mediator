package parser

import (
	"fmt"
	"github.com/celskeggs/mediator/dream/declpath"
	"github.com/celskeggs/mediator/dream/path"
	"github.com/celskeggs/mediator/dream/tokenizer"
)

func parsePath(i *input) (path.TypePath, error) {
	loc := i.Peek().Loc
	decl, err := parseDeclPath(i)
	if err != nil {
		return path.Empty(), err
	}
	if !decl.IsPlain() {
		return path.Empty(), fmt.Errorf("expected type path, not decl path %v, at %v", decl, loc)
	}
	return decl.Unwrap(), nil
}

func convertDeclSegment(tok tokenizer.TokenType) declpath.DeclType {
	switch tok {
	case tokenizer.TokSymbol:
		return declpath.DeclPlain
	case tokenizer.TokKeywordVar:
		return declpath.DeclVar
	case tokenizer.TokKeywordProc:
		return declpath.DeclProc
	case tokenizer.TokKeywordVerb:
		return declpath.DeclVerb
	default:
		return declpath.DeclInvalid
	}
}

func parseDeclPath(i *input) (declpath.DeclPath, error) {
	tpath := declpath.Empty()
	i.AcceptAll(tokenizer.TokNewline)
	if i.Accept(tokenizer.TokSlash) {
		tpath = declpath.Root()
	}
	if convertDeclSegment(i.Peek().TokenType) == declpath.DeclInvalid {
		return declpath.Empty(), fmt.Errorf("invalid token %v when looking for path at %v", i.Peek(), i.Peek().Loc)
	}
	for {
		declType := convertDeclSegment(i.Peek().TokenType)
		if declType == declpath.DeclInvalid {
			break
		} else if declType == declpath.DeclPlain {
			tok := i.Take()
			if !tpath.CanAdd() {
				return declpath.Empty(), fmt.Errorf("path %v is already complete and cannot be extended at %v", tpath, tok.Loc)
			}
			tpath = tpath.Add(tok.Str)
		} else {
			tok := i.Take()
			if !tpath.CanAddDecl() {
				return declpath.Empty(), fmt.Errorf("path %v is already complete and cannot be extended at %v", tpath, tok.Loc)
			}
			tpath = tpath.AddDecl(declType)
		}
		if !i.Accept(tokenizer.TokSlash) {
			break
		}
	}
	if tpath.IsEmpty() {
		return declpath.Empty(), fmt.Errorf("expected a path at %v", i.Peek().Loc)
	}
	return tpath, nil
}
