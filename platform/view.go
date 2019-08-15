package platform

import (
	"github.com/celskeggs/mediator/common"
	"github.com/celskeggs/mediator/platform/datum"
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

func ManhattanDistance(a, b IAtom) uint {
	ax, ay, _ := a.XYZ()
	bx, by, _ := b.XYZ()
	return MaxUint(AbsDiff(ax, bx), AbsDiff(ay, by))
}

func (w *World) View1(center datum.IDatum) []IAtom {
	return w.View(w.ViewDistance, center)
}

func expandWithContents(atoms []IAtom) (out []IAtom) {
	out = append([]IAtom{}, atoms...)
	for _, atom := range atoms {
		out = append(out, atom.Contents()...)
	}
	return out
}

// note: this does not handle the "centerD = nil" case the same as DreamMaker
func (w *World) View(distance uint, centerD datum.IDatum) []IAtom {
	client, isclient := centerD.(IClient)
	var center IAtom
	if isclient {
		center = client.Eye()
	} else if centerD != nil {
		center = centerD.(IAtom)
	}
	return w.ViewX(distance, center, center)
}

func (w *World) ViewX(distance uint, center IAtom, perspective IAtom) []IAtom {
	if center == nil || perspective == nil {
		return nil
	}

	util.FIXME("include areas")

	location := perspective.Location()
	turfloc, isturf := location.(ITurf)
	if isturf {
		_, _, tz := turfloc.XYZ()
		atoms := w.FindAll(func(atom IAtom) bool {
			turf, isturf := atom.(ITurf)
			if isturf {
				_, _, t2z := turf.XYZ()
				return t2z == tz && ManhattanDistance(turf, center) <= distance
			}
			return false
		})
		turfs := make([]ITurf, len(atoms))
		for i, turf := range atoms {
			turfs[i] = turf.(ITurf)
		}
		nturfs := limitViewers(distance, center, perspective, turfs)
		atomsAgain := make([]IAtom, len(nturfs)+1)
		for i, turf := range nturfs {
			atomsAgain[i] = turf
		}
		atomsAgain[len(nturfs)] = perspective
		return expandWithContents(atomsAgain)
	} else if location != nil {
		return expandWithContents([]IAtom{
			location, perspective,
		})
	} else {
		return expandWithContents([]IAtom{
			perspective,
		})
	}
}

type viewInfo struct {
	Opaque     bool
	Luminosity uint
	Lit        bool
	Turf       ITurf
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
		println("parameters:", x, y, vir.CornerX, vir.CornerY, vir.Distance)
		panic("InfoAt parameters out of range")
	}
	if y < vir.CornerY || y >= vir.CornerY+len(vir.Info[rx]) {
		panic("InfoAt parameters out of range")
	}
	return uint(rx), uint(ry)
}

func (vir *viewInfoRegion) PopulateTurfs(input []ITurf) (maxDepthMax, sumDepthMax int) {
	for _, turf := range input {
		tx, ty, _ := turf.XYZ()
		ox, oy := vir.XYToOffset(tx, ty)
		if vir.Info[ox][oy] != nil {
			panic("duplicate turfs for position")
		}
		dx, dy := AbsDiff(vir.PerspectiveX, tx), AbsDiff(vir.PerspectiveY, ty)
		vi := &viewInfo{
			Opaque:     turf.AsAtom().Opacity,
			Luminosity: 0,
			Lit:        true,
			Turf:       turf,
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
func limitViewers(distance uint, center IAtom, perspective IAtom, base []ITurf) []ITurf {
	centerX, centerY, _ := center.XYZ()
	perspectiveX, perspectiveY, _ := perspective.XYZ()
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
				tx, ty, _ := info.Turf.XYZ()
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
				tx, ty, _ := info.Turf.XYZ()
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
				tx, ty, _ := info.Turf.XYZ()
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
				println("cancel out icon due to luminosity")
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

				txu, tyu, _ := info.Turf.XYZ()
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

	var finalTurfs []ITurf

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
