package main

import (
	"github.com/celskeggs/mediator/autocoder/lib"
	"os"
)

func main() {
	gen := lib.StartGeneration("main")
	datum := gen.Import("github.com/celskeggs/mediator/platform/datum")
	platform := gen.Import("github.com/celskeggs/mediator/platform")
	gen.Struct("YourFirstWorld",
		gen.Include(platform.GetStruct("BaseTreeDefiner")))
	IMobPlayer := gen.Interface("IMobPlayer",
		gen.Include(platform.GetInterface("IMob")))
	MobPlayer := gen.Struct("MobPlayer",
		gen.Include(platform.GetStruct("Mob")))
	gen.Global("_", IMobPlayer,
		gen.LiteralStructPtr(MobPlayer))
	gen.FuncOn(MobPlayer, "RawClone", nil, []lib.Type{datum.GetInterface("IDatum")},
		gen.Assign(gen.Self().Field("IMob"), gen.Self().Field("IMob").Invoke("RawClone").Cast(platform.GetStruct("IMob"))),
		gen.Return(gen.SelfRef()))
	err := gen.WriteTo(os.Stdout)
	if err != nil {
		panic("error: " + err.Error())
	}
}
