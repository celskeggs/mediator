package main

import (
	"github.com/celskeggs/mediator/autocoder/lib"
	"os"
)

func main() {
	gen := lib.StartGeneration("main")
	datum := gen.Import("github.com/celskeggs/mediator/platform/datum")
	format := gen.Import("github.com/celskeggs/mediator/platform/format")
	platform := gen.Import("github.com/celskeggs/mediator/platform")
	gen.Struct("YourFirstWorld",
		gen.ElemInclude(platform.GetStruct("BaseTreeDefiner")))

	// MobPlayer
	IMobPlayer := gen.Interface("IMobPlayer",
		gen.ElemInclude(platform.GetInterface("IMob")))
	MobPlayer := gen.Struct("MobPlayer",
		gen.ElemInclude(platform.GetStruct("Mob")))
	gen.Global("_", IMobPlayer,
		gen.LiteralStructPtr(MobPlayer))
	gen.FuncOn(MobPlayer, "RawClone", nil, []lib.Type{datum.GetInterface("IDatum")},
		gen.Assign(gen.Self().Field("IMob"), gen.Self().Field("IMob").Invoke("RawClone").Cast(platform.GetStruct("IMob"))),
		gen.Return(gen.SelfRef()))
	gen.FuncOn(MobPlayer.Ptr(), "Bump", []lib.Type{platform.GetInterface("IAtom")}, nil,
		gen.Self().Invoke("OutputString", format.Invoke("Format", gen.LiteralString("You bump into []."), gen.Param(0))).Statement(),
		gen.Self().Invoke("OutputSound", gen.Self().Invoke("World").Invoke(
			"Sound", gen.LiteralString("ouch.wav"), gen.LiteralBool(false), gen.LiteralBool(false), gen.LiteralInt(0), gen.LiteralInt(100))).Statement())

	// CustomArea
	CustomArea := gen.Struct("CustomArea",
		gen.ElemInclude(platform.GetInterface("IArea")),
		gen.ElemField("Music", gen.TypeString()))
	ICustomArea := gen.Interface("ICustomArea",
		gen.ElemInclude(platform.GetInterface("IArea")),
		gen.ElemFunc("AsCustomArea", nil, []lib.Type{CustomArea.Ptr()}))
	gen.Global("_", ICustomArea,
		gen.LiteralStructPtr(CustomArea))
	gen.FuncOn(CustomArea, "RawClone", nil, []lib.Type{datum.GetInterface("IDatum")},
		gen.Assign(gen.Self().Field("IArea"), gen.Self().Field("IArea").Invoke("RawClone").Cast(platform.GetStruct("IArea"))),
		gen.Return(gen.SelfRef()))

	err := gen.WriteTo(os.Stdout)
	if err != nil {
		panic("error: " + err.Error())
	}
}
