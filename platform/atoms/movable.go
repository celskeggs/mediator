package atoms

import (
	"github.com/celskeggs/mediator/common"
	"github.com/celskeggs/mediator/platform/types"
	"github.com/celskeggs/mediator/util"
)

//mediator:declare AtomMovableData /atom/movable /atom
type AtomMovableData struct{}

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

func (d *AtomData) ProcMove(src *types.Datum, newloc types.Value, direction types.Value) types.Value {
	util.NiceToHave("implement pixel movement/slides")

	oldloc := src.Var("loc")
	oldarea := ContainingArea(src)
	if direction.(common.Direction) != 0 {
		src.SetVar("dir", direction)
	}
	if newloc != oldloc && newloc != nil {
		newarea := ContainingArea(newloc)
		if oldloc != nil {
			if !types.AsBool(oldloc.Invoke("Exit", src, newloc)) {
				return types.Int(0)
			}
			util.NiceToHave("handle Cross and Uncross and Crossed and Uncrossed")
		}
		if newarea != oldarea && oldarea != nil {
			if !types.AsBool(oldarea.Invoke("Exit", src, newarea)) {
				return types.Int(0)
			}
		}
		if !types.AsBool(newloc.Invoke("Enter", src, oldloc)) {
			util.FIXME("bump obstacles")
			return types.Int(0)
		}
		if newarea != oldarea && newarea != nil {
			if !types.AsBool(newarea.Invoke("Enter", src, oldarea)) {
				return types.Int(0)
			}
		}
		src.SetVar("loc", newloc)
		if oldloc != nil {
			oldloc.Invoke("Exited", src, newloc)
		}
		if newarea != oldarea && oldarea != nil {
			oldarea.Invoke("Exited", src, newarea)
		}
		newloc.Invoke("Entered", src, oldloc)
		if newarea != oldarea && newarea != nil {
			newarea.Invoke("Entered", src, oldarea)
		}
	}
	return types.Int(1)
}

func (d *AtomData) ProcBump(src *types.Datum, obstacle types.Value) types.Value {
	// nothing to do in the general case
	util.NiceToHave("group support for mob bumping")
	return nil
}
