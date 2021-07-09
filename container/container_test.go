package container_test

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/container"
)

type KVStoreKey struct {
	name string
}

type ModuleKey string

type MsgClientA struct {
	key ModuleKey
}

type KeeperA struct {
	key KVStoreKey
}

type KeeperB struct {
	key        KVStoreKey
	msgClientA MsgClientA
}

type Handler struct{}

func ProvideKVStoreKey(scope container.Scope) KVStoreKey {
	return KVStoreKey{name: scope.Name()}
}

func ProvideModuleKey(scope container.Scope) ModuleKey {
	return ModuleKey(scope.Name())
}

func ProvideMsgClientA(key ModuleKey) MsgClientA {
	return MsgClientA{key}
}

func ProvideKeeperA(key KVStoreKey) (KeeperA, Handler) {
	return KeeperA{key}, Handler{}
}

func ProvideKeeperB(key KVStoreKey, a MsgClientA) (KeeperB, Handler) {
	return KeeperB{
		key:        key,
		msgClientA: a,
	}, Handler{}
}

func TestRun(t *testing.T) {
	require.NoError(t,
		container.Run(func(handlers []Handler) {
			// TODO
		}),
		container.DefineGroupTypes(reflect.TypeOf(Handler{})),
		container.Provide(
			ProvideKVStoreKey,
			ProvideModuleKey,
			ProvideMsgClientA,
		),
		container.ProvideWithScope(container.NewScope("a"), ProvideKeeperA),
		container.ProvideWithScope(container.NewScope("b"), ProvideKeeperB),
	)
}
