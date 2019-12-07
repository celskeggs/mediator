package world

import (
	"github.com/celskeggs/mediator/common"
	"github.com/celskeggs/mediator/platform/atom"
	"github.com/celskeggs/mediator/platform/types"
	"github.com/celskeggs/mediator/util"
	"github.com/celskeggs/mediator/webclient/sprite"
	"log"
)

//mediator:declare ClientData /client /datum
type ClientData struct {
	VarKey      string
	VarView     int
	mob         *types.Ref
	eye         *types.Ref
	textBuffer  []string
	soundBuffer []sprite.Sound
}

func NewClientData(_ *types.Datum, _ ...types.Value) ClientData {
	return ClientData{}
}

func (d *ClientData) GetMob(src *types.Datum) types.Value {
	return d.mob.Dereference()
}

func (d *ClientData) SetMob(src *types.Datum, mob types.Value) {
	if !types.IsType(mob, "/mob") {
		panic("attempt to set client mob to not a /mob")
	}
	d.mob = types.Reference(mob)
	d.SetEye(src, mob)
	atom.MobSetClient(mob, src)
}

func (d *ClientData) GetEye(src *types.Datum) types.Value {
	if d.eye != nil {
		return d.eye.Dereference()
	} else {
		return nil
	}
}

func (d *ClientData) SetEye(src *types.Datum, eye types.Value) {
	if !types.IsType(eye, "/atom") {
		panic("attempt to set client eye to not an /atom")
	}
	d.eye = types.Reference(eye)
}

func (d *ClientData) GetVirtualEye(src *types.Datum) types.Value {
	world := atom.WorldOf(src).(*World)
	maxX, maxY, _ := world.MaxXYZ()
	if world.setVirtualEye {
		eyeZ := types.Unuint(d.GetEye(src).Var("z"))
		turf := world.LocateXYZ((maxX+1)/2, (maxY+1)/2, eyeZ)
		if turf != nil {
			return turf
		}
	}
	return d.GetEye(src)
}

func InvokeVerb(client types.Value, verb string) {
	switch verb {
	case ".north":
		client.Invoke("North")
	case ".south":
		client.Invoke("South")
	case ".east":
		client.Invoke("East")
	case ".west":
		client.Invoke("West")
	default:
		log.Println("got unknown verb:", verb)
	}
}

func (d *ClientData) relMove(src *types.Datum, direction common.Direction) types.Value {
	var turf types.Value
	mob := d.GetMob(src)
	world := atom.WorldOf(src)
	if mob != nil {
		x, y, z := XYZ(mob)
		dx, dy := direction.XY()
		turf = world.LocateXYZ(uint(int(x)+dx), uint(int(y)+dy), z)
	}
	return src.Invoke("Move", turf, direction)
}

func (d *ClientData) ProcNorth(src *types.Datum) types.Value {
	return d.relMove(src, common.North)
}

func (d *ClientData) ProcSouth(src *types.Datum) types.Value {
	return d.relMove(src, common.South)
}

func (d *ClientData) ProcEast(src *types.Datum) types.Value {
	return d.relMove(src, common.East)
}

func (d *ClientData) ProcWest(src *types.Datum) types.Value {
	return d.relMove(src, common.West)
}

func (d *ClientData) ProcMove(src *types.Datum, loc types.Value, dir types.Value) types.Value {
	mob := d.GetMob(src)
	util.FIXME("cancel automated movement if necessary")
	if mob != nil {
		return mob.Invoke("Move", loc, dir)
	}
	return types.Bool(false)
}

func (d *ClientData) OperatorWrite(src *types.Datum, output types.Value) types.Value {
	if text, ok := output.(types.String); ok {
		d.textBuffer = append(d.textBuffer, string(text))
	} else if sound, ok := output.(sprite.Sound); ok {
		d.soundBuffer = append(d.soundBuffer, sound)
	} else {
		panic("not sure how to send output " + output.String() + " to client")
	}
	return nil
}

func ClientDataChunk(v types.Value) (*ClientData, bool) {
	impl, ok := types.Unpack(v)
	if !ok {
		return nil, false
	}
	chunk := impl.Chunk("github.com/celskeggs/mediator/platform/world.ClientData")
	if chunk == nil {
		return nil, false
	}
	return chunk.(*ClientData), true
}

func PullClientRequests(client *types.Datum) (textDisplay []string, sounds []sprite.Sound) {
	d, ok := ClientDataChunk(client)
	if !ok {
		panic("attempt to PullClientRequests on something that's not a /client")
	}
	textDisplay, sounds = d.textBuffer, d.soundBuffer
	d.textBuffer = nil
	d.soundBuffer = nil
	return textDisplay, sounds
}

func (w *World) RenderClientViewAsAtoms(client types.Value) (center types.Value, atoms []*types.Datum) {
	util.FIXME("actually do this correctly")
	eye := client.Var("eye").(*types.Datum)
	veye := client.Var("virtual_eye").(*types.Datum)
	view := types.Unuint(client.Var("view"))
	return veye, w.ViewX(view, veye, eye)
}

func (w *World) constructNewMob() types.Value {
	mob := w.Realm().New(w.Mob)
	if !types.IsType(mob, "/mob") {
		panic("constructed mob is not a /mob")
	}
	util.FIXME("initialize name and gender")
	return mob
}

func (w *World) findExistingMob(key string) types.Value {
	if key == "" {
		return nil
	}
	return w.FindOne(func(atom *types.Datum) bool {
		return types.IsType(atom, "/mob") && types.Unstring(atom.Var("key")) == key
	})
}

func (d *ClientData) ProcNew(src *types.Datum, usr types.Value) types.Value {
	mob := usr
	util.FIXME("add support for 'prototype mobs'")
	if mob == nil {
		mob = atom.WorldOf(src).(*World).constructNewMob()
	}
	util.NiceToHave("add support for Topics")
	d.SetMob(src, mob)
	mob.Invoke("Login")
	return mob
}

func (d *ClientData) ProcDel(src *types.Datum) types.Value {
	util.FIXME("call Logout() on mob")
	util.FIXME("should killing the connection go here, maybe in addition to other places?")
	return nil
}
