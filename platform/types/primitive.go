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

func (s String) Reference() *Ref {
	return &Ref{s}
}

func (s String) Var(name string) Value {
	panic("no variable " + name + " on string")
}

func (s String) SetVar(name string, value Value) {
	panic("no variable " + name + " on string")
}

func (s String) Invoke(name string, parameters ...Value) Value {
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

func (i Int) Reference() *Ref {
	return &Ref{i}
}

func (i Int) Var(name string) Value {
	panic("no variable " + name + " on int")
}

func (i Int) SetVar(name string, value Value) {
	panic("no variable " + name + " on int")
}

func (i Int) Invoke(name string, parameters ...Value) Value {
	panic("no proc " + name + " on int")
}

func (i Int) String() string {
	return fmt.Sprintf("[int: %d]", int(i))
}

type Bool bool

var _ Value = Bool(false)

func Unbool(b Value) bool {
	return bool(b.(Bool))
}

func (b Bool) Reference() *Ref {
	return &Ref{b}
}

func (b Bool) Var(name string) Value {
	panic("no variable " + name + " on bool")
}

func (b Bool) SetVar(name string, value Value) {
	panic("no variable " + name + " on bool")
}

func (b Bool) Invoke(name string, parameters ...Value) Value {
	panic("no proc " + name + " on bool")
}

func (b Bool) String() string {
	if b {
		return "[true]"
	} else {
		return "[false]"
	}
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

func (path TypePath) Reference() *Ref {
	return &Ref{v: path}
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

func (path TypePath) Invoke(name string, parameters ...Value) Value {
	panic("cannot invoke method on type path")
}
