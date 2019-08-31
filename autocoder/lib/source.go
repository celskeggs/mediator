package lib

import (
	"github.com/celskeggs/mediator/autocoder/iface"
	"github.com/celskeggs/mediator/autocoder/gotype"
)

type writeSource struct {
	G *generator
}

var _ iface.OutSource = &writeSource{}

func (w *writeSource) Struct(name string, body iface.AutocodeStruct) gotype.Type {
	validateIdentifier(name)
	w.G.Write("type %s struct {\n", name)
	w.G.Indent()
	body(w.G, &writeStructInterface{G: w.G})
	w.G.Unindent()
	w.G.Write("}\n\n")
	return gotype.LocalType(name)
}

func (w *writeSource) Interface(name string, body iface.AutocodeInterface) gotype.Type {
	validateIdentifier(name)
	w.G.Write("type %s interface {\n", name)
	w.G.Indent()
	body(w.G, &writeStructInterface{G: w.G})
	w.G.Unindent()
	w.G.Write("}\n\n")
	return gotype.LocalType(name)
}

func (w *writeSource) Global(name string, goType gotype.Type, value iface.Expr) {
	validateIdentifier(name)
	w.G.Write("var %s %v = %v\n\n", name, goType, value.(expression))
}

func (w *writeSource) Func(name string, params []gotype.Type, results []gotype.Type, body iface.AutocodeFunc) {
	validateIdentifier(name)
	vnc := newVarNameCtx()
	paramStr, paramExprs := vnc.VarsFromTypes(params)
	w.G.Write("func %s(%s) (%s) {\n",
		name, paramStr, stringTypes(results))
	w.G.Indent()
	body(w.G, &writeFunc{G: w.G, VNC: vnc}, paramExprs)
	w.G.Unindent()
	w.G.Write("}\n\n")
}

func (w *writeSource) FuncOn(structType gotype.Type, name string, params []gotype.Type, results []gotype.Type, body iface.AutocodeFuncOn) {
	validateIdentifier(name)
	vnc := newVarNameCtx()
	this := vnc.VarFromType(structType)
	paramStr, paramExprs := vnc.VarsFromTypes(params)
	w.G.Write("func (%v %v) %s(%s) (%s) {\n",
		this, structType, name, paramStr, stringTypes(results))
	w.G.Indent()
	body(w.G, &writeFunc{G: w.G, VNC: vnc}, this, paramExprs)
	w.G.Unindent()
	w.G.Write("}\n\n")
}
