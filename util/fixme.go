package util

import (
	"fmt"
	"runtime"
)

var displayedFixmes = map[string]bool{}

// mark things to be done in a way that won't get forgotten
func FIXME(msg string) {
	if !displayedFixmes[msg] {
		displayedFixmes[msg] = true
		_, file, line, ok := runtime.Caller(1)
		if ok {
			println("*** FIXME:", msg, fmt.Sprintf("at %s:%d", file, line))
		} else {
			println("*** FIXME:", msg)
		}
	}
}

// based on https://stackoverflow.com/a/3090386/3369324

func NiceToHave(msg string) {
	// don't even bother saying anything at this time
}
