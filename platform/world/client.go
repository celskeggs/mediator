package world

import (
	"fmt"
	"github.com/celskeggs/mediator/common"
	"github.com/celskeggs/mediator/platform/atoms"
	"github.com/celskeggs/mediator/platform/datum"
	"github.com/celskeggs/mediator/platform/types"
	"github.com/celskeggs/mediator/util"
	"github.com/celskeggs/mediator/webclient/sprite"
	"log"
	"strings"
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

func NewClientData(_ *types.Datum, _ *ClientData, _ ...types.Value) {
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
	atoms.MobSetClient(mob, src)
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
	world := atoms.WorldOf(src).(*World)
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
	clientDatum, clientData := ClientDataChunk(client)
	util.FIXME("handle defined verbs, not just built-in verbs")
	util.FIXME("support expanding partially-typed verbs")
	util.FIXME("support on-screen verb panels")
	util.FIXME("support right-clicking to access context menu verbs")
	switch verb {
	case ".north":
		client.Invoke("North")
	case ".south":
		client.Invoke("South")
	case ".east":
		client.Invoke("East")
	case ".west":
		client.Invoke("West")
	case ".verbs":
		for _, verb := range clientData.ListVerbs(clientDatum) {
			client.Invoke("<<", types.String("found verb: "+verb))
		}
	default:
		args := strings.Split(strings.TrimSpace(verb), " ")
		clientData.ResolveVerb(clientDatum, args[0], args[1:])
	}
}

func (d *ClientData) relMove(src *types.Datum, direction common.Direction) types.Value {
	var turf types.Value
	mob := d.GetMob(src)
	world := atoms.WorldOf(src)
	if mob != nil {
		x, y, z := XYZ(mob)
		dx, dy := direction.XY()
		turf = world.LocateXYZ(uint(int(x)+dx), uint(int(y)+dy), z)
	}
	if turf != nil {
		types.AssertType(turf, "/turf")
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
	return types.Int(0)
}

func (d *ClientData) OperatorWrite(src *types.Datum, output types.Value) types.Value {
	if text, ok := output.(types.String); ok {
		d.textBuffer = append(d.textBuffer, string(text))
	} else if sound, ok := output.(sprite.Sound); ok {
		sound = sound.FixMID()
		d.soundBuffer = append(d.soundBuffer, sound)
	} else {
		panic("not sure how to send output " + output.String() + " to client")
	}
	return nil
}

func ClientDataChunk(v types.Value) (*types.Datum, *ClientData) {
	impl, ok := types.Unpack(v)
	if !ok {
		panic("expected a /client, not non-datum " + v.String())
	}
	chunk := impl.Chunk("github.com/celskeggs/mediator/platform/world.ClientData")
	if chunk == nil {
		panic("expected a /client, not datum " + v.String())
	}
	return v.(*types.Datum), chunk.(*ClientData)
}

func PullClientRequests(client *types.Datum) (textDisplay []string, sounds []sprite.Sound) {
	_, d := ClientDataChunk(client)
	textDisplay, sounds = d.textBuffer, d.soundBuffer
	d.textBuffer = nil
	d.soundBuffer = nil
	return textDisplay, sounds
}

func (w *World) RenderClientViewAsAtoms(client types.Value) (center types.Value, atoms []types.Value) {
	util.FIXME("actually do this correctly")
	eye := client.Var("eye").(*types.Datum)
	veye := client.Var("virtual_eye").(*types.Datum)
	view := types.Unuint(client.Var("view"))
	return veye, w.ViewX(view, veye, eye, false)
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
		mob = atoms.WorldOf(src).(*World).constructNewMob()
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

func (d *ClientData) ResolveVerb(src *types.Datum, verbName string, args []string) {
	verbUsr := src
	for _, verbSrcVal := range atoms.WorldOf(src).FindAllType("/atom") {
		verbSrc := verbSrcVal.(*types.Datum)
		for _, verbVal := range datum.Elements(verbSrc.Var("verbs")) {
			verb := verbVal.(atoms.Verb)
			if verb.Matches(verbName, verbSrc, verbUsr) {
				resolved, err := verb.ResolveArgs(verbSrc, verbUsr, args)
				if err != nil {
					log.Printf("cannot resolve verb %v: %v\n", verb, err)
					src.Invoke("<<", types.String(fmt.Sprintf("cannot resolve verb %v: %v\n", verb, err)))
				} else {
					verb.Apply(verbSrc, verbUsr, resolved)
				}
				return
			}
		}
	}
	log.Println("got unknown verb:", verbName)
	src.Invoke("<<", types.String(fmt.Sprintf("Not a known verb: %q", verbName)))
}

func (d *ClientData) ListVerbs(src *types.Datum) (verbs []string) {
	verbUsr := src
	for _, verbSrc := range atoms.WorldOf(src).FindAllType("/atom") {
		for _, verbVal := range datum.Elements(verbSrc.Var("verbs")) {
			verb := verbVal.(atoms.Verb)
			if verb.Matches(verb.VisibleName, verbSrc.(*types.Datum), verbUsr) {
				verbs = append(verbs, verb.VisibleName)
			}
		}
	}
	return verbs
}
