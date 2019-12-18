package icon

import (
	"github.com/celskeggs/mediator/common"
	"github.com/celskeggs/mediator/dmi"
	"github.com/celskeggs/mediator/platform/types"
	"github.com/celskeggs/mediator/util"
)

type Icon struct {
	dmiPath      string
	dmiInfo      *dmi.DMIInfo
	stateIndexes map[string]uint
	stride       uint
}

func directionToIndex(direction common.Direction, dirs int) uint {
	if dirs == 1 {
		return 0
	} else if dirs == 4 {
		return direction.NearestCardinal().FourDirectionIndex()
	} else if dirs == 8 {
		return direction.EightDirectionIndex()
	} else {
		panic("unexpected 'directions' value should have been caught by validate() earlier")
	}
}

func (icon *Icon) lookupIndex(state string, direction common.Direction, frame uint) (index uint) {
	dmiState := icon.dmiInfo.States[state]
	if frame >= uint(dmiState.Frames) {
		util.FIXME("don't panic here")
		panic("invalid frame number")
	}
	return icon.stateIndexes[state] +
		frame*uint(dmiState.Directions) +
		directionToIndex(direction, dmiState.Directions)
}

func (icon *Icon) indexToPosition(index, width, height uint) (x, y uint) {
	if icon.stride == 0 {
		return 0, 0
	}
	return (index % icon.stride) * width, (index / icon.stride) * height
}

func (icon *Icon) Render(state string, dir common.Direction) (iconname string, sourceX, sourceY, sourceWidth, sourceHeight uint) {
	util.FIXME("implement animations")
	index := icon.lookupIndex(state, dir, 0)
	util.FIXME("implement icon sizes correctly")
	sourceX, sourceY = icon.indexToPosition(index, 32, 32)
	return icon.dmiPath, sourceX, sourceY, 32, 32
}

var _ types.Value = &Icon{}

func (icon *Icon) Var(name string) types.Value {
	panic("no such var " + name + " on icon")
}

func (icon *Icon) SetVar(name string, value types.Value) {
	panic("no such var " + name + " on icon")
}

func (icon *Icon) Invoke(usr *types.Datum, name string, parameters ...types.Value) types.Value {
	util.FIXME("implement Icon procedures")
	panic("no such proc " + name + " on icon")
}

func (icon *Icon) String() string {
	return "[icon]"
}
