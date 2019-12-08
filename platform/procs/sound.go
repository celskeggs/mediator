package procs

import (
	"github.com/celskeggs/mediator/platform/types"
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

func NewSoundFull(file types.Value, repeat types.Value, wait types.Value, channel types.Value, volume types.Value) sprite.Sound {
	var filename string
	if s, ok := file.(types.String); ok {
		filename = types.Unstring(s)
	} else {
		filename = file.(sprite.Sound).File
	}
	var chnum uint
	if channel != nil {
		chnum = types.Unuint(channel)
	}
	vol := uint(100)
	if volume != nil {
		vol = types.Unuint(volume)
	}
	return sprite.Sound{
		File:    filename,
		Repeat:  types.AsBool(repeat),
		Wait:    types.AsBool(wait),
		Channel: chnum,
		Volume:  vol,
	}
}

func NewSoundFrom(base sprite.Sound, repeat bool, wait bool, channel uint, volume uint) sprite.Sound {
	return sprite.Sound{
		File:    base.File,
		Repeat:  repeat,
		Wait:    wait,
		Channel: channel,
		Volume:  volume,
	}
}
