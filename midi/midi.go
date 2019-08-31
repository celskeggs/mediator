package midi

import (
	"os/exec"
	"os"
	"github.com/celskeggs/mediator/util"
	"path"
	"log"
	"github.com/pkg/errors"
)

func ConvertMIDICached(sourcePath string, cacheDir string) (string, error) {
	if cacheDir == "" {
		return "", errors.New("expected non-empty cache dir string")
	}
	err := os.MkdirAll(cacheDir, os.FileMode(0755))
	if err != nil {
		return "", err
	}
	baseName, err := util.SHA256sum(sourcePath)
	if err != nil {
		return "", err
	}
	destPath := path.Join(cacheDir, baseName+".ogg")
	fi, err := os.Stat(destPath)
	if !os.IsNotExist(err) && err != nil {
		return "", err
	}
	if os.IsNotExist(err) || fi.Size() == 0 {
		log.Println("converting MIDI", sourcePath)
		err := ConvertMIDIToOgg(sourcePath, destPath)
		if err != nil {
			return "", err
		}
	} else {
		log.Println("reusing existing conversion of MIDI", sourcePath)
	}
	return destPath, nil
}

func ConvertMIDIToOgg(sourcePath string, destPath string) error {
	cmd := exec.Command("timidity", "-o", destPath, "-Ov", "--", sourcePath)
	err := cmd.Run()
	if err != nil {
		return errors.Wrap(err, "running timidity")
	}
	stat, err := os.Stat(destPath)
	if err != nil {
		return err
	}
	if stat.Size() == 0 {
		return errors.New("empty OGG file")
	}
	return nil
}
