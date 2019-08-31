package iface

import (
	"github.com/celskeggs/mediator/autocoder/gotype"
)

// autocode implementations must accept being called multiple times and produce the same result each time!

type Expr interface {
	Ref() Expr
	Field(name string) Expr
	Call(args ...Expr) Expr
	Invoke(name string, args ...Expr) Expr
	Cast(goType gotype.Type) Expr
	Equals(other Expr) Expr
}

type AutocodeFuncOn func(Gen, OutFunc, Expr, []Expr)
type AutocodeFunc func(Gen, OutFunc, []Expr)
type AutocodeBlock func(Gen, OutFunc)
type OutFunc interface {
	Assign(lvalue Expr, rvalue Expr)
	AssignField(target Expr, field string, value Expr)
	Invoke(target Expr, name string, args ...Expr)

	Return(results ...Expr)
	Panic(text string)

	DeclareVar(vartype gotype.Type, initializer Expr) Expr
	DeclareVars(vartypes []gotype.Type, initializers ...Expr) []Expr

	For(condition Expr, body AutocodeBlock)
	If(condition Expr, ifTrue AutocodeBlock, ifFalse AutocodeBlock)
}

type AutocodeStruct func(Gen, OutStruct)
type OutStruct interface {
	Include(goType gotype.Type)
	Field(name string, goType gotype.Type)
}

type AutocodeInterface func(Gen, OutInterface)
type OutInterface interface {
	Include(goType gotype.Type)
	Func(name string, params []gotype.Type, results []gotype.Type)
}

type AutocodeSource func(Gen, OutSource)
type OutSource interface {
	Struct(name string, body AutocodeStruct) gotype.Type
	Interface(name string, body AutocodeInterface) gotype.Type
	Global(name string, goType gotype.Type, initializer Expr)
	Func(name string, params []gotype.Type, results []gotype.Type, body AutocodeFunc)
	FuncOn(goType gotype.Type, name string, params []gotype.Type, results []gotype.Type, body AutocodeFuncOn)
}

type Gen interface {
	Import(s string) Package
	AddError(err error)

	LiteralStruct(goType gotype.Type, initial map[string]Expr) Expr
	String(string) Expr
	Int(int) Expr
	Bool(bool) Expr
	Nil() Expr
}

type Package interface {
	Type(string) gotype.Type
	Field(name string) Expr
	Invoke(name string, args ...Expr) Expr
}
