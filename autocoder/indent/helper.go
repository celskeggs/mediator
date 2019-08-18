package indent

import "strings"

func Indent(block string) string {
	if block == "" {
		return ""
	}
	return "\t" + strings.Replace(block, "\n", "\n\t", -1)
}
