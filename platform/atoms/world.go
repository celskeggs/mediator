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
	FindAll(predicate func(*types.Datum) bool) []types.Value
	FindAllType(tp types.TypePath) []types.Value
	FindOne(predicate func(*types.Datum) bool) types.Value
	FindOneType(tp types.TypePath) types.Value
	View(distance uint, centerD *types.Datum, oview bool) []types.Value
	View1(centerD *types.Datum, oview bool) []types.Value
}

func WorldOf(t *types.Datum) World {
	return t.Realm().WorldRef().(World)
}
