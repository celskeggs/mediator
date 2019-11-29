package util

import (
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"testing"
)

func TestSha256sum(t *testing.T) {
	f, err := ioutil.TempFile("", "sha256sum-test")
	assert.NoError(t, err)
	defer func() {
		err := os.Remove(f.Name())
		assert.NoError(t, err)
	}()
	f.Close()

	// hash of empty file
	hash, err := SHA256sum(f.Name())
	assert.NoError(t, err)
	assert.Equal(t, "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855", hash)
}
