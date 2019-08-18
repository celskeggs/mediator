package lib

import (
	"github.com/celskeggs/mediator/autocoder/gotype"
	"github.com/celskeggs/mediator/autocoder/iface"
)

type packageRef struct {
	Path    string
	Package string
}

var _ iface.Package = packageRef{}

func (p packageRef) Type(typeName string) gotype.Type {
	return gotype.PackageType(p.Package, typeName)
}

// additional functions defined in expr.go
