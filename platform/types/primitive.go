package types

import (
	"fmt"
	"strings"
)

type String string

var _ Value = String("")

func Unstring(b Value) string {
	return string(b.(String))
}

func (s String) Var(name string) Value {
	panic("no variable " + name + " on string")
}

func (s String) SetVar(name string, value Value) {
	panic("no variable " + name + " on string")
}

func (s String) Invoke(usr *Datum, name string, parameters ...Value) Value {
	panic("no proc " + name + " on string")
}

func (s String) String() string {
	return fmt.Sprintf("[string: %q]", string(s))
}

type Int int

var _ Value = Int(0)

func Unint(i Value) int {
	return int(i.(Int))
}

func Unuint(i Value) uint {
	iv := Unint(i)
	if iv < 0 {
		panic("attempt to types.Unuint negative number!")
	}
	return uint(iv)
}

func (i Int) Var(name string) Value {
	panic("no variable " + name + " on int")
}

func (i Int) SetVar(name string, value Value) {
	panic("no variable " + name + " on int")
}

func (i Int) Invoke(usr *Datum, name string, parameters ...Value) Value {
	panic("no proc " + name + " on int")
}

func (i Int) String() string {
	return fmt.Sprintf("[int: %d]", int(i))
}

type TypePath string

var _ Value = TypePath("")

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

func (path TypePath) String() string {
	return string(path)
}

func (path TypePath) Var(name string) Value {
	panic("cannot get variable on type path")
}

func (path TypePath) SetVar(name string, value Value) {
	panic("cannot set variable on type path")
}

func (path TypePath) Invoke(usr *Datum, name string, parameters ...Value) Value {
	panic("cannot invoke method on type path")
}

func AsBool(v Value) bool {
	if v == nil {
		return false
	} else if i, ok := v.(Int); ok {
		return int(i) != 0
	} else if s, ok := v.(String); ok {
		return string(s) != ""
	} else {
		return true
	}
}

func FromBool(b bool) Int {
	if b {
		return 1
	} else {
		return 0
	}
}
