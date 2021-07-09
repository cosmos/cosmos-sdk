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

type Handler struct {
	Handle func()
}

type Command struct {
	Run func()
}

func ProvideKVStoreKey(scope container.Scope) KVStoreKey {
	return KVStoreKey{name: scope.Name()}
}

func ProvideModuleKey(scope container.Scope) ModuleKey {
	return ModuleKey(scope.Name())
}

func ProvideMsgClientA(key ModuleKey) MsgClientA {
	return MsgClientA{key}
}

func ProvideKeeperA(key KVStoreKey) (KeeperA, Handler, Command) {
	return KeeperA{key}, Handler{}, Command{}
}

func ProvideKeeperB(key KVStoreKey, a MsgClientA) (KeeperB, Handler, []Command) {
	return KeeperB{
		key:        key,
		msgClientA: a,
	}, Handler{}, []Command{{}, {}}
}

func TestRun(t *testing.T) {
	t.Skip("Expecting this test to fail for now")
	require.NoError(t,
		container.Run(func(handlers map[container.Scope]Handler, commands []Command) {}),
		container.AutoGroupTypes(reflect.TypeOf(Command{})),
		container.OnePerScopeTypes(reflect.TypeOf(Handler{})),
		container.Provide(
			ProvideKVStoreKey,
			ProvideModuleKey,
			ProvideMsgClientA,
		),
		container.ProvideWithScope(container.NewScope("a"), ProvideKeeperA),
		container.ProvideWithScope(container.NewScope("b"), ProvideKeeperB),
	)
}
