package webclient

import "github.com/celskeggs/mediator/webclient/sprite"

type ServerSession interface {
	Close()
	OnMessage(Command)
	// send nil to the view send callback to close connection
	BeginSend(func(update *sprite.ViewUpdate) error)
}

type ServerAPI interface {
	Connect() ServerSession
	CoreResourcePath() string
	ListResources() (nameToPath map[string]string, download []string, err error)
}
