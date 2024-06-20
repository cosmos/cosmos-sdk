package indexertesting

import (
	"fmt"

	schema2 "cosmossdk.io/schema"
	"cosmossdk.io/schema/listener"

	"github.com/stretchr/testify/require"
	"github.com/tidwall/btree"
	"pgregory.net/rapid"

	"cosmossdk.io/schema/testing/schemagen"
)

type AppSimulatorOptions struct {
	AppSchema          map[string]schema2.ModuleSchema
	Listener           listener.Listener
	EventAlignedWrites bool
	MaxUpdatesPerBlock int
	Seed               int
}

type AppSimulator struct {
	options    AppSimulatorOptions
	modules    *btree.Map[string, *moduleState]
	blockNum   uint64
	tb         require.TestingT
	updatesGen *rapid.Generator[[]updateData]
}

func NewAppSimulator(tb require.TestingT, options AppSimulatorOptions) *AppSimulator {
	modules := &btree.Map[string, *moduleState]{}
	for module, schema := range options.AppSchema {
		require.True(tb, schema2.ValidateName(module))
		require.NoError(tb, schema.Validate())

		modState := &moduleState{
			ModuleSchema: schema,
			Objects:      &btree.Map[string, *objectState]{},
		}
		modules.Set(module, modState)
		for _, objectType := range schema.ObjectTypes {
			state := &btree.Map[string, schema2.ObjectUpdate]{}
			objState := &objectState{
				ObjectType: objectType,
				Objects:    state,
				UpdateGen:  schemagen.StatefulObjectUpdate(objectType, state),
			}
			modState.Objects.Set(objectType.Name, objState)
		}
	}

	sim := &AppSimulator{
		options: options,
		modules: modules,
		tb:      tb,
	}

	maxUpdates := options.MaxUpdatesPerBlock
	if maxUpdates <= 0 {
		maxUpdates = 100
	}
	sim.updatesGen = rapid.Custom(func(t *rapid.T) []updateData {
		numUpdates := rapid.IntRange(1, maxUpdates).Draw(t, "numUpdates")
		updates := make([]updateData, numUpdates)
		for i := 0; i < numUpdates; i++ {
			moduleIdx := rapid.IntRange(0, sim.modules.Len()-1).Draw(t, "moduleIdx")
			keys, values := sim.modules.KeyValues()
			modState := values[moduleIdx]
			objectIdx := rapid.IntRange(0, modState.Objects.Len()-1).Draw(t, "objectIdx")
			objState := modState.Objects.Values()[objectIdx]
			update := objState.UpdateGen.Draw(t, "update")
			updates[i] = updateData{
				module: keys[moduleIdx],
				update: update,
			}
		}
		return updates
	})

	return sim
}

type updateData struct {
	module string
	update schema2.ObjectUpdate
}

func (a *AppSimulator) Initialize() {
	if f := a.options.Listener.InitializeModuleSchema; f != nil {
		a.modules.Scan(func(moduleName string, mod *moduleState) bool {
			err := f(moduleName, mod.ModuleSchema)
			require.NoError(a.tb, err)
			return true
		})
	}

	if f := a.options.Listener.Initialize; f != nil {
		_, err := f(listener.InitializationData{
			HasEventAlignedWrites: a.options.EventAlignedWrites,
		})
		require.NoError(a.tb, err)
	}
}

func (a *AppSimulator) NextBlock() {
	a.blockNum++

	if f := a.options.Listener.StartBlock; f != nil {
		err := f(a.blockNum)
		if err != nil {
			require.NoError(a.tb, err)
		}
	}

	updates := a.updatesGen.Example(int(a.blockNum) + a.options.Seed)
	for _, data := range updates {
		err := a.applyUpdate(data.module, data.update, false)
		require.NoError(a.tb, err)
	}

	if f := a.options.Listener.Commit; f != nil {
		err := f()
		require.NoError(a.tb, err)
	}
}

func (a *AppSimulator) applyUpdate(module string, update schema2.ObjectUpdate, retainDeletions bool) error {
	if a.options.Listener.OnObjectUpdate != nil {
		err := a.options.Listener.OnObjectUpdate(listener.ObjectUpdateData{
			ModuleName: module,
			Update:     update,
		})
		if err != nil {
			return err
		}
	}

	modState, ok := a.modules.Get(module)
	if !ok {
		return fmt.Errorf("module %v not found", module)
	}

	objState, ok := modState.Objects.Get(update.TypeName)
	if !ok {
		return fmt.Errorf("object type %v not found in module %v", update.TypeName, module)
	}

	require.NoError(a.tb, objState.ObjectType.ValidateObjectUpdate(update))

	keyStr := fmt.Sprintf("%v", update.Key)
	if update.Delete {
		if retainDeletions {
			cur, ok := objState.Objects.Get(keyStr)
			if !ok {
				return fmt.Errorf("object not found for deletion: %v", update.Key)
			}

			cur.Delete = true
			objState.Objects.Set(keyStr, cur)
		} else {
			objState.Objects.Delete(keyStr)
		}
	} else {
		objState.Objects.Set(keyStr, update)
	}

	return nil
}

type moduleState struct {
	ModuleSchema schema2.ModuleSchema
	Objects      *btree.Map[string, *objectState]
}

type objectState struct {
	ObjectType schema2.ObjectType
	Objects    *btree.Map[string, schema2.ObjectUpdate]
	UpdateGen  *rapid.Generator[schema2.ObjectUpdate]
}
