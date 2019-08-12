package worldmap

import (
	"errors"
	"fmt"
	"github.com/celskeggs/mediator/parsemap"
	"github.com/celskeggs/mediator/platform"
	"github.com/celskeggs/mediator/platform/datum"
	"github.com/celskeggs/mediator/util"
	"io/ioutil"
)

type mapCell struct {
	Area     platform.IArea
	Turf     platform.ITurf
	Contents []platform.IAtom
}

func (cell *mapCell) Add(atom platform.IAtom) error {
	datum.AssertConsistent(atom)
	if area, isarea := atom.(platform.IArea); isarea {
		if cell.Area != nil {
			return errors.New("more than one area in cell")
		}
		cell.Area = area
	} else if turf, isturf := atom.(platform.ITurf); isturf {
		if cell.Turf != nil {
			return errors.New("more than one turf in cell")
		}
		cell.Turf = turf
	} else {
		cell.Contents = append(cell.Contents, atom)
	}
	return nil
}

func (cell *mapCell) Stitch(x, y, z uint) {
	util.FIXME("populate areas and turfs if not provided")
	if cell.Turf != nil {
		cell.Turf.SetXYZ(x, y, z)
	}
	if cell.Area != nil && cell.Turf != nil {
		cell.Turf.SetLocation(cell.Area)
	}
	if cell.Turf != nil {
		for _, atom := range cell.Contents {
			atom.SetLocation(cell.Turf)
		}
	}
}

type worldMap [][][]mapCell

type loaderObserver struct {
	world *platform.World
	m     worldMap
}

func (lo *loaderObserver) SetSize(l parsemap.Location) {
	lo.m = make([][][]mapCell, l.X)
	for x := uint32(0); x < l.X; x++ {
		lo.m[x] = make([][]mapCell, l.Y)
		for y := uint32(0); y < l.Y; y++ {
			lo.m[x][y] = make([]mapCell, l.Z)
			for z := uint32(0); z < l.Z; z++ {
				lo.m[x][z][z] = mapCell{}
			}
		}
	}
}

func (lo *loaderObserver) AddAtom(l parsemap.Location, path string) error {
	if !lo.world.Tree.Exists(datum.TypePath(path)) {
		return fmt.Errorf("no such type path: %s", path)
	}
	atom, ok := lo.world.Tree.New(datum.TypePath(path)).(platform.IAtom)
	if !ok {
		panic("expected type path " + path + " specified in AddAtom to be an atom")
	}
	return lo.m[l.X][l.Y][l.Z].Add(atom)
}

// sets up all of the object locations
func (lo *loaderObserver) StitchMap() {
	util.FIXME("make singleton areas, not per-cell areas or even per-area areas")
	for x := 0; x < len(lo.m); x++ {
		for y := 0; y < len(lo.m[x]); y++ {
			for z := 0; z < len(lo.m[x][y]); z++ {
				cell := lo.m[x][y][z]
				cell.Stitch(uint(x+1), uint(y+1), uint(z+1))
			}
		}
	}
}

func LoadMap(world *platform.World, text string) error {
	l := loaderObserver{
		world: world,
	}
	err := parsemap.ProduceMap(text, &l)
	if err != nil {
		return err
	}
	l.StitchMap()
	util.NiceToHave("handle changing this both at compile time and at runtime")
	world.MaxX = uint(len(l.m))
	world.MaxY = uint(len(l.m[0]))
	world.MaxZ = uint(len(l.m[0][0]))
	return nil
}

func LoadMapFromFile(world *platform.World, filename string) error {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	return LoadMap(world, string(content))
}
