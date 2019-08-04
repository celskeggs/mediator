package platform

import (
	"github.com/celskeggs/mediator/common"
	"github.com/celskeggs/mediator/platform/icon"
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
	Icon      *icon.Icon
	IconState string
	Layer     int
}

func (a Appearance) ToSprite(x, y uint, dir common.Direction) (bool, int, sprite.GameSprite) {
	util.FIXME("implement correct sizing")
	if a.Icon == nil {
		return false, 0, sprite.GameSprite{}
	}
	iconName, sourceX, sourceY, sourceWidth, sourceHeight := a.Icon.Render(a.IconState, dir)
	return true, a.Layer, sprite.GameSprite{
		Icon:         iconName,
		SourceX:      sourceX,
		SourceY:      sourceY,
		SourceWidth:  sourceWidth,
		SourceHeight: sourceHeight,
		X:            x,
		Y:            y,
		Width:        32,
		Height:       32,
	}
}
