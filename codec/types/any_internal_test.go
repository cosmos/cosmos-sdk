package types

import (
	"testing"

	"github.com/gogo/protobuf/proto"
	"github.com/stretchr/testify/require"
)

type Dog struct {
	Name string `protobuf:"bytes,1,opt,name=size,proto3" json:"size,omitempty"`
}

func (d Dog) Greet() string { return d.Name }

// We implement a minimal proto.Message interface
func (d *Dog) Reset()                  { d.Name = "" }
func (d *Dog) String() string          { return d.Name }
func (d *Dog) ProtoMessage()           {}
func (d *Dog) XXX_MessageName() string { return "tests/dog" }

type Animal interface {
	Greet() string
}

var _ Animal = (*Dog)(nil)
var _ proto.Message = (*Dog)(nil)

func TestAnyPackUnpack(t *testing.T) {
	registry := NewInterfaceRegistry()
	registry.RegisterInterface("Animal", (*Animal)(nil))
	registry.RegisterImplementations(
		(*Animal)(nil),
		&Dog{},
	)

	spot := &Dog{Name: "Spot"}
	var animal Animal

	// with cache
	any, err := NewAnyWithValue(spot)
	require.NoError(t, err)
	require.Equal(t, spot, any.GetCachedValue())
	err = registry.UnpackAny(any, &animal)
	require.NoError(t, err)
	require.Equal(t, spot, animal)

	// without cache
	any.cachedValue = nil
	err = registry.UnpackAny(any, &animal)
	require.NoError(t, err)
	require.Equal(t, spot, animal)
}
