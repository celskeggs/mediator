package parser

import (
	"github.com/celskeggs/mediator/dream/tokenizer"
	"fmt"
	"github.com/celskeggs/mediator/util"
)

type input struct {
	Channel   <-chan tokenizer.Token
	NextToken tokenizer.Token
}

func (i *input) HasNext() bool {
	if i.NextToken.IsNone() {
		next, ok := <-i.Channel
		if !ok {
			return false
		}
		if next.IsNone() {
			panic("should not have gotten TokNone")
		}
		i.NextToken = next
	}
	return true
}

func (i *input) Peek() tokenizer.Token {
	if !i.HasNext() {
		return tokenizer.NoToken()
	}
	return i.NextToken
}

func (i *input) Consume() {
	if !i.HasNext() {
		panic("consume when no tokens are available")
	}
	i.NextToken = tokenizer.NoToken()
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
	return fmt.Errorf("expected token of type %v but got token %v (next afterwards is %v)", tokenType, tok, ntok)
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

func ParseDM(tokens <-chan tokenizer.Token) (*DreamMakerFile, error) {
	defer func() {
		for range tokens {
			// drain input
		}
	}()
	input := &input{tokens, tokenizer.NoToken()}
	return parseFile(input)
}

func ParseFile(filename string) (*DreamMakerFile, error) {
	runeCh := make(chan rune)
	tokenCh := make(chan tokenizer.Token)
	indentedCh := make(chan tokenizer.Token)
	var dmf *DreamMakerFile

	err := util.RunInParallel(
		func() error {
			return tokenizer.FileToRuneChannel(filename, runeCh)
		},
		func() error {
			return tokenizer.Tokenize(runeCh, tokenCh)
		},
		func() error {
			return tokenizer.ProcessIndentation(tokenCh, indentedCh)
		},
		func() error {
			parsed, err := ParseDM(indentedCh)
			if err != nil {
				return err
			}
			dmf = parsed
			return nil
		},
	)
	if err != nil {
		return nil, err
	}
	return dmf, nil
}
