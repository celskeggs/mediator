package framework

import (
	"flag"
	"github.com/celskeggs/mediator/platform/icon"
	"github.com/celskeggs/mediator/platform/types"
	"github.com/celskeggs/mediator/platform/world"
	"github.com/celskeggs/mediator/platform/worldmap"
	"github.com/celskeggs/mediator/resourcepack"
	"github.com/celskeggs/mediator/websession"
)

type ResourceDefaults struct {
	MapPath          string
	ResourcePackPath string
}

var mapPath = flag.String("map", "map.dmm", "the path to the game map")

func BuildWorld(tree types.TypeTree, setup func(*world.World), defaults ResourceDefaults, parseFlags bool) (*world.World, *resourcepack.ResourcePack) {
	*mapPath = defaults.MapPath

	var packPath string
	if parseFlags {
		packPath = websession.FindResourcePack()
	} else {
		packPath = defaults.ResourcePackPath
	}

	pack, err := resourcepack.Load(packPath)
	if err != nil {
		panic("cannot load resource pack: " + err.Error())
	}

	cache, err := icon.NewIconCache(pack)
	if err != nil {
		panic("cannot load icon cache: " + err.Error())
	}
	gameworld := world.NewWorld(types.NewRealm(tree), cache)
	setup(gameworld)

	err = worldmap.LoadMapFromPack(gameworld, pack, *mapPath)
	if err != nil {
		panic("cannot load world: " + err.Error())
	}
	gameworld.UpdateDefaultViewDistance()

	return gameworld, pack
}

func Launch(tree types.TypeTree, setup func(*world.World), defaults ResourceDefaults) {
	gameworld, pack := BuildWorld(tree, setup, defaults, true)

	err := websession.LaunchServer(gameworld.ServerAPI(), pack)
	if err != nil {
		panic("error in server: " + err.Error())
	}
	panic("server exited unexpectedly")
}
