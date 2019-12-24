package detect

import (
	"fmt"
	"path"
	"path/filepath"
)

func ChopSrc(filepath string) (string, bool) {
	if !path.IsAbs(filepath) {
		panic("ChopSrc must receive an absolute path")
	}
	if filepath == "/" {
		return "", false
	}
	dir, filename := path.Split(filepath)
	if filename == "src" {
		return "", true
	} else {
		chop, ok := ChopSrc(path.Clean(dir))
		if !ok {
			return "", false
		}
		return path.Join(chop, filename), true
	}
}

func DetectImportPath(filename string) (string, error) {
	abspath, err := filepath.Abs(filename)
	if err != nil {
		return "", err
	}
	packagedir, _ := path.Split(abspath)
	importpath, ok := ChopSrc(packagedir)
	if !ok {
		return "", fmt.Errorf("cannot extract import path from %q (abs of %q)", abspath, filename)
	}
	return importpath, nil
}
