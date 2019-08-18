package lib

import "github.com/celskeggs/mediator/autocoder/iface"

type writeFunc struct {
	G *generator
}

var _ iface.OutFunc = &writeFunc{}

func (w *writeFunc) AssignField(target iface.Expr, field string, value iface.Expr) {
	w.G.Write("%v.%s = %v\n", target.(expression), field, value.(expression))
}

func (w *writeFunc) Invoke(target iface.Expr, name string, args ...iface.Expr) {
	w.G.Write("%v\n", target.Invoke(name, args...))
}

func (w *writeFunc) Return(results ...iface.Expr) {
	w.G.Write("return %s\n", commaSeparated(results))
}
