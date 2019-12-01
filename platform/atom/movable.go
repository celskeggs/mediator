package atom

import (
	"github.com/celskeggs/mediator/common"
	"github.com/celskeggs/mediator/platform/types"
	"github.com/celskeggs/mediator/util"
)

//mediator:declare AtomMovableData /atom/movable /atom
type AtomMovableData struct{}

func NewAtomMovableData(_ ...types.Value) AtomMovableData {
	return AtomMovableData{}
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
			if !types.Unbool(oldloc.Invoke("Exit", src, newloc)) {
				return types.Bool(false)
			}
			util.NiceToHave("handle Cross and Uncross and Crossed and Uncrossed")
		}
		if newarea != oldarea && oldarea != nil {
			if !types.Unbool(oldarea.Invoke("Exit", src, newarea)) {
				return types.Bool(false)
			}
		}
		if !types.Unbool(newloc.Invoke("Enter", src, oldloc)) {
			util.FIXME("bump obstacles")
			return types.Bool(false)
		}
		if newarea != oldarea && newarea != nil {
			if !types.Unbool(newarea.Invoke("Enter", src, oldarea)) {
				return types.Bool(false)
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
	return types.Bool(true)
}

func (d *AtomData) ProcBump(src *types.Datum, obstacle types.Value) types.Value {
	// nothing to do in the general case
	util.NiceToHave("group support for mob bumping")
	return nil
}
