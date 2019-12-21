package atoms

import (
	"github.com/celskeggs/mediator/common"
	"github.com/celskeggs/mediator/platform/datum"
	"github.com/celskeggs/mediator/platform/icon"
	"github.com/celskeggs/mediator/platform/types"
	"github.com/celskeggs/mediator/util"
)

//mediator:declare AtomData /atom /datum
type AtomData struct {
	VarAppearance Appearance
	VarDensity    int
	VarOpacity    int
	VarDir        common.Direction
	VarVerbs      []Verb
	location      *types.Ref
	contents      []*types.Ref
}

func NewAtomData(src *types.Datum, data *AtomData, args ...types.Value) {
	data.VarDir = common.South
	data.VarAppearance = Appearance{
		Name: "atom",
	}
	if len(args) >= 1 {
		data.SetLoc(src, args[0])
	}
}

func (d *AtomData) GetDesc(src *types.Datum) types.Value {
	return types.String(d.VarAppearance.Desc)
}

func (d *AtomData) SetDesc(src *types.Datum, value types.Value) {
	d.VarAppearance.Desc = types.Unstring(value)
}

func (d *AtomData) GetIcon(src *types.Datum) types.Value {
	return d.VarAppearance.Icon
}

func (d *AtomData) SetIcon(src *types.Datum, value types.Value) {
	d.VarAppearance.Icon = value.(*icon.Icon)
}

func (d *AtomData) GetIconState(src *types.Datum) types.Value {
	return types.String(d.VarAppearance.IconState)
}

func (d *AtomData) SetIconState(src *types.Datum, value types.Value) {
	d.VarAppearance.IconState = types.Unstring(value)
}

func (d *AtomData) GetLayer(src *types.Datum) types.Value {
	return types.Int(d.VarAppearance.Layer)
}

func (d *AtomData) SetLayer(src *types.Datum, value types.Value) {
	d.VarAppearance.Layer = types.Unint(value)
}

func (d *AtomData) GetName(src *types.Datum) types.Value {
	return types.String(d.VarAppearance.Name)
}

func (d *AtomData) SetName(src *types.Datum, value types.Value) {
	d.VarAppearance.Name = types.Unstring(value)
}

func (d *AtomData) GetSuffix(src *types.Datum) types.Value {
	return types.String(d.VarAppearance.Suffix)
}

func (d *AtomData) SetSuffix(src *types.Datum, value types.Value) {
	d.VarAppearance.Suffix = types.Unstring(value)
}

func (d *AtomData) GetContents(src *types.Datum) types.Value {
	util.FIXME("should this really be a copy?")
	contents := make([]*types.Ref, len(d.contents))
	copy(contents, d.contents)
	return datum.NewListFromRefs(contents...)
}

func (d *AtomData) GetLoc(src *types.Datum) types.Value {
	return d.location.Dereference()
}

func AtomDataChunk(v types.Value) (*AtomData, bool) {
	impl, ok := types.Unpack(v)
	if !ok {
		return nil, false
	}
	chunk := impl.Chunk("github.com/celskeggs/mediator/platform/atoms.AtomData")
	if chunk == nil {
		return nil, false
	}
	return chunk.(*AtomData), true
}

func removeFromContents(contents []*types.Ref, remove *types.Datum) []*types.Ref {
	for i, elem := range contents {
		if elem.Dereference() == remove {
			copy(contents[i:], contents[i+1:])
			contents = contents[:len(contents)-1]
			return contents
		}
	}
	panic("did not find expected atom in contents")
}

func (d *AtomData) SetLoc(src *types.Datum, location types.Value) {
	if d.location != nil {
		oldloc, ok := AtomDataChunk(d.location.Dereference())
		if !ok {
			panic("location of atom was not a valid atom or nil")
		}
		oldloc.contents = removeFromContents(oldloc.contents, src)
	}
	d.location = nil
	if location != nil {
		newloc, ok := AtomDataChunk(location)
		if !ok {
			panic("attempt to move atom to non-atom location: " + location.String())
		}
		d.location = types.Reference(location)
		for _, elem := range newloc.contents {
			if elem.Dereference() == src {
				panic("should not have found self in new location's contents")
			}
		}
		newloc.contents = append(newloc.contents, types.Reference(src))
	}
}

func (d *AtomData) GetX(src *types.Datum) types.Value {
	util.FIXME("x, y, z should not be read-only")
	if d.location == nil {
		return types.Int(0)
	} else {
		return d.location.Dereference().Var("x")
	}
}

func (d *AtomData) GetY(src *types.Datum) types.Value {
	if d.location == nil {
		return types.Int(0)
	} else {
		return d.location.Dereference().Var("y")
	}
}

func (d *AtomData) GetZ(src *types.Datum) types.Value {
	if d.location == nil {
		return types.Int(0)
	} else {
		return d.location.Dereference().Var("z")
	}
}

func (d *AtomData) ProcExit(src *types.Datum, usr *types.Datum, atom types.Value, newloc types.Value) types.Value {
	// always allow by default
	return types.Int(1)
}

func (d *AtomData) ProcEnter(src *types.Datum, usr *types.Datum, atom types.Value, oldloc types.Value) types.Value {
	// always allow by default
	return types.Int(1)
}

func (d *AtomData) ProcExited(src *types.Datum, usr *types.Datum, atom types.Value, newloc types.Value) types.Value {
	// nothing to do
	return nil
}

func (d *AtomData) ProcEntered(src *types.Datum, usr *types.Datum, atom types.Value, oldloc types.Value) types.Value {
	// nothing to do
	return nil
}

func (d *AtomData) ProcStat(src *types.Datum, usr *types.Datum) types.Value {
	// nothing to do
	return nil
}
