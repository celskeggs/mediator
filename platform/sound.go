package platform

import (
	"github.com/celskeggs/mediator/webclient/sprite"
)

func NewSound(file string) sprite.Sound {
	return sprite.Sound{
		File:    file,
		Repeat:  false,
		Wait:    false,
		Channel: 0,
		Volume:  100,
	}
}
