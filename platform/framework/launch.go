package framework

import (
	"flag"
	"github.com/celskeggs/mediator/platform/icon"
	"github.com/celskeggs/mediator/platform/types"
	"github.com/celskeggs/mediator/platform/world"
	"github.com/celskeggs/mediator/platform/worldmap"
	"github.com/celskeggs/mediator/util"
	"github.com/celskeggs/mediator/websession"
)

type Game interface {
	BeforeMap(world *world.World)
}

type ResourceDefaults struct {
	CoreResourcesDir string
	IconsDir         string
	MapPath          string
}

var mapPath = flag.String("map", "map.dmm", "the path to the game map")

func BuildWorld(tree types.TypeTree, game Game, defaults ResourceDefaults, parseFlags bool) *world.World {
	websession.SetDefaultFlags(defaults.CoreResourcesDir, defaults.IconsDir)
	*mapPath = defaults.MapPath

	var resources string
	if parseFlags {
		_, resources, _ = websession.ParseFlags()
	} else {
		resources = defaults.IconsDir
	}
	util.FIXME("figure out where this should be used")
	_ = icon.NewIconCache(resources)

	gameworld := world.NewWorld(types.NewRealm(tree))

	game.BeforeMap(gameworld)

	err := worldmap.LoadMapFromFile(gameworld, *mapPath)
	if err != nil {
		panic("cannot load world: " + err.Error())
	}
	gameworld.UpdateDefaultViewDistance()

	return gameworld
}

func Launch(tree types.TypeTree, game Game, defaults ResourceDefaults) {
	gameworld := BuildWorld(tree, game, defaults, true)

	websession.LaunchServerFromFlags(gameworld.ServerAPI())
}
