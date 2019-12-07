package atoms

import (
	"github.com/celskeggs/mediator/platform/icon"
	"github.com/celskeggs/mediator/platform/types"
)

type World interface {
	PlayerExists(client types.Value) bool
	MaxXYZ() (uint, uint, uint)
	SetMaxXYZ(x, y, z uint)
	LocateXYZ(x, y, z uint) (turf types.Value)
	Realm() *types.Realm
	Icon(name string) *icon.Icon
}

func WorldOf(t *types.Datum) World {
	return t.Realm().WorldRef().(World)
}
