package listenertest

import (
	"cosmossdk.io/schema"
	"cosmossdk.io/schema/listener"
	schematesting "cosmossdk.io/schema/testing"
)

type AppSimulatorOptions struct {
	AppSchema          map[string]schema.ModuleSchema
	Listener           listener.Listener
	EventAlignedWrites bool
	MaxUpdatesPerBlock int
	Seed               int
}

type AppDataSimulator struct {
	state    *schematesting.TestAppState
	options  AppSimulatorOptions
	blockNum uint64
}

func NewAppSimulator(options AppSimulatorOptions) *AppDataSimulator {
	sim := &AppDataSimulator{
		options: options,
	}

	maxUpdates := options.MaxUpdatesPerBlock
	if maxUpdates <= 0 {
		maxUpdates = 100
	}
	//sim.updatesGen = rapid.Custom(func(t *rapid.T) []updateData {
	//	numUpdates := rapid.IntRange(1, maxUpdates).Draw(t, "numUpdates")
	//	updates := make([]updateData, numUpdates)
	//	for i := 0; i < numUpdates; i++ {
	//		moduleIdx := rapid.IntRange(0, sim.modules.Len()-1).Draw(t, "moduleIdx")
	//		keys, values := sim.modules.KeyValues()
	//		modState := values[moduleIdx]
	//		objectIdx := rapid.IntRange(0, modState.Objects.Len()-1).Draw(t, "objectIdx")
	//		objState := modState.Objects.Values()[objectIdx]
	//		update := objState.UpdateGen.Draw(t, "update")
	//		updates[i] = updateData{
	//			module: keys[moduleIdx],
	//			update: update,
	//		}
	//	}
	//	return updates
	//})

	return sim
}

func (a *AppDataSimulator) Initialize() {
	//if f := a.options.Listener.InitializeModuleSchema; f != nil {
	//	a.modules.Scan(func(moduleName string, mod *moduleState) bool {
	//		err := f(moduleName, mod.ModuleSchema)
	//		require.NoError(a.tb, err)
	//		return true
	//	})
	//}
	//
	//if f := a.options.Listener.Initialize; f != nil {
	//	_, err := f(listener.InitializationData{
	//		HasEventAlignedWrites: a.options.EventAlignedWrites,
	//	})
	//	require.NoError(a.tb, err)
	//}
}

//func (a *AppDataSimulator) NextBlock() {
//	a.blockNum++
//
//	if f := a.options.Listener.StartBlock; f != nil {
//		err := f(a.blockNum)
//		if err != nil {
//			require.NoError(a.tb, err)
//		}
//	}
//
//	updates := a.updatesGen.Example(int(a.blockNum) + a.options.Seed)
//	for _, data := range updates {
//		err := a.applyUpdate(data.module, data.update, false)
//		require.NoError(a.tb, err)
//	}
//
//	if f := a.options.Listener.Commit; f != nil {
//		err := f()
//		require.NoError(a.tb, err)
//	}
//}
