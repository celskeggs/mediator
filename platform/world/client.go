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
	"sort"
	"strings"
)

//mediator:declare ClientData /client /datum
type ClientData struct {
	VarKey      string
	VarView     int
	VarStatobj  *types.Ref
	mob         *types.Ref
	eye         *types.Ref
	textBuffer  []string
	soundBuffer []sprite.Sound
	flicks      []sprite.Flick
	statDisplay sprite.StatDisplay
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
	util.FIXME("support expanding partially-typed verbs")
	mob := clientData.GetMob(clientDatum)
	var mobDatum *types.Datum
	if mob != nil {
		mobDatum = mob.(*types.Datum)
	}
	switch verb {
	case ".north":
		client.Invoke(mobDatum, "North")
	case ".south":
		client.Invoke(mobDatum, "South")
	case ".east":
		client.Invoke(mobDatum, "East")
	case ".west":
		client.Invoke(mobDatum, "West")
	case ".verbs":
		client.Invoke(mobDatum, "<<", types.String("looking for verbs..."))
		verbs, _ := clientData.ListVerbs(clientDatum)
		for _, verb := range verbs {
			client.Invoke(mobDatum, "<<", types.String("found verb: "+verb))
		}
	default:
		args := strings.Split(strings.TrimSpace(verb), " ")
		clientData.ResolveVerb(clientDatum, args[0], args[1:])
	}
}

func (d *ClientData) relMove(src *types.Datum, usr *types.Datum, direction common.Direction) types.Value {
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
	return src.Invoke(usr, "Move", turf, direction)
}

func (d *ClientData) ProcNorth(src *types.Datum, usr *types.Datum) types.Value {
	return d.relMove(src, usr, common.North)
}

func (d *ClientData) ProcSouth(src *types.Datum, usr *types.Datum) types.Value {
	return d.relMove(src, usr, common.South)
}

func (d *ClientData) ProcEast(src *types.Datum, usr *types.Datum) types.Value {
	return d.relMove(src, usr, common.East)
}

func (d *ClientData) ProcWest(src *types.Datum, usr *types.Datum) types.Value {
	return d.relMove(src, usr, common.West)
}

func (d *ClientData) ProcMove(src *types.Datum, usr *types.Datum, loc types.Value, dir types.Value) types.Value {
	mob := d.GetMob(src)
	util.FIXME("cancel automated movement if necessary")
	if mob != nil {
		return mob.Invoke(usr, "Move", loc, dir)
	}
	return types.Int(0)
}

func (d *ClientData) OperatorWrite(src *types.Datum, usr *types.Datum, output types.Value) types.Value {
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

func (d *ClientData) ProcStat(src *types.Datum, usr *types.Datum) types.Value {
	statobj := d.VarStatobj.Dereference()
	if statobj != nil {
		statobj.Invoke(usr, "Stat")
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

func FlickClient(client *types.Datum, flick sprite.Flick) {
	_, d := ClientDataChunk(client)
	d.flicks = append(d.flicks, flick)
}

func PullClientRequests(client *types.Datum) (textDisplay []string, sounds []sprite.Sound, flicks []sprite.Flick) {
	_, d := ClientDataChunk(client)
	textDisplay, sounds, flicks = d.textBuffer, d.soundBuffer, d.flicks
	d.textBuffer = nil
	d.soundBuffer = nil
	d.flicks = nil
	return textDisplay, sounds, flicks
}

func (w *World) RenderClientView(client types.Value) (center types.Value, viewAtoms []types.Value, stat sprite.StatDisplay, verbs []string, verbsOn map[*types.Datum][]string) {
	cdatum, cc := ClientDataChunk(client)
	util.FIXME("actually do this correctly")
	eye := client.Var("eye").(*types.Datum)
	veye := client.Var("virtual_eye").(*types.Datum)
	view := types.Unuint(client.Var("view"))
	verbs, verbsOn = cc.ListVerbs(cdatum)
	return veye, w.ViewX(view, veye, eye, atoms.ViewVisual), cc.statDisplay, verbs, verbsOn
}

func (w *World) constructNewMob(key string) types.Value {
	mob := w.Realm().New(w.Mob, nil)
	if !types.IsType(mob, "/mob") {
		panic("constructed mob is not a /mob")
	}
	if key == "" {
		panic("not sure what to do when key is an empty string")
	}
	mob.SetVar("name", types.String(key))
	util.FIXME("initialize default gender")
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

func (d *ClientData) ProcNew(src *types.Datum, _ *types.Datum, usr types.Value) types.Value {
	mob := usr
	util.FIXME("add support for 'prototype mobs'")
	if mob == nil {
		mob = atoms.WorldOf(src).(*World).constructNewMob(d.VarKey)
	}
	util.NiceToHave("add support for Topics")
	d.SetMob(src, mob)
	mob.Invoke(mob.(*types.Datum), "Login")
	return mob
}

func (d *ClientData) ProcDel(src *types.Datum, usr *types.Datum) types.Value {
	util.FIXME("call Logout() on mob")
	util.FIXME("should killing the connection go here, maybe in addition to other places?")
	return nil
}

func (d *ClientData) ResolveVerb(src *types.Datum, verbName string, args []string) {
	util.FIXME("autocomplete user's input, following the 'space autocompletes' rule, rather than just letting the player type")
	mob := src.Var("mob")
	if mob == nil {
		util.FIXME("see if there are cases where verbs can be executed without a mob")
		// cannot execute verbs without a mob
		return
	}
	verbUsr := mob.(*types.Datum)
	for _, verbSrcVal := range atoms.WorldOf(src).FindAllType("/atom") {
		verbSrc := verbSrcVal.(*types.Datum)
		for _, verbVal := range datum.Elements(verbSrc.Var("verbs")) {
			verb := verbVal.(atoms.Verb)
			if verb.Matches(verbName, verbSrc, verbUsr, args) {
				resolved, err := verb.ResolveArgs(verbSrc, verbUsr, args)
				if err != nil {
					log.Printf("cannot resolve verb %v: %v\n", verb, err)
					src.Invoke(verbUsr, "<<", types.String(fmt.Sprintf("cannot resolve verb %v: %v\n", verb, err)))
				} else {
					verb.Apply(verbSrc, verbUsr, resolved)
				}
				return
			}
		}
	}
	log.Println("got unknown verb:", verbName)
	src.Invoke(verbUsr, "<<", types.String(fmt.Sprintf("Not a known verb: %q", verbName)))
}

func (d *ClientData) listVerbsOnAtomInternal(src *types.Datum, usr *types.Datum, atom *types.Datum) (verbs []string) {
	for _, verbVal := range datum.Elements(atom.Var("verbs")) {
		verb := verbVal.(atoms.Verb)
		if verb.Matches(verb.VisibleName, atom, usr, nil) {
			verbs = append(verbs, verb.VisibleName)
		}
	}
	return verbs
}

func (w *World) ListVerbsOnAtom(client types.Value, atom *types.Datum) (verbs []string) {
	cdatum, cd := ClientDataChunk(client)
	mob := client.Var("mob")
	if mob == nil {
		util.FIXME("see if there are cases where verbs can be executed without a mob")
		// cannot execute verbs without a mob
		return nil
	}
	return cd.listVerbsOnAtomInternal(cdatum, mob.(*types.Datum), atom)
}

func (d *ClientData) ListVerbs(src *types.Datum) (verbs []string, available map[*types.Datum][]string) {
	mob := src.Var("mob")
	if mob == nil {
		util.FIXME("see if there are cases where verbs can be executed without a mob")
		// cannot execute verbs without a mob
		return
	}
	verbUsr := mob.(*types.Datum)
	allVerbs := map[string]struct{}{}
	available = map[*types.Datum][]string{}
	for _, verbSrc := range atoms.WorldOf(src).FindAllType("/atom") {
		verbsOnAtom := d.listVerbsOnAtomInternal(src, verbUsr, verbSrc.(*types.Datum))
		for _, verb := range verbsOnAtom {
			allVerbs[verb] = struct{}{}
		}
		if verbSrc != mob && len(verbsOnAtom) > 0 {
			available[verbSrc.(*types.Datum)] = verbsOnAtom
		}
	}
	for verb := range allVerbs {
		verbs = append(verbs, verb)
	}
	sort.Strings(verbs)
	return verbs, available
}
