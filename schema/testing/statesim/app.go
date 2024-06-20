package statesim

import (
	"fmt"

	"github.com/stretchr/testify/require"
	"github.com/tidwall/btree"
	"pgregory.net/rapid"

	"cosmossdk.io/schema"
	"cosmossdk.io/schema/listener"
)

type App struct {
	moduleStates *btree.Map[string, *Module]
	updateGen    *rapid.Generator[listener.ObjectUpdateData]
}

func NewApp(appSchema map[string]schema.ModuleSchema) *App {
	moduleStates := &btree.Map[string, *Module]{}
	var moduleNames []string

	for moduleName, moduleSchema := range appSchema {
		moduleState := NewModule(moduleSchema)
		moduleStates.Set(moduleName, moduleState)
		moduleNames = append(moduleNames, moduleName)
	}

	moduleNameSelector := rapid.Map(rapid.IntRange(0, len(moduleNames)), func(u int) string {
		return moduleNames[u]
	})

	updateGen := rapid.Custom(func(t *rapid.T) listener.ObjectUpdateData {
		moduleName := moduleNameSelector.Draw(t, "moduleName")
		moduleState, ok := moduleStates.Get(moduleName)
		require.True(t, ok)
		return listener.ObjectUpdateData{
			ModuleName: moduleName,
			Update:     moduleState.UpdateGen().Draw(t, "update"),
		}
	})

	return &App{
		moduleStates: moduleStates,
		updateGen:    updateGen,
	}
}

func (a *App) ApplyUpdate(moduleName string, update schema.ObjectUpdate) error {
	moduleState, ok := a.moduleStates.Get(moduleName)
	if !ok {
		return fmt.Errorf("module %s not found", moduleName)
	}

	return moduleState.ApplyUpdate(update)
}

func (a *App) UpdateGen() *rapid.Generator[listener.ObjectUpdateData] {
	return a.updateGen
}

func (a *App) ScanModuleSchemas(f func(string, schema.ModuleSchema) error) error {
	var err error
	a.moduleStates.Scan(func(key string, value *Module) bool {
		err = f(key, value.moduleSchema)
		return err == nil
	})
	return err
}
