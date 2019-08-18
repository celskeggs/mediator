package lib

import (
	"github.com/celskeggs/mediator/autocoder/gotype"
	"github.com/celskeggs/mediator/autocoder/iface"
	"fmt"
	"strings"
	"github.com/celskeggs/mediator/autocoder/indent"
)

type expression struct {
	Display string
}

var _ iface.Expr = expression{}

func (e expression) String() string {
	return e.Display
}

func formatExpression(format string, args ...interface{}) iface.Expr {
	return expression{
		Display: fmt.Sprintf(format, args...),
	}
}

func (g *generator) LiteralStruct(goType gotype.Type, initial map[string]iface.Expr) iface.Expr {
	var prefix string
	if goType.IsPtr() {
		prefix = "&"
		goType = goType.UnwrapPtr()
	}

	var fields []string
	for k, v := range initial {
		fields = append(fields, fmt.Sprintf("%s: %v,", k, v.(expression)))
	}

	return formatExpression("%s%v {\n%s}", prefix, goType, indent.Indent(strings.Join(fields, "")))
}

func (g *generator) String(s string) iface.Expr {
	return formatExpression("%s", escapeString(s))
}

func (g *generator) Int(i int) iface.Expr {
	return formatExpression("%d", i)
}

func (g *generator) Bool(b bool) iface.Expr {
	return formatExpression("%v", b)
}

func (e expression) Ref() iface.Expr {
	return formatExpression("&%v", e)
}

func (e expression) Field(name string) iface.Expr {
	validateIdentifier(name)
	return formatExpression("%v.%s", e, name)
}

func (e expression) Call(args ...iface.Expr) iface.Expr {
	return formatExpression("%v(%s)", e, commaSeparated(args))
}

func (e expression) Invoke(name string, args ...iface.Expr) iface.Expr {
	return e.Field(name).Call(args...)
}

func (e expression) Cast(goType gotype.Type) iface.Expr {
	return formatExpression("%v.(%v)", e, goType)
}

func (p packageRef) Field(name string) iface.Expr {
	return formatExpression("%s.%s", p.Package, name)
}

func (p packageRef) Invoke(name string, params ...iface.Expr) iface.Expr {
	return p.Field(name).Call(params...)
}
