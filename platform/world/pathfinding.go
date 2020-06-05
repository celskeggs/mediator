package world

import (
	"github.com/celskeggs/mediator/common"
	"github.com/celskeggs/mediator/pathfinding"
	"github.com/celskeggs/mediator/platform/atoms"
	"github.com/celskeggs/mediator/platform/types"
	"github.com/celskeggs/mediator/util"
)

// not intended for public use
func UpdateWalk(src types.Value) {
	ws := atoms.GetWalkState(src)
	if ws.WalkTarget != nil {
		ws.WalkCountdown--
		if ws.WalkCountdown <= 0 {
			ws.WalkCountdown = ws.WalkLagTicks
			atoms.SetWalkState(src, ws)
			StepTo(nil, src, ws.WalkTarget, ws.WalkMinimum)
			// can't push SetWalkState later, because what if StepTo dispatches to a Move that calls walk(...)?
		} else {
			atoms.SetWalkState(src, ws)
		}
	}
}

func WalkTo(ref, target types.Value, min, lag int) {
	util.FIXME("add 'speed' parameter to walk_to")
	if lag < 1 {
		lag = 1
	}
	util.FIXME("edge cases for min?")
	atoms.SetWalkState(ref, atoms.WalkState{
		WalkLagTicks:  lag,
		WalkCountdown: lag,
		WalkMinimum:   min,
		WalkTarget:    target,
	})
}

func StepTo(usr *types.Datum, ref, target types.Value, min int) bool {
	util.FIXME("add 'speed' parameter to step_to")

	newLocation := GetStepTo(ref, target, min)
	return newLocation != nil && types.AsBool(ref.Invoke(usr, "Move", newLocation, GetDir(ref, newLocation)))
}

func GetStepTo(ref, target types.Value, min int) types.Value {
	util.FIXME("make this actually follow the real algorithm")

	sx, sy, z := XYZ(ref)
	tx, ty, tz := XYZ(target)

	if z != tz {
		return nil
	}

	w := atoms.WorldOf(ref.(*types.Datum))

	path := pathfinding.Search(
		pathfinding.Point{X: sx, Y: sy},
		pathfinding.Point{X: tx, Y: ty},
		func(p pathfinding.Point) bool {
			turf := w.LocateXYZ(p.X, p.Y, z)
			return turf != nil && !types.AsBool(turf.Var("density"))
		},
	)

	if len(path) > min+1 {
		return w.LocateXYZ(path[min].X, path[min].Y, z)
	} else {
		return nil
	}
}

func GetDir(from, to types.Value) common.Direction {
	util.FIXME("should we do anything with the Z direction?")
	x1, y1 := XY(from)
	x2, y2 := XY(to)
	return common.GetDir(x1, y1, x2, y2)
}
