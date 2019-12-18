package atoms

import (
	"github.com/celskeggs/mediator/platform/datum"
	"github.com/celskeggs/mediator/platform/types"
	"github.com/celskeggs/mediator/util"
)

//mediator:declare TurfData /turf /atom
type TurfData struct {
	X uint
	Y uint
	Z uint
}

func NewTurfData(src *types.Datum, _ *TurfData, _ ...types.Value) {
	src.SetVar("name", types.String("turf"))
	src.SetVar("layer", types.Int(TurfLayer))
}

func (t *TurfData) GetX(src *types.Datum) types.Value {
	util.FIXME("better XYZ handling for Turfs")
	return types.Int(t.X)
}

func (t *TurfData) GetY(src *types.Datum) types.Value {
	return types.Int(t.Y)
}

func (t *TurfData) GetZ(src *types.Datum) types.Value {
	return types.Int(t.Z)
}

func (t *TurfData) SetX(src *types.Datum, x types.Value) {
	util.NiceToHave("see if these can be made private, and if so, if they should be")
	t.X = uint(types.Unint(x))
}

func (t *TurfData) SetY(src *types.Datum, y types.Value) {
	t.Y = uint(types.Unint(y))
}

func (t *TurfData) SetZ(src *types.Datum, z types.Value) {
	t.Z = uint(types.Unint(z))
}

func (t *TurfData) ProcExit(src *types.Datum, usr *types.Datum, atom types.Value, newloc types.Value) types.Value {
	util.NiceToHave("call Uncross here")
	return types.Int(1)
}

func (t *TurfData) ProcEnter(src *types.Datum, usr *types.Datum, atom types.Value, oldloc types.Value) types.Value {
	util.NiceToHave("call Cross here")
	if types.AsBool(atom.Var("density")) {
		if types.AsBool(src.Var("density")) {
			atom.Invoke(usr, "Bump", src)
			return types.Int(0)
		}
		util.NiceToHave("something about only atoms that take up the full tile?")
		for _, existingAtom := range datum.Elements(src.Var("contents")) {
			if types.AsBool(existingAtom.Var("density")) {
				atom.Invoke(usr, "Bump", existingAtom)
				return types.Int(0)
			}
		}
	}
	return types.Int(1)
}

func (t *TurfData) ProcExited(src *types.Datum, usr *types.Datum, atom types.Value, newloc types.Value) types.Value {
	util.NiceToHave("call Uncrossed here")
	return nil
}

func (t *TurfData) ProcEntered(src *types.Datum, usr *types.Datum, atom types.Value, oldloc types.Value) types.Value {
	util.NiceToHave("call Crossed here")
	return nil
}
