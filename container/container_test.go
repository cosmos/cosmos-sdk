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

func ProvideMsgClientA(_ container.Scope, key ModuleKey) MsgClientA {
	return MsgClientA{key}
}

type ModuleA struct{}

func (ModuleA) Provide(key KVStoreKey) (KeeperA, Handler, Command) {
	return KeeperA{key}, Handler{}, Command{}
}

type ModuleB struct{}

type BDependencies struct {
	container.StructArgs

	Key KVStoreKey
	A   MsgClientA
}

type BProvides struct {
	KeeperB  KeeperB
	Handler  Handler
	Commands []Command
}

func (ModuleB) Provide(dependencies BDependencies) BProvides {
	return BProvides{
		KeeperB: KeeperB{
			key:        dependencies.Key,
			msgClientA: dependencies.A,
		},
		Handler:  Handler{},
		Commands: []Command{{}, {}},
	}
}

func TestRun(t *testing.T) {
	t.Skip("Expecting this test to fail for now")
	require.NoError(t,
		container.Run(
			func(handlers map[container.Scope]Handler, commands []Command, a KeeperA, b KeeperB) {
				// TODO:
				// require one Handler for module a and a scopes
				// require 3 commands
				// require KeeperA have store key a
				// require KeeperB have store key b and MsgClientA
			}),
		container.AutoGroupTypes(reflect.TypeOf(Command{})),
		container.OnePerScopeTypes(reflect.TypeOf(Handler{})),
		container.Provide(
			ProvideKVStoreKey,
			ProvideModuleKey,
			ProvideMsgClientA,
		),
		container.ProvideWithScope(container.NewScope("a"), wrapProvideMethod(ModuleA{})),
		container.ProvideWithScope(container.NewScope("b"), wrapProvideMethod(ModuleB{})),
	)
}

func wrapProvideMethod(module interface{}) container.ReflectConstructor {
	method := reflect.TypeOf(module).Method(0)
	methodTy := method.Type
	var in []reflect.Type
	var out []reflect.Type

	for i := 1; i < methodTy.NumIn(); i++ {
		in = append(in, methodTy.In(i))
	}
	for i := 0; i < methodTy.NumOut(); i++ {
		out = append(out, methodTy.Out(i))
	}

	return container.ReflectConstructor{
		In:  in,
		Out: out,
		Fn: func(values []reflect.Value) []reflect.Value {
			values = append([]reflect.Value{reflect.ValueOf(module)}, values...)
			return method.Func.Call(values)
		},
		Location: container.LocationFromPC(method.Func.Pointer()),
	}
}
