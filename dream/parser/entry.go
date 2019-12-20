package parser

import (
	"fmt"
	"github.com/celskeggs/mediator/dream/ast"
	"github.com/celskeggs/mediator/dream/preprocessor"
	"github.com/celskeggs/mediator/dream/tokenizer"
	"github.com/celskeggs/mediator/util"
	"github.com/pkg/errors"
	"io"
	"os"
)

type input struct {
	Channel    <-chan tokenizer.Token
	NextTokens []tokenizer.Token
}

func (i *input) HasLookahead(n int) bool {
	for len(i.NextTokens) <= n {
		next, ok := <-i.Channel
		if !ok {
			return false
		}
		if next.IsNone() {
			panic("should not have gotten TokNone")
		}
		i.NextTokens = append(i.NextTokens, next)
	}
	return true
}

func (i *input) HasNext() bool {
	return i.HasLookahead(0)
}

func (i *input) LookAhead(n int) tokenizer.Token {
	if !i.HasLookahead(n) {
		return tokenizer.NoToken()
	}
	return i.NextTokens[n]
}

func (i *input) Peek() tokenizer.Token {
	return i.LookAhead(0)
}

func (i *input) Consume() {
	if !i.HasNext() {
		panic("consume when no tokens are available")
	}
	copy(i.NextTokens[:len(i.NextTokens)-1], i.NextTokens[1:])
	i.NextTokens = i.NextTokens[:len(i.NextTokens)-1]
}

func (i *input) Take() tokenizer.Token {
	token := i.Peek()
	if !token.IsNone() {
		i.Consume()
	}
	return token
}

func (i *input) AcceptParam(tokenType tokenizer.TokenType) (tokenizer.Token, bool) {
	if i.Peek().TokenType == tokenType {
		return i.Take(), true
	}
	return i.Peek(), false
}

func (i *input) Accept(tokenType tokenizer.TokenType) bool {
	_, ok := i.AcceptParam(tokenType)
	return ok
}

func (i *input) ErrorExpect(tokenType tokenizer.TokenType) error {
	tok := i.Take()
	ntok := i.Take()
	return fmt.Errorf("expected token of type %v at %v but got token %v (next afterwards is %v)", tokenType, tok.Loc, tok, ntok)
}

func (i *input) Expect(tokenType tokenizer.TokenType) error {
	_, err := i.ExpectParam(tokenType)
	return err
}

func (i *input) ExpectParam(tokenType tokenizer.TokenType) (tokenizer.Token, error) {
	tok, ok := i.AcceptParam(tokenType)
	if !ok {
		return tokenizer.NoToken(), i.ErrorExpect(tokenType)
	}
	return tok, nil
}

func (i *input) AcceptAll(tokenType tokenizer.TokenType) (count uint) {
	for i.Accept(tokenType) {
		count += 1
	}
	return count
}

func ParseDM(tokens <-chan tokenizer.Token) (*ast.File, error) {
	defer func() {
		for range tokens {
			// drain input
		}
	}()
	input := &input{tokens, nil}
	return parseFile(input)
}

type ParseContext struct {
	parallel *util.ParallelElements
}

func NewParseContext() *ParseContext {
	return &ParseContext{
		parallel: util.NewParallel(),
	}
}

func (p *ParseContext) LoadTokens(filename string) <-chan tokenizer.Token {
	runeCh := make(chan tokenizer.RuneLoc)
	tokenCh := make(chan tokenizer.Token)
	indentedCh := make(chan tokenizer.Token)
	p.parallel.Add(func() error {
		return errors.Wrapf(tokenizer.FileToRuneChannel(filename, runeCh), "while reading %q", filename)
	})
	p.parallel.Add(func() error {
		return errors.Wrapf(tokenizer.Tokenize(runeCh, tokenCh), "while tokenizing %q", filename)
	})
	p.parallel.Add(func() error {
		return errors.Wrapf(tokenizer.ProcessIndentation(tokenCh, indentedCh), "while deindenting %q", filename)
	})
	return indentedCh
}

func ParseFile(filename string) (dmf *ast.File, err error) {
	context := NewParseContext()
	tokenCh := make(chan tokenizer.Token)

	var searchpath, maps []string

	context.parallel.Add(func() error {
		sp, m, err := preprocessor.Preprocess(context.LoadTokens, filename, tokenCh)
		searchpath = sp
		maps = m
		return err
	})
	context.parallel.Add(func() error {
		parsed, err := ParseDM(tokenCh)
		if err != nil {
			return errors.Wrap(err, "while parsing DM code")
		}
		dmf = parsed
		return nil
	})
	err = context.parallel.Join()
	if err != nil {
		return nil, err
	}
	dmf.SearchPath = searchpath
	dmf.Maps = maps
	return dmf, nil
}

func ParseFiles(filenames []string) (total *ast.File, err error) {
	total = &ast.File{
		Definitions: nil,
	}
	for _, file := range filenames {
		single, err := ParseFile(file)
		if err != nil {
			return nil, err
		}
		total.Extend(single)
	}
	total.Dump(os.Stdout)
	return total, nil
}

func DumpParsedFile(filename string, output io.Writer) error {
	dmf, err := ParseFile(filename)
	if err != nil {
		return err
	}
	err = dmf.Dump(output)
	if err != nil {
		return err
	}
	return nil
}
