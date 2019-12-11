package worldmap

import (
	"errors"
	"github.com/celskeggs/mediator/parsemap"
	"github.com/celskeggs/mediator/platform/atoms"
	"github.com/celskeggs/mediator/platform/types"
	"github.com/celskeggs/mediator/resourcepack"
	"github.com/celskeggs/mediator/util"
	"io/ioutil"
)

type mapCell struct {
	Area     *types.Datum
	Turf     *types.Datum
	Contents []*types.Datum
}

func (cell *mapCell) Add(atom *types.Datum) error {
	if types.IsType(atom, "/area") {
		if cell.Area != nil {
			return errors.New("more than one area in cell")
		}
		cell.Area = atom
	} else if types.IsType(atom, "/turf") {
		if cell.Turf != nil {
			return errors.New("more than one turf in cell")
		}
		cell.Turf = atom
	} else {
		types.AssertType(atom, "/atom")
		cell.Contents = append(cell.Contents, atom)
	}
	return nil
}

func (cell *mapCell) Stitch(x, y, z uint) {
	util.FIXME("populate areas and turfs if not provided")
	if cell.Turf != nil {
		cell.Turf.SetVar("x", types.Int(x))
		cell.Turf.SetVar("y", types.Int(y))
		cell.Turf.SetVar("z", types.Int(z))
	}
	if cell.Area != nil && cell.Turf != nil {
		cell.Turf.SetVar("loc", cell.Area)
	}
	if cell.Turf != nil {
		for _, item := range cell.Contents {
			item.SetVar("loc", cell.Turf)
		}
	}
}

type worldMap [][][]mapCell

type loaderObserver struct {
	world atoms.World
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
	atom := lo.world.Realm().New(types.TypePath(path))
	return lo.m[l.X][l.Y][l.Z].Add(atom)
}

// sets up all of the object locations
func (lo *loaderObserver) StitchMap() {
	for x := 0; x < len(lo.m); x++ {
		for y := 0; y < len(lo.m[x]); y++ {
			for z := 0; z < len(lo.m[x][y]); z++ {
				cell := lo.m[x][y][z]
				cell.Stitch(uint(x+1), uint(y+1), uint(z+1))
			}
		}
	}
}

func LoadMap(world atoms.World, text string) error {
	l := loaderObserver{
		world: world,
	}
	err := parsemap.ProduceMap(text, &l)
	if err != nil {
		return err
	}
	l.StitchMap()
	util.NiceToHave("handle changing this both at compile time and at runtime")
	world.SetMaxXYZ(uint(len(l.m)), uint(len(l.m[0])), uint(len(l.m[0][0])))
	return nil
}

func LoadMapFromFile(world atoms.World, filename string) error {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	return LoadMap(world, string(content))
}

func LoadMapFromPack(world atoms.World, pack *resourcepack.ResourcePack, name string) error {
	content, err := pack.Resource(name)
	if err != nil {
		return err
	}
	return LoadMap(world, string(content.Data))
}
