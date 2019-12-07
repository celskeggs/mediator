package atoms

import "github.com/celskeggs/mediator/platform/types"

type World interface {
	PlayerExists(client types.Value) bool
	MaxXYZ() (uint, uint, uint)
	SetMaxXYZ(x, y, z uint)
	LocateXYZ(x, y, z uint) (turf *types.Datum)
	Realm() *types.Realm
}

func WorldOf(t *types.Datum) World {
	return t.Realm().WorldRef().(World)
}
