package icon

import (
	"github.com/celskeggs/mediator/dmi"
	"github.com/celskeggs/mediator/util"
	"io/ioutil"
	"path"
)

type Icon struct {
	dmiPath string
	dmiInfo *dmi.DMIInfo
}

func (icon Icon) Render(state string) (iconname string, sourceX, sourceY, sourceWidth, sourceHeight uint) {
	util.FIXME("implement icon states")
	return icon.dmiPath, 0, 0, 32, 32
}

type IconCache struct {
	cacheMap map[string]*Icon
	resourceDir string
}

func NewIconCache(resourceDir string) *IconCache {
	return &IconCache{
		cacheMap: map[string]*Icon{},
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
	return &Icon{
		dmiPath: name,
		dmiInfo: info,
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
