package procs

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

func NewSoundFull(file string, repeat bool, wait bool, channel uint, volume uint) sprite.Sound {
	return sprite.Sound{
		File:    file,
		Repeat:  repeat,
		Wait:    wait,
		Channel: channel,
		Volume:  volume,
	}
}
