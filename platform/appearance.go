package platform

import (
	"github.com/celskeggs/mediator/util"
	"github.com/celskeggs/mediator/webclient/sprite"
)

const (
	AreaLayer = 1
	TurfLayer = 2
	ObjLayer  = 3
	MobLayer  = 4
)

type Appearance struct {
	Icon      string
	IconState string
	Layer     int
}

func (a Appearance) ToSprite(x, y uint) (int, sprite.GameSprite) {
	util.FIXME("implement IconState and correct sizing")
	return a.Layer, sprite.GameSprite{
		Icon:         a.Icon,
		X:            x,
		Y:            y,
		SourceWidth:  32,
		SourceHeight: 32,
	}
}
