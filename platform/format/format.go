package format

import (
	"strings"
	"fmt"
	"github.com/celskeggs/mediator/util"
	"github.com/celskeggs/mediator/platform"
	"unicode"
)

func isUpperCase(s string) bool {
	for _, r := range s {
		return unicode.IsUpper(r)
	}
	return false
}

func formatMacro(macro string, atom platform.IAtom) string {
	if macro == "the" || macro == "The" {
		name := atom.AsAtom().Appearance.Name
		if isUpperCase(name) {
			return name
		} else {
			return macro + " " + name
		}
	} else {
		panic("unimplemented text macro: " + macro)
	}
}

func formatAtom(atom platform.IAtom) string {
	return formatMacro("the", atom)
}

func Format(str string, atoms ...platform.IAtom) string {
	util.FIXME("make this more generic than only accepting atoms")
	parts := strings.Split(str, "[]")
	if len(parts) != len(atoms)+1 {
		panic(fmt.Sprintf("invalid format string: %d text expressions but %d parameters", len(parts)-1, len(atoms)))
	}
	moreParts := make([]string, len(parts)+len(atoms))
	moreParts[0] = parts[0]
	for i := 0; i < len(atoms); i++ {
		moreParts[2*i+1] = formatAtom(atoms[i])
		moreParts[2*i+2] = parts[i+1]
	}
	return strings.Join(moreParts, "")
}
