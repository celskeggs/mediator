package websession

import (
	"github.com/celskeggs/mediator/webclient"
	"github.com/celskeggs/mediator/webclient/sprite"
)

// WorldAPI can be single-threaded; Session will not call any function until the last call returned.
// This is true across both interfaces.
type PlayerAPI interface {
	Remove()
	IsValid() bool
	Command(cmd webclient.Command)
	Render() sprite.SpriteView
	PullRequests() (lines []string, sounds []sprite.Sound, flicks []sprite.Flick)
}

type WorldAPI interface {
	AddPlayer(key string) PlayerAPI
	Tick()
	// only a single call to SubscribeToUpdates needs to be supported by the WorldAPI
	SubscribeToUpdates() <-chan struct{}
}
