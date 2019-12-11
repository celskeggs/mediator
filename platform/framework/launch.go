package framework

import (
	"github.com/celskeggs/mediator/platform/icon"
	"github.com/celskeggs/mediator/platform/types"
	"github.com/celskeggs/mediator/platform/world"
	"github.com/celskeggs/mediator/platform/worldmap"
	"github.com/celskeggs/mediator/resourcepack"
	"github.com/celskeggs/mediator/websession"
)

type SetupFunc func(*world.World) (mapNames []string)

func BuildWorld(tree types.TypeTree, setup SetupFunc) (*world.World, *resourcepack.ResourcePack) {
	pack, err := websession.LoadResourcePack()
	if err != nil {
		panic("cannot load resource pack: " + err.Error())
	}

	cache, err := icon.NewIconCache(pack)
	if err != nil {
		panic("cannot load icon cache: " + err.Error())
	}
	gameworld := world.NewWorld(types.NewRealm(tree), cache)
	maps := setup(gameworld)
	if len(maps) > 1 {
		panic("unimplemented: more than one map")
	}

	if len(maps) > 0 {
		err = worldmap.LoadMapFromPack(gameworld, pack, maps[0])
		if err != nil {
			panic("cannot load world: " + err.Error())
		}
		gameworld.UpdateDefaultViewDistance()
	}

	return gameworld, pack
}

func Launch(tree types.TypeTree, setup SetupFunc) {
	gameworld, pack := BuildWorld(tree, setup)

	err := websession.LaunchServer(gameworld.ServerAPI(), pack)
	if err != nil {
		panic("error in server: " + err.Error())
	}
	panic("server exited unexpectedly")
}
