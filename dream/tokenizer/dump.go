package tokenizer

import (
	"io"
	"github.com/celskeggs/mediator/util"
	"os"
	"bufio"
	"fmt"
)

func FileToRuneChannel(filename string, output chan <-rune) error {
	defer close(output)
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	reader := bufio.NewReader(f)
	for {
		ch, _, err := reader.ReadRune()
		if err == io.EOF {
			return nil
		} else if err != nil {
			return err
		}
		output <- ch
	}
}

const BeginTokens = "[begining of token stream]"
const EndTokens = "[end of token stream]"

func DumpTokens(tokens <-chan Token, output io.Writer) error {
	defer func() {
		for range tokens {
			// drain input
		}
	}()
	_, err := fmt.Fprintln(output, BeginTokens)
	if err != nil {
		return err
	}
	for token := range tokens {
		_, err = fmt.Fprintln(output, token.String())
		if err != nil {
			return err
		}
	}
	_, err = fmt.Fprintln(output, EndTokens)
	if err != nil {
		return err
	}
	return nil
}

func DumpTokensFromFile(filename string, output io.Writer) error {
	runeCh := make(chan rune)
	tokenCh := make(chan Token)
	indentedCh := make(chan Token)
	return util.RunInParallel(
		func() error {
			return FileToRuneChannel(filename, runeCh)
		},
		func() error {
			return Tokenize(runeCh, tokenCh)
		},
		func() error {
			return ProcessIndentation(tokenCh, indentedCh)
		},
		func() error {
			return DumpTokens(indentedCh, output)
		},
	)
}
