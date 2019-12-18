package format

import (
	"fmt"
	"github.com/celskeggs/mediator/platform/types"
	"github.com/celskeggs/mediator/util"
	"strings"
	"unicode"
)

func isUpperCase(s string) bool {
	for _, r := range s {
		return unicode.IsUpper(r)
	}
	return false
}

func FormatMacro(macro string, atom types.Value) string {
	util.FIXME("support more types of non-atoms")
	if s, ok := atom.(types.String); ok {
		return types.Unstring(s)
	} else if macro == "the" || macro == "The" {
		name := types.Unstring(atom.Var("name"))
		if isUpperCase(name) {
			return name
		} else {
			return macro + " " + name
		}
	} else {
		panic("unimplemented text macro: " + macro)
	}
}

func FormatAtom(atom types.Value) string {
	return FormatMacro("the", atom)
}

func Format(str string, data ...types.Value) string {
	util.FIXME("make this more generic than only accepting atoms")
	parts := strings.Split(str, "[]")
	if len(parts) != len(data)+1 {
		panic(fmt.Sprintf("invalid format string: %d text expressions but %d parameters", len(parts)-1, len(data)))
	}
	moreParts := make([]string, len(parts)+len(data))
	moreParts[0] = parts[0]
	for i := 0; i < len(data); i++ {
		moreParts[2*i+1] = FormatAtom(data[i].(*types.Datum))
		moreParts[2*i+2] = parts[i+1]
	}
	return strings.Join(moreParts, "")
}
