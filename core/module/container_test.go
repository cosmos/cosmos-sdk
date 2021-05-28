package module

import (
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
		Constructor: func(deps []interface{}) ([]interface{}, error) {
			return []interface{}{keeperA{deps[0].(storeKey)}}, nil
		},
		Needs: []Key{
			{
				Type: storeKey{},
			},
		},
		Provides: []Key{
			{
				Type: (*keeperA)(nil),
			},
		},
		Scope: "a",
	}))
	require.NoError(t, c.Provide(Provider{
		Constructor: func(deps []interface{}) ([]interface{}, error) {
			return []interface{}{keeperB{
				key: deps[0].(storeKey),
				a:   deps[1].(keeperA),
			}}, nil
		},
		Needs: []Key{
			{
				Type: storeKey{},
			},
			{
				Type: (*keeperA)(nil),
			},
		},
		Provides: []Key{
			{
				Type: (*keeperB)(nil),
			},
		},
		Scope: "b",
	}))
	require.NoError(t, c.ProvideForScope(
		ScopedProvider{
			Constructor: func(scope string, deps []interface{}) ([]interface{}, error) {
				return []interface{}{storeKey{name: scope}}, nil
			},
			Needs: nil,
			Provides: []Key{
				{
					Type: storeKey{},
				},
			},
		},
	))

	res, err := c.Resolve("b", Key{Type: (*keeperB)(nil)})
	require.NoError(t, err)
	b := res.(keeperB)
	t.Logf("%+v", b)
	require.Equal(t, "b", b.key.name)
	require.Equal(t, "a", b.a.key.name)
}

func TestCycle(t *testing.T) {
	c := NewContainer()
	require.NoError(t, c.Provide(Provider{
		Constructor: func(deps []interface{}) ([]interface{}, error) {
			return nil, nil
		},
		Needs: []Key{
			{
				Type: (*keeperB)(nil),
			},
		},
		Provides: []Key{
			{
				Type: (*keeperA)(nil),
			},
		},
	}))
	require.NoError(t, c.Provide(Provider{
		Constructor: func(deps []interface{}) ([]interface{}, error) {
			return nil, nil
		},
		Needs: []Key{
			{
				Type: (*keeperA)(nil),
			},
		},
		Provides: []Key{
			{
				Type: (*keeperB)(nil),
			},
		},
	}))

	_, err := c.Resolve("b", Key{Type: (*keeperB)(nil)})
	require.EqualError(t, err, "fatal: cycle detected")
}
