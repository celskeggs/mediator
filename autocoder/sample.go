package main

import (
	"github.com/celskeggs/mediator/autocoder/lib"
	"os"
	"github.com/celskeggs/mediator/autocoder/iface"
	"github.com/celskeggs/mediator/autocoder/gotype"
)

func sampleMain(g iface.Gen, out iface.OutSource) {
	datum := g.Import("github.com/celskeggs/mediator/platform/datum")
	format := g.Import("github.com/celskeggs/mediator/platform/format")
	platform := g.Import("github.com/celskeggs/mediator/platform")

	out.Struct("YourFirstWorld", func(g iface.Gen, out iface.OutStruct) {
		out.Include(platform.Type("BaseTreeDefiner"))
	})

	// MobPlayer
	IMobPlayer := out.Interface("IMobPlayer",
		func(g iface.Gen, out iface.OutInterface) {
			out.Include(platform.Type("IMob"))
		})
	MobPlayer := out.Struct("MobPlayer",
		func(g iface.Gen, out iface.OutStruct) {
			out.Include(platform.Type("IMob"))
		})
	out.Global("_", IMobPlayer, g.LiteralStruct(MobPlayer, nil).Ref())
	out.FuncOn(MobPlayer, "RawClone", nil, []gotype.Type{datum.Type("IDatum")},
		func(g iface.Gen, out iface.OutFunc, this iface.Expr, params []iface.Expr) {
			out.AssignField(this, "IMob", this.Field("IMob").Invoke("RawClone").Cast(platform.Type("IMob")))
			out.Return(this.Ref())
		})
	out.FuncOn(MobPlayer.Ptr(), "Bump", []gotype.Type{platform.Type("IAtom")}, nil,
		func(g iface.Gen, out iface.OutFunc, this iface.Expr, params []iface.Expr) {
			obstacle := params[0]
			out.Invoke(this, "OutputString", format.Invoke("Format", g.String("You bump into []."), obstacle))
			out.Invoke(this, "OutputSound", this.Invoke("World").Invoke(
				"Sound", g.String("ouch.wav"), g.Bool(false), g.Bool(false), g.Int(0), g.Int(100)))
		})

	// CustomArea
	CustomArea := out.Struct("CustomArea",
		func(g iface.Gen, out iface.OutStruct) {
			out.Include(platform.Type("IArea"))
			out.Field("Music", gotype.TypeString())
		})
	ICustomArea := out.Interface("ICustomArea",
		func(g iface.Gen, out iface.OutInterface) {
			out.Include(platform.Type("IArea"))
			out.Func("AsCustomArea", nil, []gotype.Type{CustomArea.Ptr()})
		})
	out.Global("_", ICustomArea, g.LiteralStruct(CustomArea, nil).Ref())
	out.FuncOn(CustomArea, "RawClone", nil, []gotype.Type{datum.Type("IDatum")},
		func(g iface.Gen, out iface.OutFunc, this iface.Expr, params []iface.Expr) {
			out.AssignField(this, "IArea", this.Field("IArea").Invoke("RawClone").Cast(platform.Type("IArea")))
			out.Return(this.Ref())
		})
}

func main() {
	err := lib.Generate("main", sampleMain, os.Stdout)
	if err != nil {
		panic("error: " + err.Error())
	}
}
