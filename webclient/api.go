package webclient

import "github.com/celskeggs/mediator/webclient/sprite"

type ServerSession interface {
	Close()
	OnMessage(Command)
	// send nil to close connection
	BeginSend(func(*sprite.SpriteView) error)
}

type ServerAPI interface {
	Connect() ServerSession
	CoreResourcePath() string
	ListResources() (nameToPath map[string]string, err error)
}
