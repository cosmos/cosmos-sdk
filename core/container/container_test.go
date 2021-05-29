package container

import (
	"testing"

	"github.com/stretchr/testify/require"
)

type ssKey struct {
	name string
	db   db
}

type scKey struct {
	name string
	db   db
}

type kvStoreKey struct {
	ssKey
	scKey
}

type keeperA struct {
	key kvStoreKey
}

type keeperB struct {
	key kvStoreKey
	a   keeperA
}

type db struct{}

func dbProvider() db {
	return db{}
}

func ssKeyProvider(scope Scope, db db) ssKey {
	return ssKey{db: db, name: string(scope)}
}

func scKeyProvider(scope Scope, db db) scKey {
	return scKey{db: db, name: string(scope)}
}

type kvStoreKeyInput struct {
	StructArgs
	SSKey ssKey
	SCKey scKey
}

func kvStoreKeyProvider(scope Scope, input kvStoreKeyInput) kvStoreKey {
	return kvStoreKey{input.SSKey, input.SCKey}
}

func keeperAProvider(key kvStoreKey) keeperA {
	return keeperA{key: key}
}

func keeperBProvider(key kvStoreKey, a keeperA) keeperB {
	return keeperB{key, a}
}

func TestContainer(t *testing.T) {
	c := NewContainer()
	require.NoError(t, c.Provide(dbProvider))
	require.NoError(t, c.Provide(ssKeyProvider))
	require.NoError(t, c.Provide(scKeyProvider))
	require.NoError(t, c.Provide(kvStoreKeyProvider))
	require.NoError(t, c.ProvideWithScope(keeperAProvider, "a"))
	require.NoError(t, c.ProvideWithScope(keeperBProvider, "b"))
	require.NoError(t, c.Invoke(func(b keeperB) {
		require.Equal(t, keeperB{
			key: kvStoreKey{
				ssKey: ssKey{
					name: "b",
					db:   db{},
				},
				scKey: scKey{
					name: "b",
					db:   db{},
				},
			},
			a: keeperA{
				key: kvStoreKey{
					ssKey: ssKey{
						name: "a",
						db:   db{},
					},
					scKey: scKey{
						name: "a",
						db:   db{},
					},
				},
			},
		}, b)
	}))
}

func TestCycle(t *testing.T) {
	c := NewContainer()
	require.NoError(t, c.Provide(func(a keeperA) keeperB {
		return keeperB{}
	}))
	require.NoError(t, c.Provide(func(a keeperB) keeperA {
		return keeperA{}
	}))
	require.EqualError(t, c.Invoke(func(a keeperA) {}), "fatal: cycle detected")
}

//func TestContainer(t *testing.T) {
//	c := NewContainer()
//	require.NoError(t, c.RegisterProvider(Provider{
//		Constructor: func(deps []reflect.Value) ([]reflect.Value, error) {
//			return []reflect.Value{reflect.ValueOf(keeperA{deps[0].Interface().(storeKey)})}, nil
//		},
//		Needs: []Key{
//			{
//				Type: reflect.TypeOf(storeKey{}),
//			},
//		},
//		Provides: []Key{
//			{
//				Type: reflect.TypeOf((*keeperA)(nil)),
//			},
//		},
//		Scope: "a",
//	}))
//	require.NoError(t, c.RegisterProvider(Provider{
//		Constructor: func(deps []reflect.Value) ([]reflect.Value, error) {
//			return []reflect.Value{reflect.ValueOf(keeperB{
//				key: deps[0].Interface().(storeKey),
//				a:   deps[1].Interface().(keeperA),
//			})}, nil
//		},
//		Needs: []Input{
//			{
//				Key: Key{
//					Type: reflect.TypeOf(storeKey{}),
//				},
//			},
//			{
//				Key: Key{
//					Type: reflect.TypeOf((*keeperA)(nil)),
//				},
//			},
//		},
//		Provides: []Output{
//			{
//				Key: Key{
//					Type: reflect.TypeOf((*keeperB)(nil)),
//				},
//			},
//		},
//		Scope: "b",
//	}))
//	require.NoError(t, c.RegisterScopeProvider(
//		ScopeProvider{
//			Constructor: func(scope Scope, deps []reflect.Value) ([]reflect.Value, error) {
//				return []reflect.Value{reflect.ValueOf(storeKey{name: scope})}, nil
//			},
//			Needs: nil,
//			Provides: []Key{
//				{
//					Type: reflect.TypeOf(storeKey{}),
//				},
//			},
//		},
//	))
//
//	res, err := c.Resolve("b", Key{Type: reflect.TypeOf((*keeperB)(nil))})
//	require.NoError(t, err)
//	b := res.Interface().(keeperB)
//	t.Logf("%+v", b)
//	require.Equal(t, "b", b.key.name)
//	require.Equal(t, "a", b.a.key.name)
//}
//
//func TestCycle(t *testing.T) {
//	c := NewContainer()
//	require.NoError(t, c.RegisterProvider(Provider{
//		Constructor: func(deps []reflect.Value) ([]reflect.Value, error) {
//			return nil, nil
//		},
//		Needs: []Key{
//			{
//				Type: reflect.TypeOf((*keeperB)(nil)),
//			},
//		},
//		Provides: []Key{
//			{
//				Type: reflect.TypeOf((*keeperA)(nil)),
//			},
//		},
//	}))
//	require.NoError(t, c.RegisterProvider(Provider{
//		Constructor: func(deps []reflect.Value) ([]reflect.Value, error) {
//			return nil, nil
//		},
//		Needs: []Key{
//			{
//				Type: reflect.TypeOf((*keeperA)(nil)),
//			},
//		},
//		Provides: []Key{
//			{
//				Type: reflect.TypeOf((*keeperB)(nil)),
//			},
//		},
//	}))
//
//	_, err := c.Resolve("b", Key{Type: reflect.TypeOf((*keeperB)(nil))})
//	require.EqualError(t, err, "fatal: cycle detected")
//}
