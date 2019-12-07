package icon

import (
	"fmt"
	"github.com/celskeggs/mediator/dmi"
	"io/ioutil"
	"math"
	"path"
	"sort"
)

func getStatesSorted(info *dmi.DMIInfo) (out []string) {
	for state := range info.States {
		out = append(out, state)
	}
	sort.Slice(out, func(i, j int) bool {
		return info.States[out[i]].Index < info.States[out[j]].Index
	})
	return out
}

func validate(state dmi.DMIState) error {
	if state.Directions != 1 && state.Directions != 4 && state.Directions != 8 {
		return fmt.Errorf("unexpected number of directions: %d", state.Directions)
	}
	if state.Frames < 1 {
		return fmt.Errorf("expected at least one frame, not: %d", state.Frames)
	}
	return nil
}

func precomputeStateIndexes(info *dmi.DMIInfo) (map[string]uint, uint, error) {
	indexes := map[string]uint{}
	sorted := getStatesSorted(info)
	nextIndex := 0
	for _, iconState := range sorted {
		state := info.States[iconState]
		if err := validate(state); err != nil {
			return nil, 0, err
		}
		indexes[iconState] = uint(nextIndex)
		nextIndex += state.Directions * state.Frames
	}
	// assuming that DMIs are stored in approximate squares, ceil(sqrt(num_subicons)) will tell us the width of each row
	stride := uint(math.Ceil(math.Sqrt(float64(nextIndex))))
	return indexes, stride, nil
}

type IconCache struct {
	cacheMap    map[string]*Icon
	resourceDir string
}

func NewIconCache(resourceDir string) *IconCache {
	return &IconCache{
		cacheMap:    map[string]*Icon{},
		resourceDir: resourceDir,
	}
}

func (i *IconCache) loadInternal(name string) (*Icon, error) {
	filepath := path.Join(i.resourceDir, name)
	png, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil, err
	}
	info, err := dmi.ParseDMI(png)
	if err != nil {
		return nil, err
	}
	indexes, stride, err := precomputeStateIndexes(info)
	if err != nil {
		return nil, err
	}
	return &Icon{
		dmiPath:      name,
		dmiInfo:      info,
		stateIndexes: indexes,
		stride:       stride,
	}, nil
}

func (i *IconCache) Load(name string) (*Icon, error) {
	if icon, found := i.cacheMap[name]; found {
		return icon, nil
	}
	icon, err := i.loadInternal(name)
	if err != nil {
		return nil, err
	}
	i.cacheMap[name] = icon
	return icon, nil
}

func (i *IconCache) LoadOrPanic(name string) *Icon {
	icon, err := i.Load(name)
	if err != nil {
		panic("while loading icon " + name + ": " + err.Error())
	}
	if icon == nil {
		panic("icon should not be nil")
	}
	return icon
}
