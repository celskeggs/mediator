package tokenizer

import (
	"bufio"
	"fmt"
	"github.com/celskeggs/mediator/util"
	"io"
	"os"
)

type RuneLoc struct {
	Rune rune
	Loc  SourceLocation
}

func FileToRuneChannel(filename string, output chan<- RuneLoc) error {
	defer close(output)
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	reader := bufio.NewReader(f)
	line := 1
	column := 1
	for {
		ch, _, err := reader.ReadRune()
		if err == io.EOF {
			return nil
		} else if err != nil {
			return err
		}
		output <- RuneLoc{
			Rune: ch,
			Loc: SourceLocation{
				File:   filename,
				Line:   line,
				Column: column,
			},
		}
		if ch == '\n' {
			line += 1
			column = 1
		} else if ch == '\t' {
			column += 4
		} else {
			column += 1
		}
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
	runeCh := make(chan RuneLoc)
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
