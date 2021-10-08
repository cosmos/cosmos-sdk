package main

import (
	"reflect"

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

func ProvideModuleKey(scope container.Scope) (ModuleKey, error) {
	return ModuleKey(scope.Name()), nil
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
	container.In

	Key KVStoreKey
	A   MsgClientA
}

type BProvides struct {
	container.Out

	KeeperB  KeeperB
	Commands []Command
}

func (ModuleB) Provide(dependencies BDependencies, _ container.Scope) (BProvides, Handler, error) {
	return BProvides{
		KeeperB: KeeperB{
			key:        dependencies.Key,
			msgClientA: dependencies.A,
		},
		Commands: []Command{{}, {}},
	}, Handler{}, nil
}

func main() {
	container.Run(
		func(handlers map[string]Handler, commands []Command, a KeeperA, b KeeperB) {
			// require.Len(t, handlers, 2)
			// require.Equal(t, Handler{}, handlers["a"])
			// require.Equal(t, Handler{}, handlers["b"])
			// require.Len(t, commands, 3)
			// require.Equal(t, KeeperA{
			// 	key: KVStoreKey{name: "a"},
			// }, a)
			// require.Equal(t, KeeperB{
			// 	key: KVStoreKey{name: "b"},
			// 	msgClientA: MsgClientA{
			// 		key: "b",
			// 	},
			// }, b)
		},
		container.AutoGroupTypes(reflect.TypeOf(Command{})),
		container.OnePerScopeTypes(reflect.TypeOf(Handler{})),
		container.Provide(
			ProvideKVStoreKey,
			ProvideModuleKey,
			ProvideMsgClientA,
		),
		container.ProvideWithScope("a", wrapMethod0(ModuleA{})),
		container.ProvideWithScope("b", wrapMethod0(ModuleB{})),
		container.Visualizer(func(g string) {
		}),
		container.LogVisualizer(),
		container.FileVisualizer("graph", "svg"),
		container.StdoutLogger(),
	)
}

func wrapMethod0(module interface{}) interface{} {
	methodFn := reflect.TypeOf(module).Method(0).Func.Interface()
	ctrInfo, err := container.ExtractProviderDescriptor(methodFn)
	if err != nil {
		panic(err)
	}

	ctrInfo.Inputs = ctrInfo.Inputs[1:]
	fn := ctrInfo.Fn
	ctrInfo.Fn = func(values []reflect.Value) ([]reflect.Value, error) {
		return fn(append([]reflect.Value{reflect.ValueOf(module)}, values...))
	}
	return ctrInfo
}
