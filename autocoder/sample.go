package main

import (
	"github.com/celskeggs/mediator/autocoder/lib"
	"os"
	"github.com/celskeggs/mediator/autocoder/iface"
	"github.com/celskeggs/mediator/autocoder/gotype"
	"github.com/celskeggs/mediator/autocoder/autodatum"
)

func sampleMain(g iface.Gen, out iface.OutSource) {
	format := g.Import("github.com/celskeggs/mediator/platform/format")
	platform := g.Import("github.com/celskeggs/mediator/platform")

	out.Struct("YourFirstWorld", func(g iface.Gen, out iface.OutStruct) {
		out.Include(platform.Type("BaseTreeDefiner"))
	})

	// MobPlayer
	autodatum.Derive(g, out,"MobPlayer", platform.Type("IMob"),
		func(g iface.Gen, out autodatum.OutDatum) {
			out.Func("Bump", []gotype.Type{platform.Type("IAtom")}, nil,
				func(g iface.Gen, out iface.OutFunc, this iface.Expr, params []iface.Expr) {
					obstacle := params[0]
					out.Invoke(this, "OutputString", format.Invoke("Format", g.String("You bump into []."), obstacle))
					out.Invoke(this, "OutputSound", this.Invoke("World").Invoke(
						"Sound", g.String("ouch.wav"), g.Bool(false), g.Bool(false), g.Int(0), g.Int(100)))
				})
		})

	// CustomArea
	autodatum.DeriveOverride(g, out,"CustomArea", platform.Type("IArea"),
		func(g iface.Gen, out autodatum.OutDatum) {
		})
}

func main() {
	err := lib.Generate("main", sampleMain, os.Stdout)
	if err != nil {
		panic("error: " + err.Error())
	}
}
