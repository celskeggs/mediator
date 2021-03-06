package world

import (
	"github.com/celskeggs/mediator/common"
	"github.com/celskeggs/mediator/platform/atoms"
	"github.com/celskeggs/mediator/platform/icon"
	"github.com/celskeggs/mediator/platform/types"
	"github.com/celskeggs/mediator/util"
	"github.com/celskeggs/mediator/webclient/sprite"
	"github.com/celskeggs/mediator/websession"
)

type World struct {
	Name     string
	Mob      types.TypePath
	ViewDist uint

	defaultLazyEye uint

	MaxX, MaxY, MaxZ uint

	realm     *types.Realm
	iconCache *icon.IconCache

	clients map[*types.Datum]*types.Ref

	// true if this instance has an API associated with it
	// we never provide more than one API so that we avoid double-threading
	claimed bool

	// true if the virtual eye should be set to the middle of the may
	setVirtualEye bool
}

var _ atoms.World = &World{}

func (w World) MaxXYZ() (uint, uint, uint) {
	return w.MaxX, w.MaxY, w.MaxZ
}

func (w *World) SetMaxXYZ(x, y, z uint) {
	if w.MaxX != 0 || w.MaxY != 0 || w.MaxZ != 0 {
		panic("world map's size was already populated")
	}
	w.MaxX, w.MaxY, w.MaxZ = x, y, z
}

func (w *World) Realm() *types.Realm {
	return w.realm
}

func (w *World) FindAll(predicate func(*types.Datum) bool) []types.Value {
	return w.Realm().FindAll(func(d *types.Datum) bool {
		return types.IsType(d, "/atom") && (predicate == nil || predicate(d))
	})
}

func (w *World) FindAllType(tp types.TypePath) []types.Value {
	return w.Realm().FindAll(func(datum *types.Datum) bool {
		return types.IsType(datum, tp)
	})
}

func (w *World) FindOne(predicate func(*types.Datum) bool) types.Value {
	return w.Realm().FindOne(func(datum *types.Datum) bool {
		return types.IsType(datum, "/atom") && predicate(datum)
	})
}

func (w *World) FindOneType(tp types.TypePath) types.Value {
	return w.Realm().FindOne(func(datum *types.Datum) bool {
		return types.IsType(datum, tp)
	})
}

func (w *World) CreateNewPlayer(key string) *types.Datum {
	client := w.realm.NewPlain("/client")
	client.SetVar("key", types.String(key))
	if types.Unint(client.Var("view")) == 0 {
		util.NiceToHave("handle the /client/view = 0 situation correctly")
		client.SetVar("view", types.Int(w.ViewDist))
	}
	w.clients[client] = types.Reference(client)
	client.Invoke(nil, "New", w.findExistingMob(key))
	return client
}

func (w *World) RemovePlayer(client *types.Datum) {
	delete(w.clients, client)
	client.Invoke(nil, "Del")
}

func (w *World) PlayerExists(client types.Value) bool {
	_, found := w.clients[client.(*types.Datum)]
	return found
}

func (w *World) ServerAPI() websession.WorldAPI {
	if w.claimed {
		panic("second ServerAPI() call unexpected")
	}
	w.claimed = true
	return &worldAPI{
		World:   w,
		updates: make(chan struct{}, 1),
	}
}

func (w *World) LocateXYZ(x, y, z uint) types.Value {
	util.FIXME("this can definitely be more efficient")
	return w.FindOne(func(atom *types.Datum) bool {
		if types.IsType(atom, "/turf") {
			tx, ty, tz := XYZ(atom)
			if tx == x && ty == y && tz == z {
				return true
			}
		}
		return false
	})
}

func (w *World) UpdateDefaultViewDistance() {
	// if the map is <= 21x21, adjust view to fit the whole thing
	if w.MaxX > 0 && w.MaxY > 0 && w.MaxX <= 21 && w.MaxY <= 21 {
		w.ViewDist = MaxUint(w.MaxX, w.MaxY) / 2
		// note: the documentation SAYS that we should turn on lazy_eye, but it actually doesn't get turned on.
		w.setVirtualEye = true
	}
}

func (w *World) Icon(name string) *icon.Icon {
	return w.iconCache.LoadOrPanic(name)
}

func (w *World) Flick(icon *icon.Icon, iconState string, target types.Value) {
	appearance := target.Var("appearance").(atoms.Appearance)
	appearance.Icon = icon
	appearance.IconState = iconState
	ok, _, s := appearance.ToSprite(0, 0, target.Var("dir").(common.Direction))
	if !ok {
		panic("failed to convert appearance to sprite while generating flick")
	}
	flick := sprite.Flick{
		Icon:         s.Icon,
		Frames:       s.Frames,
		SourceWidth:  s.SourceWidth,
		SourceHeight: s.SourceHeight,
		UID:          target.(*types.Datum).UID(),
	}
	for _, client := range w.FindAllType("/client") {
		FlickClient(client.(*types.Datum), flick)
	}
}

func NewWorld(realm *types.Realm, cache *icon.IconCache) *World {
	world := &World{
		Name:          "Untitled",
		Mob:           "/mob",
		ViewDist:      5,
		realm:         realm,
		iconCache:     cache,
		clients:       map[*types.Datum]*types.Ref{},
		claimed:       false,
		setVirtualEye: false,
	}
	realm.SetWorldRef(world)
	return world
}
