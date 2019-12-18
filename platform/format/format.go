package format

import (
	"github.com/celskeggs/mediator/platform/types"
	"github.com/celskeggs/mediator/util"
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
