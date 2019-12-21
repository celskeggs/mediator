package atoms

import (
	"github.com/celskeggs/mediator/platform/types"
	"github.com/celskeggs/mediator/webclient/sprite"
)

type StatContext struct {
	display sprite.StatDisplay
}

func (s *StatContext) Stat(name string, value types.Value) {
	panic("unimplemented")
}

func (s *StatContext) StatPanel(panel string) bool {
	panic("unimplemented")
}

func (s *StatContext) Display() sprite.StatDisplay {
	return s.display
}
