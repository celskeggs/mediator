package websession

import (
	"flag"
	"github.com/celskeggs/mediator/resourcepack"
)

var resourcePack = flag.String("pack", "resource_pack.tgz", "the path to the game's resource pack")
var parsed = false

func FindResourcePack() string {
	if !parsed {
		flag.Parse()
		parsed = true
	}
	return *resourcePack
}

func LoadResourcePack() (*resourcepack.ResourcePack, error) {
	return resourcepack.Load(FindResourcePack())
}

// does not return
func LaunchServerFromFlags(api WorldAPI) {
	pack, err := LoadResourcePack()
	if err != nil {
		panic("error loading resource pack: " + err.Error())
	}
	err = LaunchServer(api, pack)
	if err != nil {
		panic("error in server: " + err.Error())
	}
	panic("server exited unexpectedly")
}
