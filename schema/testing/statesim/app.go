package statesim

import (
	"fmt"

	"github.com/stretchr/testify/require"
	"github.com/tidwall/btree"
	"pgregory.net/rapid"

	"cosmossdk.io/schema"
	"cosmossdk.io/schema/appdata"
	"cosmossdk.io/schema/view"
)

// App is a collection of simulated module states corresponding to an app's schema for testing purposes.
type App struct {
	options      Options
	moduleStates *btree.Map[string, *Module]
	updateGen    *rapid.Generator[appdata.ObjectUpdateData]
}

// NewApp creates a new simulation App for the given app schema. The app schema can be nil
// if the user desires initializing modules with InitializeModule instead.
func NewApp(appSchema map[string]schema.ModuleSchema, options Options) *App {
	app := &App{
		options:      options,
		moduleStates: &btree.Map[string, *Module]{},
	}

	for moduleName, moduleSchema := range appSchema {
		moduleState := NewModule(moduleName, moduleSchema, options)
		app.moduleStates.Set(moduleName, moduleState)
	}

	moduleNameSelector := rapid.Custom(func(t *rapid.T) string {
		return rapid.SampledFrom(app.moduleStates.Keys()).Draw(t, "moduleName")
	})

	numUpdatesGen := rapid.IntRange(1, 2)
	app.updateGen = rapid.Custom(func(t *rapid.T) appdata.ObjectUpdateData {
		moduleName := moduleNameSelector.Draw(t, "moduleName")
		moduleState, ok := app.moduleStates.Get(moduleName)
		require.True(t, ok)
		numUpdates := numUpdatesGen.Draw(t, "numUpdates")
		updates := make([]schema.StateObjectUpdate, numUpdates)
		for i := 0; i < numUpdates; i++ {
			update := moduleState.UpdateGen().Draw(t, fmt.Sprintf("update[%d]", i))
			updates[i] = update
		}
		return appdata.ObjectUpdateData{
			ModuleName: moduleName,
			Updates:    updates,
		}
	})

	return app
}

// InitializeModule initializes the module with the provided schema. This returns an error if the
// module is already initialized in state.
func (a *App) InitializeModule(data appdata.ModuleInitializationData) error {
	if _, ok := a.moduleStates.Get(data.ModuleName); ok {
		return fmt.Errorf("module %s already initialized", data.ModuleName)
	}

	a.moduleStates.Set(data.ModuleName, NewModule(data.ModuleName, data.Schema, a.options))
	return nil
}

// ApplyUpdate applies the given object update to the module.
func (a *App) ApplyUpdate(data appdata.ObjectUpdateData) error {
	moduleState, ok := a.moduleStates.Get(data.ModuleName)
	if !ok {
		// we don't have this module so skip the update
		return nil
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
func (a *App) GetModule(moduleName string) (view.ModuleState, error) {
	mod, ok := a.moduleStates.Get(moduleName)
	if !ok {
		return nil, nil
	}
	return mod, nil
}

// Modules iterates over all the module state instances in the app.
func (a *App) Modules(f func(modState view.ModuleState, err error) bool) {
	a.moduleStates.Scan(func(key string, value *Module) bool {
		return f(value, nil)
	})
}

// NumModules returns the number of modules in the app.
func (a *App) NumModules() (int, error) {
	return a.moduleStates.Len(), nil
}
