package world

import (
	"github.com/celskeggs/mediator/common"
	"github.com/celskeggs/mediator/platform/atoms"
	"github.com/celskeggs/mediator/platform/datum"
	"github.com/celskeggs/mediator/platform/types"
	"github.com/celskeggs/mediator/util"
)

func MaxUint(a uint, b uint) uint {
	if a > b {
		return a
	} else {
		return b
	}
}

func AbsDiff(a uint, b uint) uint {
	if a > b {
		return a - b
	} else {
		return b - a
	}
}

func XY(atom types.Value) (uint, uint) {
	return types.Unuint(atom.Var("x")), types.Unuint(atom.Var("y"))
}

func XYZ(atom types.Value) (uint, uint, uint) {
	return types.Unuint(atom.Var("x")), types.Unuint(atom.Var("y")), types.Unuint(atom.Var("z"))
}

func ManhattanDistance(a, b types.Value) uint {
	ax, ay := XY(a)
	bx, by := XY(b)
	return MaxUint(AbsDiff(ax, bx), AbsDiff(ay, by))
}

func (w *World) View1(center *types.Datum, mode atoms.ViewMode) []types.Value {
	return w.View(w.ViewDist, center, mode)
}

func expandWithContents(atoms []types.Value) (out []types.Value) {
	out = append([]types.Value{}, atoms...)
	for _, atom := range atoms {
		out = append(out, datum.Elements(atom.Var("contents"))...)
	}
	return out
}

func atomsExcept(atoms []types.Value, except types.Value) (out []types.Value) {
	for _, atom := range atoms {
		if atom != except {
			out = append(out, atom)
		}
	}
	return out
}

func containsAtom(atoms []types.Value, cmp types.Value) bool {
	for _, atom := range atoms {
		if atom == cmp {
			return true
		}
	}
	return false
}

// note: this does not handle the "centerD = nil" case the same as DreamMaker
func (w *World) View(distance uint, centerD *types.Datum, mode atoms.ViewMode) []types.Value {
	var center *types.Datum
	if centerD != nil {
		if types.IsType(centerD, "/client") {
			center = centerD.Var("eye").(*types.Datum)
		} else if types.IsType(centerD, "/atom") {
			center = centerD
		} else {
			panic("view center is not an /atom")
		}
	}
	return w.ViewX(distance, center, center, mode)
}

func (w *World) ViewX(distance uint, center *types.Datum, perspective *types.Datum, mode atoms.ViewMode) []types.Value {
	locations := w.ViewXLocations(distance, center, perspective)
	switch mode {
	case atoms.ViewInclusive:
		contents := expandWithContents(locations)
		// make sure that 'perspective' and its contents are in the list of found objects
		// but also that perspective is not in the output more than once
		if perspective != nil {
			if !containsAtom(contents, perspective) {
				contents = append(contents, perspective)
			}
			contents = append(contents, datum.Elements(perspective.Var("contents"))...)
		}
		return contents
	case atoms.ViewVisual:
		// include 'perspective' in the list of found objects, but not its contents
		contents := expandWithContents(locations)
		if !containsAtom(contents, perspective) {
			contents = append(contents, perspective)
		}
		return contents
	case atoms.ViewExclusive:
		// make sure 'perspective' does not get added to the list of contents
		contents := expandWithContents(locations)
		contents = atomsExcept(contents, perspective)
		return contents
	default:
		panic("unknown view type " + string(uint(mode)))
	}
}

func (w *World) ViewXLocations(distance uint, center *types.Datum, perspective *types.Datum) []types.Value {
	if center == nil || perspective == nil {
		return nil
	}

	util.FIXME("include areas")

	var location types.Value = perspective
	if !types.IsType(location, "/turf") {
		location = perspective.Var("loc")
	}
	if types.IsType(location, "/turf") {
		tz := types.Unuint(location.Var("z"))
		turfs := w.FindAll(func(turf *types.Datum) bool {
			if types.IsType(turf, "/turf") {
				t2z := types.Unuint(turf.Var("z"))
				return t2z == tz && ManhattanDistance(turf, center) <= distance
			}
			return false
		})
		return limitViewers(distance, center, perspective, turfs)
	} else if location != nil {
		return []types.Value{
			location,
		}
	} else {
		return nil
	}
}

type viewInfo struct {
	Opaque     bool
	Luminosity uint
	Lit        bool
	Turf       *types.Datum
	MaxXY      int
	SumXY      int
	Vis        int
	Vis2       int
	Noted      bool
}

type viewInfoRegion struct {
	Info                       [][]*viewInfo
	CornerX, CornerY           int
	PerspectiveX, PerspectiveY uint
	Distance                   uint
}

func newViewInfoRegion(distance uint, centerX, centerY, perspectiveX, perspectiveY uint) viewInfoRegion {
	vir := viewInfoRegion{
		Info:         make([][]*viewInfo, distance*2+1),
		CornerX:      int(centerX) - int(distance),
		CornerY:      int(centerY) - int(distance),
		PerspectiveX: perspectiveX,
		PerspectiveY: perspectiveY,
		Distance:     distance,
	}
	for i := uint(0); i < distance*2+1; i++ {
		vir.Info[i] = make([]*viewInfo, distance*2+1)
	}
	return vir
}

func (vir *viewInfoRegion) InfoAt(x, y int) *viewInfo {
	rx, ry := x-vir.CornerX, y-vir.CornerY
	if x < vir.CornerX || x >= vir.CornerX+len(vir.Info) {
		return nil
	}
	if y < vir.CornerY || y >= vir.CornerY+len(vir.Info[rx]) {
		return nil
	}
	return vir.Info[rx][ry]
}

func (vir *viewInfoRegion) XYToOffset(xu, yu uint) (lx, ly uint) {
	x, y := int(xu), int(yu)
	rx, ry := x-vir.CornerX, y-vir.CornerY
	if x < vir.CornerX || x >= vir.CornerX+len(vir.Info) {
		panic("InfoAt parameters out of range")
	}
	if y < vir.CornerY || y >= vir.CornerY+len(vir.Info[rx]) {
		panic("InfoAt parameters out of range")
	}
	return uint(rx), uint(ry)
}

func (vir *viewInfoRegion) PopulateTurfs(input []types.Value) (maxDepthMax, sumDepthMax int) {
	for _, turf := range input {
		tx, ty := XY(turf)
		ox, oy := vir.XYToOffset(tx, ty)
		if vir.Info[ox][oy] != nil {
			panic("duplicate turfs for position")
		}
		dx, dy := AbsDiff(vir.PerspectiveX, tx), AbsDiff(vir.PerspectiveY, ty)
		vi := &viewInfo{
			Opaque:     types.AsBool(turf.Var("opacity")),
			Luminosity: 0,
			Lit:        true,
			Turf:       turf.(*types.Datum),
			MaxXY:      int(MaxUint(dx, dy)),
			SumXY:      int(dx + dy),
			Vis:        0,
			Vis2:       0,
		}
		vir.Info[ox][oy] = vi
		util.NiceToHave("infrared vision?")

		if vi.MaxXY > maxDepthMax {
			maxDepthMax = vi.MaxXY
		}
		if vi.SumXY > sumDepthMax {
			sumDepthMax = vi.SumXY
		}
	}
	// we let anything else just not get populated, which means Opaque=false, Lit=false, Luminosity=0
	return maxDepthMax, sumDepthMax
}

// this is an approximate reimplementation of the BYOND algorithm, based on http://www.byond.com/forum/post/2130277#comment20659267
func limitViewers(distance uint, center *types.Datum, perspective *types.Datum, base []types.Value) []types.Value {
	centerX, centerY := XY(center)
	perspectiveX, perspectiveY := XY(perspective)
	vir := newViewInfoRegion(distance, centerX, centerY, perspectiveX, perspectiveY)
	maxDepthMax, sumDepthMax := vir.PopulateTurfs(base)

	util.NiceToHave("handle blindness")
	util.NiceToHave("there's something here related to having everything be visible in some circumstances?")

	// diagonal shadow loop
	for d := 0; d < maxDepthMax; d++ {
		for _, infos := range vir.Info {
			for _, info := range infos {
				if info == nil {
					continue
				}
				tx, ty := XY(info.Turf)
				if info.MaxXY == d+1 {
					for _, neighborDir := range common.EightDirections {
						dx, dy := neighborDir.XY()
						neighbor := vir.InfoAt(int(tx)+dx, int(ty)+dy)
						if neighbor != nil && neighbor.Vis2 == d {
							if info.Opaque {
								info.Vis2 = -1
							} else {
								info.Vis2 = d + 1
							}
							break
						}
					}
				}
			}
		}
	}

	// straight shadow loop
	for d := 0; d < sumDepthMax; d++ {
		for _, infos := range vir.Info {
			for _, info := range infos {
				if info == nil {
					continue
				}
				tx, ty := XY(info.Turf)
				if info.SumXY == d+1 {
					for _, neighborDir := range common.EightDirections {
						dx, dy := neighborDir.XY()
						neighbor := vir.InfoAt(int(tx)+dx, int(ty)+dy)
						if neighbor != nil && neighbor.Vis == d {
							if info.Opaque {
								info.Vis = -1
							} else if info.Vis2 != 0 {
								info.Vis = d + 1
							}
							break
						}
					}
				}
			}
		}
	}

	vir.InfoAt(int(vir.PerspectiveX), int(vir.PerspectiveY)).Vis = 1

	updatedLighting := true
	for updatedLighting {
		updatedLighting = false
		for _, infos := range vir.Info {
			for _, info := range infos {
				if info == nil || info.Luminosity == 0 {
					continue
				}
				tx, ty := XY(info.Turf)
				for _, neighborDir := range common.EightDirections {
					dx, dy := neighborDir.XY()
					neighbor := vir.InfoAt(int(tx)+dx, int(ty)+dy)
					if neighbor == nil {
						// nothing
					} else if neighbor.Opaque {
						neighbor.Luminosity = 1
					} else if neighbor.Luminosity < info.Luminosity-1 {
						neighbor.Luminosity = info.Luminosity - 1
						updatedLighting = true
					}
				}
			}
		}
	}

	util.NiceToHave("infrared sight handling: step 7")

	for _, infos := range vir.Info {
		for _, info := range infos {
			if info == nil {
				continue
			}
			info.Vis2 = info.Vis
			if info.Luminosity == 0 && !info.Lit {
				info.Vis = 0
			}
		}
	}

	for _, infos := range vir.Info {
		for _, info := range infos {
			if info == nil {
				continue
			}
			if info.Vis == 0 && info.Opaque {
				makeANote := false

				txu, tyu := XY(info.Turf)
				tx, ty := int(txu), int(tyu)
				east, west := vir.InfoAt(tx+1, ty), vir.InfoAt(tx-1, ty)
				north, south := vir.InfoAt(tx, ty+1), vir.InfoAt(tx, ty-1)
				if (east != nil && west != nil && east.Vis != 0 && west.Vis != 0) ||
					(north != nil && south != nil && north.Vis != 0 && south.Vis != 0) {
					makeANote = true
				}
				ne, nw := vir.InfoAt(tx+1, ty+1), vir.InfoAt(tx-1, ty+1)
				se, sw := vir.InfoAt(tx+1, ty-1), vir.InfoAt(tx-1, ty-1)
				if ne != nil && ne.Vis != 0 && east.Vis != 0 && north.Vis != 0 && east.Opaque && north.Opaque && !ne.Opaque {
					makeANote = true
				}
				if nw != nil && nw.Vis != 0 && west.Vis != 0 && north.Vis != 0 && west.Opaque && north.Opaque && !nw.Opaque {
					makeANote = true
				}
				if se != nil && se.Vis != 0 && east.Vis != 0 && south.Vis != 0 && east.Opaque && south.Opaque && !se.Opaque {
					makeANote = true
				}
				if sw != nil && sw.Vis != 0 && west.Vis != 0 && south.Vis != 0 && west.Opaque && south.Opaque && !sw.Opaque {
					makeANote = true
				}

				info.Noted = makeANote
			}
		}
	}

	for _, infos := range vir.Info {
		for _, info := range infos {
			if info == nil {
				continue
			}
			if info.Noted {
				info.Vis = -1
			}
		}
	}

	// at this point, if vis2 != 0, then we have line of sight visibility but not necessarily anything else

	var finalTurfs []types.Value

	for _, infos := range vir.Info {
		for _, info := range infos {
			if info == nil {
				continue
			}
			if info.Vis != 0 {
				finalTurfs = append(finalTurfs, info.Turf)
			}
		}
	}
	return finalTurfs
}
