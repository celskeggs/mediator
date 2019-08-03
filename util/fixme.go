package util

var displayedFixmes = map[string]bool{}

// mark things to be done in a way that won't get forgotten
func FIXME(msg string) {
	if !displayedFixmes[msg] {
		displayedFixmes[msg] = true
		println("*** FIXME:", msg)
	}
}

// based on https://stackoverflow.com/a/3090386/3369324
