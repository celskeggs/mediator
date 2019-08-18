package lib

import (
	"unicode"
	"fmt"
	"strings"
	"github.com/celskeggs/mediator/util"
	"github.com/celskeggs/mediator/autocoder/gotype"
	"github.com/celskeggs/mediator/autocoder/iface"
)

func escapeString(str string) string {
	chunks := []string{"\""}
	for _, r := range str {
		chunk := string([]rune{r})
		if r == '"' {
			chunk = "\\\""
		} else if r == '\\' {
			chunk = "\\\\"
		} else if r == '\n' {
			chunk = "\\n"
		} else if !unicode.IsPrint(r) {
			panic(fmt.Sprintf("unimplemented: stringification for rune %d", r))
		}
		chunks = append(chunks, chunk)
	}
	chunks = append(chunks, "\"")
	return strings.Join(chunks, "")
}

func validateIdentifier(name string) {
	util.FIXME("don't use panics for input validation")
	if !isValidIdentifier(name) {
		panic("invalid identifier: " + name)
	}
}

func isValidIdentifier(s string) bool {
	for _, r := range s {
		if !(unicode.IsLetter(r) || r == '_' || unicode.IsNumber(r) || unicode.IsDigit(r)) {
			return false
		}
	}
	return true
}

func stringTypes(goTypes []gotype.Type) string {
	var parts []string
	for _, goType := range goTypes {
		parts = append(parts, goType.String())
	}
	return strings.Join(parts, ", ")
}

func commaSeparated(exprs []iface.Expr) string {
	strArgs := make([]string, len(exprs))
	for i, arg := range exprs {
		strArgs[i] = arg.(expression).String()
	}
	return strings.Join(strArgs, ", ")
}
