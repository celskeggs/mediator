package lib

import (
	"github.com/celskeggs/mediator/autocoder/iface"
	"github.com/celskeggs/mediator/autocoder/gotype"
	"github.com/hashicorp/go-multierror"
)

type importCollector struct {
	Imports map[string]bool
	Error   error
}

type importCollectorPackage struct {
	Package         string
	ImportCollector *importCollector
}

type importCollectorExpr struct {}

func (ic *importCollector) ImportList() (out []string) {
	for imp := range ic.Imports {
		out = append(out, imp)
	}
	return out
}

var _ iface.Gen = &importCollector{}
var _ iface.OutSource = &importCollector{}
var _ iface.OutInterface = &importCollector{}
var _ iface.OutStruct = &importCollector{}
var _ iface.OutFunc = &importCollector{}

var _ iface.Package = importCollectorPackage{}
var _ iface.Expr = importCollectorExpr{}

func collectImports(source iface.AutocodeSource) ([]string, error) {
	ic := &importCollector{
		Imports: map[string]bool{},
	}
	source(ic, ic)
	if ic.Error != nil {
		return nil, ic.Error
	}
	return ic.ImportList(), nil
}

func (ic *importCollector) Import(pkg string) iface.Package {
	ic.Imports[pkg] = true
	return importCollectorPackage{
		Package:         pkg,
		ImportCollector: ic,
	}
}

func (ic *importCollector) AddError(err error) {
	if err != nil {
		ic.Error = multierror.Append(ic.Error, err)
	}
}

func (ic *importCollector) LiteralStruct(goType gotype.Type, initial map[string]iface.Expr) iface.Expr {
	return importCollectorExpr{}
}

func (ic *importCollector) String(string) iface.Expr {
	return importCollectorExpr{}
}

func (ic *importCollector) Int(int) iface.Expr {
	return importCollectorExpr{}
}

func (ic *importCollector) Bool(bool) iface.Expr {
	return importCollectorExpr{}
}

func (ic *importCollector) Struct(name string, body iface.AutocodeStruct) gotype.Type {
	body(ic, ic)
	return gotype.LocalType(name)
}

func (ic *importCollector) Interface(name string, body iface.AutocodeInterface) gotype.Type {
	body(ic, ic)
	return gotype.LocalType(name)
}

func (ic *importCollector) Global(name string, goType gotype.Type, initializer iface.Expr) {
	// nothing to do
}

func (ic *importCollector) FuncOn(goType gotype.Type, name string, params []gotype.Type, results []gotype.Type, body iface.AutocodeFuncOn) {
	paramExprs := make([]iface.Expr, len(params))
	for i := range params {
		paramExprs[i] = importCollectorExpr{}
	}
	body(ic, ic, importCollectorExpr{}, paramExprs)
}

func (ic *importCollector) Include(goType gotype.Type) {
	// nothing to do
}

func (ic *importCollector) Func(name string, params []gotype.Type, results []gotype.Type) {
	// nothing to do
}

func (ic *importCollector) Field(name string, goType gotype.Type) {
	// nothing to do
}

func (ic *importCollector) AssignField(target iface.Expr, field string, value iface.Expr) {
	// nothing to do
}

func (ic *importCollector) Invoke(target iface.Expr, name string, args ...iface.Expr) {
	// nothing to do
}

func (ic *importCollector) Return(results ...iface.Expr) {
	// nothing to do
}

func (ice importCollectorExpr) Ref() iface.Expr {
	return ice
}

func (ice importCollectorExpr) Field(name string) iface.Expr {
	return ice
}

func (ice importCollectorExpr) Call(args ...iface.Expr) iface.Expr {
	return ice
}

func (ice importCollectorExpr) Invoke(name string, args ...iface.Expr) iface.Expr {
	return ice
}

func (ice importCollectorExpr) Cast(goType gotype.Type) iface.Expr {
	return ice
}

func (icp importCollectorPackage) Type(pkg string) gotype.Type {
	return gotype.PackageType(icp.Package, pkg)
}

func (icp importCollectorPackage) Field(name string) iface.Expr {
	return importCollectorExpr{}
}

func (icp importCollectorPackage) Invoke(name string, args ...iface.Expr) iface.Expr {
	return importCollectorExpr{}
}
