package atoms

import (
	"github.com/celskeggs/mediator/common"
	"github.com/celskeggs/mediator/platform/types"
	"github.com/celskeggs/mediator/util"
)

type WalkState struct {
	WalkTarget    types.Value
	WalkMinimum   int
	WalkCountdown int
	WalkLagTicks  int
}

//mediator:declare AtomMovableData /atom/movable /atom
type AtomMovableData struct {
	walkState WalkState
}

func NewAtomMovableData(src *types.Datum, _ *AtomMovableData, _ ...types.Value) {
	src.SetVar("name", types.String("movable"))
}

func ContainingArea(atom types.Value) types.Value {
	for atom != nil {
		if types.IsType(atom, "/area") {
			return atom
		}
		atom = atom.Var("loc")
	}
	return nil
}

func (d *AtomData) ProcMove(src *types.Datum, usr *types.Datum, newloc types.Value, direction types.Value) types.Value {
	util.NiceToHave("implement pixel movement/slides")
	util.FIXME("update directions appropriately")

	oldloc := src.Var("loc")
	oldarea := ContainingArea(src)
	if direction != nil && direction.(common.Direction) != 0 {
		src.SetVar("dir", direction)
	}
	if newloc != oldloc && newloc != nil {
		newarea := ContainingArea(newloc)
		if oldloc != nil {
			if !types.AsBool(oldloc.Invoke(usr, "Exit", src, newloc)) {
				return types.Int(0)
			}
			util.NiceToHave("handle Cross and Uncross and Crossed and Uncrossed")
		}
		if newarea != oldarea && oldarea != nil {
			if !types.AsBool(oldarea.Invoke(usr, "Exit", src, newarea)) {
				return types.Int(0)
			}
		}
		if !types.AsBool(newloc.Invoke(usr, "Enter", src, oldloc)) {
			return types.Int(0)
		}
		if newarea != oldarea && newarea != nil {
			if !types.AsBool(newarea.Invoke(usr, "Enter", src, oldarea)) {
				return types.Int(0)
			}
		}
		src.SetVar("loc", newloc)
		if oldloc != nil {
			oldloc.Invoke(usr, "Exited", src, newloc)
		}
		if newarea != oldarea && oldarea != nil {
			oldarea.Invoke(usr, "Exited", src, newarea)
		}
		newloc.Invoke(usr, "Entered", src, oldloc)
		if newarea != oldarea && newarea != nil {
			newarea.Invoke(usr, "Entered", src, oldarea)
		}
	}
	return types.Int(1)
}

func (d *AtomData) ProcBump(src *types.Datum, usr *types.Datum, obstacle types.Value) types.Value {
	// nothing to do in the general case
	util.NiceToHave("group support for mob bumping")
	return nil
}

func AtomMovableDataChunk(v types.Value) (*AtomMovableData, bool) {
	impl, ok := types.Unpack(v)
	if !ok {
		return nil, false
	}
	chunk := impl.Chunk("github.com/celskeggs/mediator/platform/atoms.AtomMovableData")
	if chunk == nil {
		return nil, false
	}
	return chunk.(*AtomMovableData), true
}

// not for general use; just for pathfinding system
func GetWalkState(src types.Value) WalkState {
	d, ok := AtomMovableDataChunk(src)
	if !ok {
		panic("should always be an /atom/movable in UpdateWalk!")
	}
	return d.walkState
}

// not for general use; just for pathfinding system
func SetWalkState(src types.Value, ws WalkState) {
	d, ok := AtomMovableDataChunk(src)
	if !ok {
		panic("should always be an /atom/movable in UpdateWalk!")
	}
	d.walkState = ws
}
