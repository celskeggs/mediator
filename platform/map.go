package platform

import (
	"io/ioutil"
	"fmt"
	"github.com/celskeggs/mediator/platform/parsemap"
)

type MapCell []*Atom
type Map [][][]MapCell

type loaderObserver struct {
	tree *Tree
	m    Map
}

func (lo *loaderObserver) SetSize(l parsemap.Location) {
	lo.m = make([][][]MapCell, l.X)
	for x := uint32(0); x < l.X; x++ {
		lo.m[x] = make([][]MapCell, l.Y)
		for y := uint32(0); y < l.Y; y++ {
			lo.m[x][y] = make([]MapCell, l.Z)
			for z := uint32(0); z < l.Z; z++ {
				lo.m[x][z][z] = MapCell{}
			}
		}
	}
}

func (lo *loaderObserver) AddAtom(l parsemap.Location, path string) error {
	if !lo.tree.Exists(TypePath(path)) {
		return fmt.Errorf("no such type path: %s", path)
	}
	atom := lo.tree.New(TypePath(path)).AsAtom()
	lo.m[l.X][l.Y][l.Z] = append(lo.m[l.X][l.Y][l.Z], atom)
	return nil
}

func LoadMap(tree *Tree, text string) (Map, error) {
	l := loaderObserver{
		tree: tree,
	}
	err := parsemap.ProduceMap(text, &l)
	if err != nil {
		return nil, err
	}
	return l.m, nil
}

func LoadMapFromFile(tree *Tree, filename string) (Map, error) {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return LoadMap(tree, string(content))
}
