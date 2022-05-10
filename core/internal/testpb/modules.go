package testpb

import (
	"fmt"
	"io"
	"sort"

	"github.com/cosmos/cosmos-sdk/container"

	"cosmossdk.io/core/appmodule"
)

func init() {
	appmodule.Register(&TestRuntimeModule{},
		appmodule.Provide(provideRuntimeState, provideStoreKey, provideApp),
	)

	appmodule.Register(&TestModuleA{},
		appmodule.Provide(provideModuleA),
	)

	appmodule.Register(&TestModuleB{},
		appmodule.Provide(provideModuleB),
	)
}

func provideRuntimeState() *runtimeState {
	return &runtimeState{}
}

func provideStoreKey(key container.ModuleKey, state *runtimeState) StoreKey {
	sk := StoreKey{name: key.Name()}
	state.storeKeys = append(state.storeKeys, sk)
	return sk
}

func provideApp(state *runtimeState, handlers map[string]Handler) App {
	return func(w io.Writer) {
		for _, key := range state.storeKeys {
			_, _ = fmt.Fprintf(w, "got store key %s\n", key.name)
		}

		var modNames []string
		for modName := range handlers {
			modNames = append(modNames, modName)
		}

		sort.Strings(modNames)
		for _, name := range modNames {
			_, _ = fmt.Fprintf(w, "running module handler %s\n", name)
			_, _ = fmt.Fprintf(w, "result: %s\n", handlers[name].DoSomething())
		}
	}
}

type App func(writer io.Writer)

type runtimeState struct {
	storeKeys []StoreKey
}

type StoreKey struct{ name string }

type Handler struct {
	DoSomething func() string
}

func (h Handler) IsOnePerModuleType() {}

func provideModuleA(key StoreKey) (KeeperA, Handler) {
	return keeperA{key: key}, Handler{DoSomething: func() string {
		return "hello"
	}}
}

type keeperA struct {
	key StoreKey
}

type KeeperA interface {
	Foo()
}

func (k keeperA) Foo() {}

func provideModuleB(key StoreKey, a KeeperA) (KeeperB, Handler) {
	return keeperB{key: key, a: a}, Handler{
		DoSomething: func() string {
			return "goodbye"
		},
	}
}

type keeperB struct {
	key StoreKey
	a   KeeperA
}

type KeeperB interface{}
