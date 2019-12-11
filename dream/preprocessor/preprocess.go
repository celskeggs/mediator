package preprocessor

import (
	"fmt"
	"github.com/celskeggs/mediator/dream/tokenizer"
	"github.com/celskeggs/mediator/util"
	"strings"
)

const SearchPathSymbol = "FILE_DIR"

// nonexistent files should simply close immediately
type FileLoader func(name string) <-chan tokenizer.Token

func tokensUntilEOL(c <-chan tokenizer.Token) ([]tokenizer.Token, bool) {
	var tokens []tokenizer.Token
	for ch := range c {
		if ch.TokenType == tokenizer.TokNewline {
			return tokens, true
		}
		tokens = append(tokens, ch)
	}
	return nil, false
}

func constantString(c <-chan tokenizer.Token, statement string, after tokenizer.SourceLocation) (string, error) {
	keyword, ok := <-c
	if !ok {
		return "", fmt.Errorf("expected string immediately after %s, not EOF after %v", statement, after)
	}
	if keyword.TokenType == tokenizer.TokDot && statement == "#define" {
		util.FIXME("find a less hacky way to handle #define FILE_PATH .")
		return ".", nil
	}
	if keyword.TokenType != tokenizer.TokStringStart {
		return "", fmt.Errorf("expected string immediately after %s, not %v at %v", statement, keyword, keyword.Loc)
	}
	after = keyword.Loc
	var parts []string
	for ch := range c {
		if ch.TokenType == tokenizer.TokStringEnd {
			return strings.Join(parts, ""), nil
		} else if ch.TokenType == tokenizer.TokStringLiteral {
			parts = append(parts, ch.Str)
		} else {
			return "", fmt.Errorf("unexpected token %v during %s string at %v", ch, statement, ch.Loc)
		}
		after = ch.Loc
	}
	return "", fmt.Errorf("expected string immediately after %s, not EOF after %v", statement, after)
}

func Preprocess(load FileLoader, filename string, output chan<- tokenizer.Token) (searchpath []string, maps []string, err error) {
	channels := []<-chan tokenizer.Token{load(filename)}
	defer func() {
		close(output)
		for _, ch := range channels {
			for range ch {
				// drain the rest of the input
			}
		}
	}()
	definitions := map[string][]tokenizer.Token{}
	for len(channels) > 0 {
		ch := channels[len(channels)-1]
		token, ok := <-ch
		if !ok {
			channels = channels[:len(channels)-1]
			continue
		}
		switch token.TokenType {
		case tokenizer.TokPreprocessorDefine:
			keyword, ok := <-ch
			if !ok {
				return nil, nil, fmt.Errorf("expected symbol immediately after #define, not EOF at %v", token.Loc)
			} else if keyword.TokenType != tokenizer.TokSymbol {
				return nil, nil, fmt.Errorf("expected symbol immediately after #define, not %v at %v", keyword, keyword.Loc)
			}
			if keyword.Str == SearchPathSymbol {
				subfile, err := constantString(ch, "#define", token.Loc)
				if err != nil {
					return nil, nil, err
				}
				searchpath = append(searchpath, subfile)
			} else {
				body, ok := tokensUntilEOL(ch)
				if !ok {
					return nil, nil, fmt.Errorf("ran out of tokens while trying to find newline after #define at %v", token.Loc)
				}
				if _, exists := definitions[keyword.Str]; exists {
					return nil, nil, fmt.Errorf("attempt to re-#define symbol %q at %v", keyword.Str, token.Loc)
				}
				definitions[keyword.Str] = body
			}
		case tokenizer.TokPreprocessorInclude:
			subfile, err := constantString(ch, "#include", token.Loc)
			if err != nil {
				return nil, nil, err
			}
			if strings.HasSuffix(subfile, ".dmm") {
				maps = append(maps, subfile)
			} else {
				channels = append(channels, load(subfile))
			}
		case tokenizer.TokSymbol:
			def, found := definitions[token.Str]
			if found {
				for _, tok := range def {
					output <- tok
				}
			} else {
				output <- token
			}
		default:
			output <- token
		}
	}
	return searchpath, maps, nil
}
