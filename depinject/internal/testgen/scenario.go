package testgen

import "cosmossdk.io/depinject"

type KVStoreKey struct {
	Name string
}

type MsgClientA struct {
	Key string
}

type KeeperA struct {
	Key  KVStoreKey
	Name string
}

type KeeperB struct {
	Key        KVStoreKey
	MsgClientA MsgClientA
}

type Handler struct {
	Handle func()
}

func (Handler) IsOnePerModuleType() {}

type Command struct {
	Run func()
}

func (Command) IsManyPerContainerType() {}

func ProvideKVStoreKey(moduleKey depinject.ModuleKey) KVStoreKey {
	return KVStoreKey{Name: moduleKey.Name()}
}

func ProvideMsgClientA(key depinject.ModuleKey) MsgClientA {
	return MsgClientA{key.Name()}
}

type ModuleA struct{}

func (ModuleA) Provide(key KVStoreKey, moduleKey depinject.OwnModuleKey) (KeeperA, Handler, Command) {
	return KeeperA{Key: key, Name: depinject.ModuleKey(moduleKey).Name()}, Handler{}, Command{}
}

type ModuleB struct{}

type BDependencies struct {
	depinject.In

	Key KVStoreKey
	A   MsgClientA
}

type BProvides struct {
	depinject.Out

	KeeperB  KeeperB
	Commands []Command
}

func (ModuleB) Provide(dependencies BDependencies) (BProvides, Handler, error) {
	return BProvides{
		KeeperB: KeeperB{
			Key:        dependencies.Key,
			MsgClientA: dependencies.A,
		},
		Commands: []Command{{}, {}},
	}, Handler{}, nil
}
