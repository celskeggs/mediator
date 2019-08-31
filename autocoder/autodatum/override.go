package autodatum

import (
	"github.com/celskeggs/mediator/autocoder/iface"
	"github.com/celskeggs/mediator/autocoder/gotype"
	"fmt"
)

func DeriveOverride(g iface.Gen, out iface.OutSource, name string, baseInterface gotype.Type, meta AutocodeDatum) {
	datum := g.Import(DatumImportPath)
	typeIDatum := datum.Type("IDatum")
	declInterface := Derive(g, out, name, baseInterface,
		func(g iface.Gen, out OutDatum) {
			out.Func("NextOverride", nil, []gotype.Type{typeIDatum, typeIDatum},
				func(g iface.Gen, out iface.OutFunc, this iface.Expr, params []iface.Expr) {
					out.Return(this, this.Field(baseInterface.Name()))
				})
			meta(g, out)
		})
	// func Extract<name>(base IBase) I<name> {
	out.Func("Extract" + name, []gotype.Type{baseInterface}, []gotype.Type{declInterface},
		func(g iface.Gen, out iface.OutFunc, params []iface.Expr) {
			baseImplementation := params[0]

			// datum.AssertConsistent(base)
			datum.Invoke("AssertConsistent", baseImplementation)

			// var iter datum.IDatum = base
			iterator := out.DeclareVar(typeIDatum, baseImplementation)
			// for {
			out.For(nil, func(g iface.Gen, out iface.OutFunc) {
				// var cur, next = iter.NextOverride()
				vars := out.DeclareVars([]gotype.Type{typeIDatum, typeIDatum}, iterator.Invoke("NextOverride"))
				cur, next := vars[0], vars[1]

				// var obj, found = cur.(I<name>)
				vars = out.DeclareVars([]gotype.Type{declInterface, gotype.TypeBool()}, cur.Cast(declInterface))
				obj, found := vars[0], vars[1]

				// if found {
				out.If(found,
					func(g iface.Gen, out iface.OutFunc) {
						// return obj
						out.Return(obj)
					}, nil)
				// }
				// if iter == nil {
				out.If(iterator.Equals(g.Nil()),
					func(g iface.Gen, out iface.OutFunc) {
						// panic("instance of IBase does not implement <name>
						out.Panic(fmt.Sprintf("instance of %v does not implement %s", baseInterface, name))
					}, nil)
				// }

				// iter = next
				out.Assign(iterator, next)
			})
			// }
		})
}
