package datum

import "strings"

type TypePath string

func (path TypePath) IsValid() bool {
	sp := string(path)
	return strings.Count(sp, "/") < len(sp) &&
		strings.Count(sp, "//") == 0 &&
		sp[0] == '/' &&
		sp[len(sp)-1] != '/'
}

func (path TypePath) Validate() {
	if !path.IsValid() {
		panic("path is not valid: " + path)
	}
}
