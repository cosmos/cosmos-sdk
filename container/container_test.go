package container_test

import (
	"fmt"
	"reflect"
	"testing"

	reflect2 "github.com/cosmos/cosmos-sdk/container/reflect"

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
	container.StructArgs

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
	require.NoError(t,
		container.Run(
			func(handlers map[container.Scope]Handler, commands []Command, a KeeperA) {
				// TODO:
				// require one Handler for module a and a scopes
				// require 3 commands
				// require KeeperA have store key a
				// require KeeperB have store key b and MsgClientA
			},
			container.Logger(func(o string) { t.Log(o) }),
			container.AutoGroupTypes(reflect.TypeOf(Command{})),
			container.OnePerScopeTypes(reflect.TypeOf(Handler{})),
			container.Provide(
				ProvideKVStoreKey,
				ProvideModuleKey,
				ProvideMsgClientA,
			),
			container.ProvideWithScope(container.NewScope("a"), wrapProvideMethod(ModuleA{})),
			//container.ProvideWithScope(container.NewScope("b"), wrapProvideMethod(ModuleB{})),
		))
}

func wrapProvideMethod(module interface{}) reflect2.Constructor {
	method := reflect.TypeOf(module).Method(0)
	methodTy := method.Type
	var in []reflect2.Input
	var out []reflect2.Output

	for i := 1; i < methodTy.NumIn(); i++ {
		in = append(in, reflect2.Input{Type: methodTy.In(i)})
	}
	for i := 0; i < methodTy.NumOut(); i++ {
		out = append(out, reflect2.Output{Type: methodTy.Out(i)})
	}

	return reflect2.Constructor{
		In:  in,
		Out: out,
		Fn: func(values []reflect.Value) []reflect.Value {
			values = append([]reflect.Value{reflect.ValueOf(module)}, values...)
			return method.Func.Call(values)
		},
		Location: reflect2.LocationFromPC(method.Func.Pointer()),
	}
}

func TestError(t *testing.T) {
	err := container.Run(func() {}, container.Error(fmt.Errorf("an error")))
	require.Error(t, err)
}
