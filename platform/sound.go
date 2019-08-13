package platform

import "github.com/celskeggs/mediator/platform/datum"

type ISound interface {
	datum.IDatum
	AsSound() *Sound
}

type Sound struct {
	datum.IDatum
}

var _ ISound = &Sound{}

func (s Sound) RawClone() datum.IDatum {
	s.IDatum = s.IDatum.RawClone()
	return &s
}

func (s *Sound) AsSound() *Sound {
	return s
}
