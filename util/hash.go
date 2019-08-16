package util

import (
	"os"
	"crypto/sha256"
	"io"
	"encoding/hex"
)

func SHA256sum(file string) (string, error) {
	f, err := os.Open(file)
	if err != nil {
		return "", err
	}
	var buf [4096]byte
	hasher := sha256.New()
	for {
		num, err := f.Read(buf[:])
		if err == io.EOF {
			break
		} else if err != nil {
			return "", err
		}
		hasher.Write(buf[:num])
	}
	return hex.EncodeToString(hasher.Sum(nil)), nil
}
