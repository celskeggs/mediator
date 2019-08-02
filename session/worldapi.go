package session

// WorldAPI can be single-threaded; Session will not call any function until the last call returned.
// This is true across both interfaces.
type PlayerAPI interface {
	Remove()
	Command(cmd Command)
	Render() SpriteView
}

type WorldAPI interface {
	AddPlayer() PlayerAPI
	// only a single call to SubscribeToUpdates needs to be supported by the WorldAPI
	SubscribeToUpdates() chan struct{}
}
