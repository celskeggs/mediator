package platform

import (
	"github.com/celskeggs/mediator/platform/datum"
	"github.com/celskeggs/mediator/webclient/sprite"
)

type ISound interface {
	datum.IDatum
	AsSound() *Sound
}

type Sound struct {
	datum.IDatum
	sprite.Sound
}

var _ ISound = &Sound{}

func (s Sound) RawClone() datum.IDatum {
	s.IDatum = s.IDatum.RawClone()
	return &s
}

func (s *Sound) AsSound() *Sound {
	return s
}
