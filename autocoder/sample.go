package main

import (
	"github.com/celskeggs/mediator/autocoder/replica"
	"os/exec"
	"os"
)

func main() {
	var outName = "main_gen.go"

	err := replica.GenerateTo(&replica.DefinedTree{
		Types: []replica.DefinedType{
			{
				TypePath: "/mob/player",
				Funcs: []replica.DefinedFunc{
					{
						Name: "Bump",
						Params: []replica.DefinedParam{
							{Name: "obstacle", Type: "platform.IAtom"},
						},
						Body: `
this.OutputString(format.Format("You bump into [].", obstacle))
this.OutputSound(this.World().Sound("ouch.wav", false, false, 0, 100))
`,
					},
				},
				Inits: []replica.DefinedInit{
					{
						ShortName: "name",
						Value: "\"player\"",
					},
					{
						ShortName: "icon",
						Value: "icons.LoadOrPanic(\"player.dmi\")",
					},
				},
			},
			{
				TypePath: "/mob/rat",
				Inits: []replica.DefinedInit{
					{
						ShortName: "name",
						Value: "\"rat\"",
					},
					{
						ShortName: "icon",
						Value: "icons.LoadOrPanic(\"rat.dmi\")",
					},
				},
			},
			{
				TypePath: "/turf/floor",
				Inits: []replica.DefinedInit{
					{
						ShortName: "name",
						Value: "\"floor\"",
					},
					{
						ShortName: "icon",
						Value: "icons.LoadOrPanic(\"floor.dmi\")",
					},
				},
			},
			{
				TypePath: "/turf/wall",
				Inits: []replica.DefinedInit{
					{
						ShortName: "name",
						Value: "\"wall\"",
					},
					{
						ShortName: "icon",
						Value: "icons.LoadOrPanic(\"wall.dmi\")",
					},
					{
						ShortName: "density",
						Value: "true",
					},
					{
						ShortName: "opacity",
						Value: "true",
					},
				},
			},
			{
				TypePath: "/obj/cheese",
				Inits: []replica.DefinedInit{
					{
						ShortName: "name",
						Value: "\"cheese\"",
					},
					{
						ShortName: "icon",
						Value: "icons.LoadOrPanic(\"cheese.dmi\")",
					},
				},
			},
			{
				TypePath: "/obj/scroll",
				Inits: []replica.DefinedInit{
					{
						ShortName: "name",
						Value: "\"scroll\"",
					},
					{
						ShortName: "icon",
						Value: "icons.LoadOrPanic(\"scroll.dmi\")",
					},
				},
			},
			{
				TypePath: "/area/outside",
				Inits: []replica.DefinedInit{
					{
						ShortName: "name",
						Value: "\"outside\"",
					},
					{
						ShortName: "desc",
						Value: "\"Nice and jazzy, here...\"",
					},
					{
						ShortName: "music",
						Value: "\"jazzy.ogg\"",
					},
				},
			},
			{
				TypePath: "/area/cave",
				Inits: []replica.DefinedInit{
					{
						ShortName: "name",
						Value: "\"cave\"",
					},
					{
						ShortName: "desc",
						Value: "\"Watch out for the giant rat!\"",
					},
					{
						ShortName: "music",
						Value: "\"cavern.ogg\"",
					},
				},
			},
			{
				BasePath: "/area",
				TypePath: "/area",
				Fields: []replica.DefinedField{
					{
						Name: "music",
						Type: "string",
						Default: "\"\"",
					},
				},
				Funcs: []replica.DefinedFunc{
					{
						Name: "Entered",
						Params: []replica.DefinedParam{
							{Name: "atom", Type: "platform.IAtomMovable"},
							{Name: "oldloc", Type: "platform.IAtom"},
						},
						Body: `
if mob, ismob := atom.(platform.IMob); ismob {
	mob.OutputString(this.AsAtom().Appearance.Desc)
	mob.OutputSound(this.World().Sound(this.Music, true, false, 1, 100))
}`,
					},
				},
			},
		},
		WorldMap:                "map.dmm",
		DefaultCoreResourcesDir: "../resources",
		DefaultIconsDir:         "resources",
		WorldName:               "Your First World",
		WorldMob:                "/mob/player",
	}, outName)
	if err != nil {
		panic("generate error: " + err.Error())
	}

	cmd := exec.Command("diff", "-u", "main.go", outName)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Run()
}
