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

func (icon *Icon) countFrames(state string) int {
	return icon.dmiInfo.States[state].Frames
}

func (icon *Icon) lookupIndex(state string, direction common.Direction, frame int) (index uint) {
	dmiState := icon.dmiInfo.States[state]
	if frame < 0 || frame >= dmiState.Frames {
		util.FIXME("don't panic here")
		panic("invalid frame number")
	}
	return icon.stateIndexes[state] +
		uint(frame*dmiState.Directions) +
		directionToIndex(direction, dmiState.Directions)
}

type SourceXY struct {
	X uint `json:"x"`
	Y uint `json:"y"`
}

func (icon *Icon) indexToPosition(index, width, height uint) SourceXY {
	if icon.stride == 0 {
		return SourceXY{
			X: 0,
			Y: 0,
		}
	}
	return SourceXY{
		X: (index % icon.stride) * width,
		Y: (index / icon.stride) * height,
	}
}

func (icon *Icon) Render(state string, dir common.Direction) (iconname string, frames []SourceXY, sourceWidth, sourceHeight uint) {
	util.FIXME("implement icon sizes correctly")
	frames = make([]SourceXY, icon.countFrames(state))
	if len(frames) == 0 {
		panic("should never be NO frames to render")
	}
	for i := 0; i < len(frames); i++ {
		index := icon.lookupIndex(state, dir, i)
		frames[i] = icon.indexToPosition(index, 32, 32)
	}
	return icon.dmiPath, frames, 32, 32
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
