package atom

import (
	"github.com/celskeggs/mediator/common"
	"github.com/celskeggs/mediator/platform/icon"
	"github.com/celskeggs/mediator/platform/types"
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
	Name      string
	Desc      string
}

var _ types.Value = Appearance{}

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

func (a Appearance) Var(name string) types.Value {
	util.FIXME("support opacity (and other variables)")
	switch name {
	case "desc":
		return types.String(a.Desc)
	case "icon":
		return a.Icon
	case "icon_state":
		return types.String(a.IconState)
	case "layer":
		return types.Int(a.Layer)
	case "name":
		return types.String(a.Name)
	default:
		panic("no such field " + name + " on appearance")
	}
}

func (a Appearance) SetVar(name string, value types.Value) {
	panic("cannot modify immutable appearance")
}

func (a Appearance) Invoke(name string, parameters ...types.Value) types.Value {
	panic("no such function " + name + " on appearance")
}

func (a Appearance) String() string {
	return "[appearance]"
}
