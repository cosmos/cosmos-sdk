package basecoin

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// define a Demo struct that implements TxLayer
type Demo struct{}

var _ TxLayer = Demo{}

func (d Demo) Next() Tx             { return Tx{} }
func (d Demo) Wrap() Tx             { return Tx{d} }
func (d Demo) ValidateBasic() error { return nil }

// define a Fake struct that doesn't implement TxLayer
type Fake struct{}

func (f Fake) Wrap() Tx             { return Tx{f} }
func (f Fake) ValidateBasic() error { return nil }

// Make sure the layer
func TestLayer(t *testing.T) {
	assert := assert.New(t)

	// a fake tx, just don't use it...
	nl := Fake{}.Wrap()
	assert.False(nl.IsLayer())
	assert.Nil(nl.GetLayer())

	// a tx containing a TxLayer should respond properly
	l := Demo{}.Wrap()
	assert.True(l.IsLayer())
	assert.NotNil(l.GetLayer())
}
