package framework

import (
	"flag"
	"github.com/celskeggs/mediator/platform"
	"github.com/celskeggs/mediator/platform/icon"
	"github.com/celskeggs/mediator/platform/types"
	"github.com/celskeggs/mediator/platform/worldmap"
	"github.com/celskeggs/mediator/websession"
)

type Game interface {
	Definer() platform.TreeDefiner
	ElaborateTree(*types.TypeTree, *icon.IconCache)
	BeforeMap(world *platform.World)
}

type ResourceDefaults struct {
	CoreResourcesDir string
	IconsDir         string
	MapPath          string
}

var mapPath = flag.String("map", "map.dmm", "the path to the game map")

func BuildWorld(game Game, defaults ResourceDefaults, parseFlags bool) *platform.World {
	websession.SetDefaultFlags(defaults.CoreResourcesDir, defaults.IconsDir)
	*mapPath = defaults.MapPath

	var resources string
	if parseFlags {
		_, resources, _ = websession.ParseFlags()
	} else {
		resources = defaults.IconsDir
	}
	tree := platform.NewAtomicTree(game.Definer())
	icons := icon.NewIconCache(resources)
	game.ElaborateTree(tree, icons)

	world := platform.NewWorld(tree)

	game.BeforeMap(world)

	err := worldmap.LoadMapFromFile(world, *mapPath)
	if err != nil {
		panic("cannot load world: " + err.Error())
	}
	world.UpdateDefaultViewDistance()

	return world
}

func Launch(game Game, defaults ResourceDefaults) {
	world := BuildWorld(game, defaults, true)

	websession.LaunchServerFromFlags(world.ServerAPI())
}
