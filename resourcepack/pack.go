package resourcepack

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"github.com/hashicorp/go-multierror"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"time"
)

func Build(files map[string]string, output string) (e error) {
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
			Typeflag: tar.TypeReg,
			Name:     innerName,
			Size:     int64(len(contents)),
			Mode:     0o644,
			ModTime:  fi.ModTime(),
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

type ResourcePack struct {
	Resources map[string]Resource
}

type Resource struct {
	Name     string
	Data     []byte
	Modified time.Time
}

func (r Resource) IsMap() bool {
	return strings.HasSuffix(r.Name, ".dmm")
}

func (r Resource) IsIcon() bool {
	return strings.HasSuffix(r.Name, ".dmi")
}

func (r Resource) IsWeb() bool {
	return strings.HasSuffix(r.Name, ".css") || strings.HasSuffix(r.Name, ".js") || strings.HasSuffix(r.Name, ".html")
}

func (p *ResourcePack) Resource(name string) (Resource, error) {
	r, ok := p.Resources[name]
	if !ok {
		return Resource{}, fmt.Errorf("cannot find resource %q in resource pack", name)
	}
	return r, nil
}

func (p *ResourcePack) Icons() (icons []string) {
	for name, resource := range p.Resources {
		if resource.IsIcon() {
			icons = append(icons, name)
		}
	}
	return icons
}

func Load(filename string) (_ *ResourcePack, e error) {
	in, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer func() {
		e2 := in.Close()
		if e2 != nil {
			e = multierror.Append(e, e2)
		}
	}()
	zin, err := gzip.NewReader(in)
	if err != nil {
		return nil, err
	}
	defer func() {
		e2 := zin.Close()
		if e2 != nil {
			e = multierror.Append(e, e2)
		}
	}()
	tarin := tar.NewReader(zin)
	files := map[string]Resource{}
	for {
		hdr, err := tarin.Next()
		if err == io.EOF {
			return &ResourcePack{
				Resources: files,
			}, nil
		} else if err != nil {
			return nil, err
		}
		if hdr.Typeflag != tar.TypeReg {
			fmt.Printf("ignoring unexpected non-regular-file %q in resource pack\n", hdr.Name)
			continue
		}
		data, err := ioutil.ReadAll(tarin)
		if err != nil {
			return nil, err
		}
		if _, exists := files[hdr.Name]; exists {
			return nil, fmt.Errorf("duplicate file %q found in resource pack", hdr.Name)
		}
		files[hdr.Name] = Resource{
			Name:     hdr.Name,
			Data:     data,
			Modified: hdr.ModTime,
		}
	}
}
