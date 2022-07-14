//go:build !depinject

package testgen

import "cosmossdk.io/depinject"

var ScenarioConfig = depinject.Configs(
	depinject.Provide(ProvideMsgClientA),
	depinject.ProvideInModule("runtime", ProvideKVStoreKey),
	depinject.ProvideInModule("a", ModuleA.Provide),
	depinject.ProvideInModule("b", ModuleB.Provide),
)

func Build(modA ModuleA, modB ModuleB) (map[string]Handler, []Command, KeeperA, KeeperB, error,

) {
	moduleKeyContext := &depinject.ModuleKeyContext{}
	kVStoreKeyForA := ProvideKVStoreKey(moduleKeyContext.For("a"))
	keeperA, handler, command := ModuleA.Provide(modA, kVStoreKeyForA, depinject.OwnModuleKey(moduleKeyContext.For("a")))
	kVStoreKeyForB := ProvideKVStoreKey(moduleKeyContext.For("b"))
	msgClientAForB := ProvideMsgClientA(moduleKeyContext.For("b"))
	bProvides, handler2, err := ModuleB.Provide(modB, BDependencies{Key: kVStoreKeyForB, A: msgClientAForB})
	if err != nil {
		return nil, nil, KeeperA{}, KeeperB{}, err
	}
	handlerMap := map[string]Handler{"a": handler, "b": handler2}
	commandSlice := []Command{command}
	commandSlice = append(commandSlice, bProvides.Commands...)
	return handlerMap, commandSlice, keeperA, bProvides.KeeperB, nil

}
