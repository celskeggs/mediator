package lib

import (
	"github.com/celskeggs/mediator/autocoder/iface"
	"github.com/celskeggs/mediator/autocoder/gotype"
)

type writeStructInterface struct {
	G *generator
}

var _ iface.OutInterface = &writeStructInterface{}
var _ iface.OutStruct = &writeStructInterface{}

func (w *writeStructInterface) Include(goType gotype.Type) {
	w.G.Write("%v\n", goType)
}

func (w *writeStructInterface) Field(name string, goType gotype.Type) {
	validateIdentifier(name)
	w.G.Write("%s %v\n", name, goType)
}

func (w *writeStructInterface) Func(name string, params []gotype.Type, results []gotype.Type) {
	validateIdentifier(name)
	w.G.Write("%s(%s) (%s)\n", name, stringTypes(params), stringTypes(results))
}
