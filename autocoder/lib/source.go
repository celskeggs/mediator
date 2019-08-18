package lib

import (
	"github.com/celskeggs/mediator/autocoder/iface"
	"github.com/celskeggs/mediator/autocoder/gotype"
	"strings"
	"fmt"
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

type varNameCtx struct {
	Next map[rune]int
}

func newVarNameCtx() *varNameCtx {
	return &varNameCtx{
		Next: map[rune]int{},
	}
}

func (v *varNameCtx) Name(base string) string {
	if base == "" {
		base = "var"
	} else {
		base = strings.ToLower(base)
	}
	r := []rune(base)[0]
	cur := v.Next[r]
	v.Next[r] += 1
	if cur == 0 {
		return base[0:1]
	} else {
		return fmt.Sprintf("%s%d", base[0:1], cur)
	}
}

func (v *varNameCtx) Var(base string) iface.Expr {
	return formatExpression("%s", v.Name(base))
}

func (v *varNameCtx) VarFromType(base gotype.Type) iface.Expr {
	return v.Var(base.Name())
}

func (v *varNameCtx) VarsFromTypes(basis []gotype.Type) (paramStr string, exprs []iface.Expr) {
	var paramStrs []string
	for _, base := range basis {
		expr := v.VarFromType(base)
		paramStrs = append(paramStrs, fmt.Sprintf("%v %v", expr, base))
		exprs = append(exprs, expr)
	}
	return strings.Join(paramStrs, ", "), exprs
}

func (w *writeSource) FuncOn(structType gotype.Type, name string, params []gotype.Type, results []gotype.Type, body iface.AutocodeFuncOn) {
	validateIdentifier(name)
	vnc := newVarNameCtx()
	this := vnc.VarFromType(structType)
	paramStr, paramExprs := vnc.VarsFromTypes(params)
	w.G.Write("func (%v %v) %s(%s) (%s) {\n",
		this, structType, name, paramStr, stringTypes(results))
	w.G.Indent()
	body(w.G, &writeFunc{G: w.G}, this, paramExprs)
	w.G.Unindent()
	w.G.Write("}\n\n")
}
