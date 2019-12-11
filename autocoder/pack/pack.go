package pack

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"github.com/celskeggs/mediator/dream/parser"
	"github.com/celskeggs/mediator/util"
	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
	"io/ioutil"
	"os"
	"path"
	"strings"
)

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
	for _, searchdir := range dmf.SearchPath {
		resources, err := ScanDirectory(searchdir)
		if err != nil {
			return nil, err
		}
		paths = append(paths, resources...)
	}
	return paths, nil
}

func BuildTarball(files map[string]string, output string) (e error) {
	out, err := os.OpenFile(output, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o755)
	if err != nil {
		return err
	}
	defer func() {
		e2 := out.Close()
		if e2 != nil {
			e = multierror.Append(e, e2)
		}
		if e != nil {
			// remove output so that there's less of a chance of having a malformed file
			e3 := os.Remove(output)
			e = multierror.Append(e, e3)
		}
	}()
	zout, err := gzip.NewWriterLevel(out, gzip.BestCompression)
	if err != nil {
		return err
	}
	defer func() {
		e2 := zout.Close()
		if e2 != nil {
			e = multierror.Append(e, e2)
		}
	}()
	tarw := tar.NewWriter(zout)
	defer func() {
		e2 := tarw.Close()
		if e2 != nil {
			e = multierror.Append(e, e2)
		}
	}()
	for innerName, outerName := range files {
		fi, err := os.Stat(outerName)
		if err != nil {
			return err
		}
		if !fi.Mode().IsRegular() {
			return fmt.Errorf("expected file %q for insertion to be regular", outerName)
		}
		contents, err := ioutil.ReadFile(outerName)
		if err != nil {
			return err
		}
		err = tarw.WriteHeader(&tar.Header{
			Typeflag:   tar.TypeReg,
			Name:       innerName,
			Size:       int64(len(contents)),
			Mode:       0o644,
			ModTime:    fi.ModTime(),
		})
		if err != nil {
			return err
		}
		_, err = tarw.Write(contents)
		if err != nil {
			return err
		}
	}
	return nil
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
	err = BuildTarball(mapping, outputPack)
	if err != nil {
		return err
	}
	return nil
}
