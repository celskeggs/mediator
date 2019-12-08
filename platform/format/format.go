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

func formatMacro(macro string, atom *types.Datum) string {
	if macro == "the" || macro == "The" {
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

func FormatAtom(atom types.Value) types.String {
	return types.String(formatMacro("the", atom.(*types.Datum)))
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
		moreParts[2*i+1] = string(FormatAtom(data[i].(*types.Datum)))
		moreParts[2*i+2] = parts[i+1]
	}
	return strings.Join(moreParts, "")
}
