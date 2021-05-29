package module

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

type storeKey struct {
	name string
}

type keeperA struct {
	key storeKey
}

type keeperB struct {
	key storeKey
	a   keeperA
}

func TestContainer(t *testing.T) {
	c := NewContainer()
	require.NoError(t, c.Provide(Provider{
		Constructor: func(deps []reflect.Value) ([]reflect.Value, error) {
			return []reflect.Value{reflect.ValueOf(keeperA{deps[0].Interface().(storeKey)})}, nil
		},
		Needs: []Key{
			{
				Type: reflect.TypeOf(storeKey{}),
			},
		},
		Provides: []Key{
			{
				Type: reflect.TypeOf((*keeperA)(nil)),
			},
		},
		Scope: "a",
	}))
	require.NoError(t, c.Provide(Provider{
		Constructor: func(deps []reflect.Value) ([]reflect.Value, error) {
			return []reflect.Value{reflect.ValueOf(keeperB{
				key: deps[0].Interface().(storeKey),
				a:   deps[1].Interface().(keeperA),
			})}, nil
		},
		Needs: []Key{
			{
				Type: reflect.TypeOf(storeKey{}),
			},
			{
				Type: reflect.TypeOf((*keeperA)(nil)),
			},
		},
		Provides: []Key{
			{
				Type: reflect.TypeOf((*keeperB)(nil)),
			},
		},
		Scope: "b",
	}))
	require.NoError(t, c.ProvideForScope(
		ScopedProvider{
			Constructor: func(scope Scope, deps []reflect.Value) ([]reflect.Value, error) {
				return []reflect.Value{reflect.ValueOf(storeKey{name: scope})}, nil
			},
			Needs: nil,
			Provides: []Key{
				{
					Type: reflect.TypeOf(storeKey{}),
				},
			},
		},
	))

	res, err := c.Resolve("b", Key{Type: reflect.TypeOf((*keeperB)(nil))})
	require.NoError(t, err)
	b := res.Interface().(keeperB)
	t.Logf("%+v", b)
	require.Equal(t, "b", b.key.name)
	require.Equal(t, "a", b.a.key.name)
}

func TestCycle(t *testing.T) {
	c := NewContainer()
	require.NoError(t, c.Provide(Provider{
		Constructor: func(deps []reflect.Value) ([]reflect.Value, error) {
			return nil, nil
		},
		Needs: []Key{
			{
				Type: reflect.TypeOf((*keeperB)(nil)),
			},
		},
		Provides: []Key{
			{
				Type: reflect.TypeOf((*keeperA)(nil)),
			},
		},
	}))
	require.NoError(t, c.Provide(Provider{
		Constructor: func(deps []reflect.Value) ([]reflect.Value, error) {
			return nil, nil
		},
		Needs: []Key{
			{
				Type: reflect.TypeOf((*keeperA)(nil)),
			},
		},
		Provides: []Key{
			{
				Type: reflect.TypeOf((*keeperB)(nil)),
			},
		},
	}))

	_, err := c.Resolve("b", Key{Type: reflect.TypeOf((*keeperB)(nil))})
	require.EqualError(t, err, "fatal: cycle detected")
}
