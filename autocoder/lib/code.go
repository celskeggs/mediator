package lib

import (
	"github.com/celskeggs/mediator/autocoder/iface"
	"github.com/celskeggs/mediator/autocoder/gotype"
)

type writeFunc struct {
	G   *generator
	VNC *varNameCtx
}

var _ iface.OutFunc = &writeFunc{}

func (w *writeFunc) Assign(lvalue iface.Expr, rvalue iface.Expr) {
	w.G.Write("%s = %v\n", lvalue.(expression), rvalue.(expression))
}

func (w *writeFunc) AssignField(target iface.Expr, field string, value iface.Expr) {
	w.G.Write("%v.%s = %v\n", target.(expression), field, value.(expression))
}

func (w *writeFunc) Invoke(target iface.Expr, name string, args ...iface.Expr) {
	w.G.Write("%v\n", target.Invoke(name, args...))
}

func (w *writeFunc) Return(results ...iface.Expr) {
	w.G.Write("return %s\n", commaSeparated(results))
}

func (w *writeFunc) Panic(text string) {
	w.G.Write("panic(%s)\n", escapeString(text))
}

func (w *writeFunc) DeclareVar(vartype gotype.Type, initializer iface.Expr) iface.Expr {
	name := w.VNC.VarFromType(vartype)
	if initializer == nil {
		w.G.Write("var %v %v\n", name, vartype)
	} else {
		w.G.Write("var %v %v = %v\n", name, vartype, initializer.(expression))
	}
	return name
}

func (w *writeFunc) DeclareVars(vartypes []gotype.Type, initializers ...iface.Expr) []iface.Expr {
	varStr, names := w.VNC.VarsFromTypes(vartypes)
	if len(initializers) == 0 {
		w.G.Write("var %s\n", varStr)
	} else {
		w.G.Write("var %s = %s\n", varStr, commaSeparated(initializers))
	}
	return names
}

func (w *writeFunc) For(condition iface.Expr, body iface.AutocodeBlock) {
	if condition == nil {
		w.G.Write("for {\n")
	} else {
		w.G.Write("for %v {\n", condition)
	}
	w.G.Indent()
	body(w.G, w)
	w.G.Unindent()
	w.G.Write("}\n")
}

func (w *writeFunc) If(condition iface.Expr, ifTrue iface.AutocodeBlock, ifFalse iface.AutocodeBlock) {
	w.G.Write("if %v {\n", condition)
	w.G.Indent()
	ifTrue(w.G, w)
	w.G.Unindent()
	if ifFalse != nil {
		w.G.Write("} else {\n")
		ifFalse(w.G, w)
	}
	w.G.Write("}\n")
}
