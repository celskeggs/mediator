package websession

import "flag"

var coreResources = flag.String("core", "resources", "the path to the core resources/ directory")
var extraResources = flag.String("resources", "icons", "the path to the game-specific icons/ directory")
var cacheResources = flag.String("cache", "cache", "the path to the cache/ directory")
var parsed = false

func SetDefaultFlags(coreResourcesDir, extraResourcesDir string) {
	if parsed {
		panic("already parsed flags")
	}
	*coreResources = coreResourcesDir
	*extraResources = extraResourcesDir
}

func ParseFlags() (corePath, extraPath, cachePath string) {
	if !parsed {
		flag.Parse()
		parsed = true
	}
	return *coreResources, *extraResources, *cacheResources
}

// does not return
func LaunchServerFromFlags(api WorldAPI) {
	core, extra, cache := ParseFlags()
	err := LaunchServer(api, core, extra, cache)
	if err != nil {
		panic("error in server: " + err.Error())
	}
	panic("server exited unexpectedly")
}
