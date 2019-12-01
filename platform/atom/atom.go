package atom

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
	VarDensity    bool
	VarOpacity    bool
	VarDir        common.Direction
	location      *types.Ref
	contents      map[*types.Datum]*types.Ref
}

func NewAtomData(_ ...types.Value) AtomData {
	util.FIXME("handle location early per docs")
	return AtomData{
		VarDir: common.South,
		VarAppearance: Appearance{
			Name: "atom",
		},
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

func (d *AtomData) GetContents(src *types.Datum) types.Value {
	util.FIXME("should this really be a copy?")
	var contents []*types.Ref
	for _, ref := range d.contents {
		contents = append(contents, ref)
	}
	return datum.NewList(contents...)
}

func (d *AtomData) GetLoc(src *types.Datum) types.Value {
	return d.location.Dereference()
}

func AtomDataChunk(v types.Value) (*AtomData, bool) {
	impl, ok := types.Unpack(v)
	if !ok {
		return nil, false
	}
	chunk := impl.Chunk("github.com/celskeggs/mediator/platform/atom.AtomData")
	if chunk == nil {
		return nil, false
	}
	return chunk.(*AtomData), true
}

func (d *AtomData) SetLoc(src *types.Datum, location types.Value) {
	if d.location != nil {
		oldloc, ok := AtomDataChunk(d.location.Dereference())
		if !ok {
			panic("location of atom was not a valid atom or nil")
		}
		contents := oldloc.contents
		if _, found := contents[src]; !found {
			panic("did not find self in location's contents")
		}
		delete(contents, src)
	}
	d.location = nil
	if location != nil {
		newloc, ok := AtomDataChunk(location)
		if !ok {
			panic("attempt to move atom to non-atom location")
		}
		d.location = location.Reference()
		if newloc.contents == nil {
			newloc.contents = map[*types.Datum]*types.Ref{}
		}
		if _, found := newloc.contents[src]; found {
			panic("should not have found self in new location's contents")
		}
		newloc.contents[src] = src.Reference()
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

func (d *AtomData) ProcExit(src *types.Datum, atom types.Value, newloc types.Value) types.Value {
	// always allow by default
	return types.Bool(true)
}

func (d *AtomData) ProcEnter(src *types.Datum, atom types.Value, oldloc types.Value) types.Value {
	// always allow by default
	return types.Bool(true)
}

func (d *AtomData) ProcExited(src *types.Datum, atom types.Value, newloc types.Value) types.Value {
	// nothing to do
	return nil
}

func (d *AtomData) ProcEntered(src *types.Datum, atom types.Value, oldloc types.Value) types.Value {
	// nothing to do
	return nil
}
