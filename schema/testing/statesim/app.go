package statesim

import (
	"fmt"

	"github.com/stretchr/testify/require"
	"github.com/tidwall/btree"
	"pgregory.net/rapid"

	"cosmossdk.io/schema"
	"cosmossdk.io/schema/appdata"
)

// App is a collection of simulated module states corresponding to an app's schema for testing purposes.
type App struct {
	moduleStates *btree.Map[string, *Module]
	updateGen    *rapid.Generator[appdata.ObjectUpdateData]
}

// NewApp creates a new simulation App for the given app schema.
func NewApp(appSchema map[string]schema.ModuleSchema, options Options) *App {
	moduleStates := &btree.Map[string, *Module]{}
	var moduleNames []string

	for moduleName, moduleSchema := range appSchema {
		moduleState := NewModule(moduleSchema, options)
		moduleStates.Set(moduleName, moduleState)
		moduleNames = append(moduleNames, moduleName)
	}

	moduleNameSelector := rapid.Map(rapid.IntRange(0, len(moduleNames)), func(u int) string {
		return moduleNames[u]
	})

	numUpdatesGen := rapid.IntRange(1, 2)
	updateGen := rapid.Custom(func(t *rapid.T) appdata.ObjectUpdateData {
		moduleName := moduleNameSelector.Draw(t, "moduleName")
		moduleState, ok := moduleStates.Get(moduleName)
		require.True(t, ok)
		numUpdates := numUpdatesGen.Draw(t, "numUpdates")
		updates := make([]schema.ObjectUpdate, numUpdates)
		for i := 0; i < numUpdates; i++ {
			update := moduleState.UpdateGen().Draw(t, fmt.Sprintf("update[%d]", i))
			updates[i] = update
		}
		return appdata.ObjectUpdateData{
			ModuleName: moduleName,
			Updates:    updates,
		}
	})

	return &App{
		moduleStates: moduleStates,
		updateGen:    updateGen,
	}
}

// ApplyUpdate applies the given object update to the module.
func (a *App) ApplyUpdate(data appdata.ObjectUpdateData) error {
	moduleState, ok := a.moduleStates.Get(data.ModuleName)
	if !ok {
		return fmt.Errorf("module %s not found", data.ModuleName)
	}

	for _, update := range data.Updates {
		err := moduleState.ApplyUpdate(update)
		if err != nil {
			return err
		}
	}

	return nil
}

// UpdateGen is a generator for object update data against the app. It is stateful and includes a certain number of
// updates and deletions to existing objects.
func (a *App) UpdateGen() *rapid.Generator[appdata.ObjectUpdateData] {
	return a.updateGen
}

// GetModule returns the module state for the given module name.
func (a *App) GetModule(moduleName string) (*Module, bool) {
	return a.moduleStates.Get(moduleName)
}

// ScanModules scans all the module state instances in the app.
func (a *App) ScanModules(f func(moduleName string, modState *Module) error) error {
	var err error
	a.moduleStates.Scan(func(key string, value *Module) bool {
		err = f(key, value)
		return err == nil
	})
	return err
}
