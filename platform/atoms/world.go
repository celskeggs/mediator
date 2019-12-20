package atoms

import (
	"github.com/celskeggs/mediator/platform/icon"
	"github.com/celskeggs/mediator/platform/types"
)

type ViewMode uint

const (
	ViewInclusive ViewMode = iota // corresponds to view proc
	ViewExclusive                 // corresponds to oview proc
	ViewVisual                    // actual visual drawn elements; like 'view' but excludes contents of player
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
	View(distance uint, centerD *types.Datum, mode ViewMode) []types.Value
	View1(centerD *types.Datum, mode ViewMode) []types.Value
	// if there's a Stat() operation in progress, this stores a reference to the context so that we can update the panel
	StatContext() StatContext
}

func WorldOf(t *types.Datum) World {
	return t.Realm().WorldRef().(World)
}

type StatContext interface {
	Stat(name string, value types.Value)
	StatPanel(panel string) bool
}
