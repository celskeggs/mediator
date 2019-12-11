package icon

import (
	"fmt"
	"github.com/celskeggs/mediator/dmi"
	"github.com/celskeggs/mediator/resourcepack"
	"math"
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
	icons map[string]*Icon
}

func NewIconCache(pack *resourcepack.ResourcePack) (*IconCache, error) {
	ic := &IconCache{
		icons: map[string]*Icon{},
	}
	for _, resource := range pack.Resources {
		if resource.IsIcon() {
			icon, err := loadInternal(resource)
			if err != nil {
				return nil, err
			}
			ic.icons[resource.Name] = icon
		}
	}
	return ic, nil
}

func loadInternal(resource resourcepack.Resource) (*Icon, error) {
	info, err := dmi.ParseDMI(resource.Data)
	if err != nil {
		return nil, err
	}
	indexes, stride, err := precomputeStateIndexes(info)
	if err != nil {
		return nil, err
	}
	return &Icon{
		dmiPath:      resource.Name,
		dmiInfo:      info,
		stateIndexes: indexes,
		stride:       stride,
	}, nil
}

func (i *IconCache) Load(name string) (*Icon, bool) {
	icon, found := i.icons[name]
	return icon, found
}

func (i *IconCache) LoadOrPanic(name string) *Icon {
	icon, ok := i.Load(name)
	if !ok {
		panic(fmt.Sprintf("no such icon %q found in resource pack", name))
	}
	return icon
}
