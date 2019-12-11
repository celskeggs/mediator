package pack

import (
	"fmt"
	"github.com/celskeggs/mediator/dream/parser"
	"github.com/celskeggs/mediator/resourcepack"
	"github.com/celskeggs/mediator/util"
	"github.com/pkg/errors"
	"go/build"
	"io/ioutil"
	"os"
	"path"
	"strings"
)

const CoreResourcesPath = "github.com/celskeggs/mediator/resources"
const OutputSuffix = ".tgz"

func IsResourceExtension(ext string) bool {
	switch ext {
	case ".dmm":
		return true
	case ".dmi":
		return true
	case ".mid":
		return true
	case ".ogg":
		return true
	case ".wav":
		return true
	case ".js":
		return true
	case ".html":
		return true
	case ".css":
		return true
	default:
		return false
	}
}

func ScanDirectory(dir string) (paths []string, _ error) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	for _, fi := range files {
		if IsResourceExtension(path.Ext(fi.Name())) {
			paths = append(paths, path.Join(dir, fi.Name()))
		}
	}
	return paths, nil
}

func ScanResources(dmf *parser.DreamMakerFile) (paths []string, _ error) {
	coreResources, err := build.Default.Import(CoreResourcesPath, "", build.FindOnly)
	if err != nil {
		return nil, errors.Wrapf(err, "while finding core resources at path %v", CoreResourcesPath)
	}
	paths, err = ScanDirectory(coreResources.Dir)
	if err != nil {
		return nil, err
	}
	for _, searchdir := range dmf.SearchPath {
		resources, err := ScanDirectory(searchdir)
		if err != nil {
			return nil, err
		}
		paths = append(paths, resources...)
	}
	return paths, nil
}

func GenerateResourcePack(dmf *parser.DreamMakerFile, outputPack string) error {
	if !strings.HasSuffix(outputPack, OutputSuffix) {
		return fmt.Errorf("output resource pack name does not end in %s: %q", OutputSuffix, outputPack)
	}
	err := os.Remove(outputPack)
	if err != nil && !os.IsNotExist(err) {
		return errors.Wrap(err, "while deleting existing output pack")
	}
	util.FIXME("do .mid -> .ogg conversion in here somewhere")
	resources, err := ScanResources(dmf)
	if err != nil {
		return err
	}
	mapping := map[string]string{}
	for _, resource := range resources {
		name := path.Base(resource)
		if prev, exists := mapping[name]; exists {
			return fmt.Errorf("duplicate resource %q found at both %q and %q", name, prev, resource)
		}
		mapping[name] = resource
	}
	err = resourcepack.Build(mapping, outputPack)
	if err != nil {
		return err
	}
	return nil
}
