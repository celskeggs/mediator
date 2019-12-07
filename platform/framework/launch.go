package framework

import (
	"flag"
	"github.com/celskeggs/mediator/platform/icon"
	"github.com/celskeggs/mediator/platform/types"
	"github.com/celskeggs/mediator/platform/world"
	"github.com/celskeggs/mediator/platform/worldmap"
	"github.com/celskeggs/mediator/websession"
)

type ResourceDefaults struct {
	CoreResourcesDir string
	IconsDir         string
	MapPath          string
}

var mapPath = flag.String("map", "map.dmm", "the path to the game map")

func BuildWorld(tree types.TypeTree, setup func(*world.World), defaults ResourceDefaults, parseFlags bool) *world.World {
	websession.SetDefaultFlags(defaults.CoreResourcesDir, defaults.IconsDir)
	*mapPath = defaults.MapPath

	var resources string
	if parseFlags {
		_, resources, _ = websession.ParseFlags()
	} else {
		resources = defaults.IconsDir
	}

	cache := icon.NewIconCache(resources)
	gameworld := world.NewWorld(types.NewRealm(tree), cache)
	setup(gameworld)

	err := worldmap.LoadMapFromFile(gameworld, *mapPath)
	if err != nil {
		panic("cannot load world: " + err.Error())
	}
	gameworld.UpdateDefaultViewDistance()

	return gameworld
}

func Launch(tree types.TypeTree, setup func(*world.World), defaults ResourceDefaults) {
	gameworld := BuildWorld(tree, setup, defaults, true)

	websession.LaunchServerFromFlags(gameworld.ServerAPI())
}
