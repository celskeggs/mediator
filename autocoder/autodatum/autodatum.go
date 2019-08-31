package autodatum

import (
	"github.com/celskeggs/mediator/autocoder/gotype"
	"github.com/celskeggs/mediator/autocoder/iface"
	"unicode"
	"github.com/celskeggs/mediator/util"
	"fmt"
)

const PlatformImportPath = "github.com/celskeggs/mediator/platform"
const DatumImportPath = "github.com/celskeggs/mediator/platform/datum"

type AutocodeDatum func(iface.Gen, OutDatum)
type OutDatum interface {
	Field(name string, goType gotype.Type)
	FuncDecl(name string, params []gotype.Type, results []gotype.Type)
	Func(name string, params []gotype.Type, results[] gotype.Type, body iface.AutocodeFuncOn)
}

type adaptOutDatum struct {
	fieldImpl    func(name string, goType gotype.Type)
	funcDeclImpl func(name string, params []gotype.Type, results []gotype.Type)
	funcImpl     func(name string, params []gotype.Type, results []gotype.Type, body iface.AutocodeFuncOn)
}

func (a *adaptOutDatum) Field(name string, goType gotype.Type) {
	if a.fieldImpl != nil {
		a.fieldImpl(name, goType)
	}
}

func (a *adaptOutDatum) FuncDecl(name string, params []gotype.Type, results []gotype.Type) {
	if a.funcDeclImpl != nil {
		a.funcDeclImpl(name, params, results)
	}
}

func (a *adaptOutDatum) Func(name string, params []gotype.Type, results []gotype.Type, body iface.AutocodeFuncOn) {
	if a.funcImpl != nil {
		a.funcImpl(name, params, results, body)
	}
}

func isInterfaceName(name string) bool {
	util.FIXME("do a better job of validating whether this is an interface")
	rname := []rune(name)
	return len(name) >= 2 && rname[0] == 'I' && unicode.IsUpper(rname[1])
}

func isStructName(name string) bool {
	rname := []rune(name)
	return len(name) >= 1 && unicode.IsUpper(rname[0]) && !isInterfaceName(name)
}

func Derive(g iface.Gen, out iface.OutSource, name string, baseInterface gotype.Type, meta AutocodeDatum) (declInterface gotype.Type) {
	if !isStructName(name) {
		panic("invalid name in Derive")
	}
	typeIDatum := g.Import(DatumImportPath).Type("IDatum")
	if !isInterfaceName(baseInterface.Name()) {
		g.AddError(fmt.Errorf("not a valid interface: %v", baseInterface))
	}
	declStruct := gotype.LocalType(name)
	// public interface for type
	declInterface = out.Interface("I" + name, func(g iface.Gen, out iface.OutInterface) {
		out.Include(baseInterface)
		out.Func("As" + name, nil, []gotype.Type{declStruct.Ptr()})
		meta(g, &adaptOutDatum{
			funcDeclImpl: out.Func,
		})
	})
	// core struct for type
	out.Struct(name, func(g iface.Gen, out iface.OutStruct) {
		out.Include(baseInterface)
		meta(g, &adaptOutDatum{
			fieldImpl: out.Field,
		})
	})
	// implementation validator
	out.Global("_", declInterface, g.LiteralStruct(declStruct, nil).Ref())
	// RawClone() implementation
	out.FuncOn(declStruct, "RawClone", nil, []gotype.Type{typeIDatum},
		func(g iface.Gen, out iface.OutFunc, this iface.Expr, params []iface.Expr) {
			// d.IDatum = d.IDatum.RawClone().(datum.IDatum)
			out.AssignField(this, baseInterface.Name(), this.Field(baseInterface.Name()).Invoke("RawClone").Cast(baseInterface))
			// return &d
			out.Return(this.Ref())
		})
	// As<Type>() implementation
	out.FuncOn(declStruct.Ptr(), "As" + name, nil, []gotype.Type{declStruct.Ptr()},
		func(g iface.Gen, out iface.OutFunc, this iface.Expr, params []iface.Expr) {
			out.Return(this)
		})
	// additional functions
	meta(g, &adaptOutDatum{
		funcImpl: func(name string, params []gotype.Type, results []gotype.Type, body iface.AutocodeFuncOn) {
			out.FuncOn(declStruct.Ptr(), name, params, results, body)
		},
	})
	return declInterface
}
