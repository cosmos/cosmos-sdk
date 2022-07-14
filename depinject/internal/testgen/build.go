//go:build depinject

package testgen

import "cosmossdk.io/depinject"

var appConfig = depinject.Configs(
	depinject.Provide(ProvideMsgClientA),
	depinject.ProvideInModule("runtime", ProvideKVStoreKey),
	depinject.ProvideInModule("a", ModuleA.Provide),
	depinject.ProvideInModule("b", ModuleB.Provide),
)

func Build(modA ModuleA, modB ModuleB) (
	handlers map[string]Handler,
	commands []Command,
	a KeeperA,
	b KeeperB,
	err error,
) {
	err = depinject.InjectDebug(
		depinject.Codegen(),
		depinject.Configs(
			appConfig,
			depinject.Supply(modA, modB),
		),
		&handlers,
		&commands,
		&a,
		&b,
	)
	return
}
