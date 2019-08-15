package datum

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

type brokenDatum struct {
	IDatum
}

func (b brokenDatum) RawClone() IDatum {
	return &b
}

func TestBrokenRawClone(t *testing.T) {
	ntt := NewTypeTree()
	a := &brokenDatum{
		IDatum: ntt.DeriveNew("/datum"),
	}
	setImpl(a)

	assert.PanicsWithValue(t, "RawClone() failed to do a full deep clone", func() {
		a.Clone()
	})
}

type brokenDatum2 struct {
	IDatum
}

func TestMissingRawClone(t *testing.T) {
	ntt := NewTypeTree()
	a := &brokenDatum2{
		IDatum: ntt.DeriveNew("/datum"),
	}
	setImpl(a)

	assert.PanicsWithValue(t, "RawClone() did not return the same type that went in", func() {
		a.Clone()
	})
}
