package indexertesting

import (
	"fmt"

	"github.com/stretchr/testify/require"
	"github.com/tidwall/btree"
	"pgregory.net/rapid"

	indexerbase "cosmossdk.io/indexer/base"
	"cosmossdk.io/indexer/testing/schemagen"
)

type AppSimulatorOptions struct {
	AppSchema          map[string]indexerbase.ModuleSchema
	Listener           indexerbase.Listener
	EventAlignedWrites bool
	MaxUpdatesPerBlock int
	Seed               int
}

type AppSimulator struct {
	options  AppSimulatorOptions
	modules  *btree.Map[string, *moduleState]
	blockNum uint64
	tb       require.TestingT
}

func NewAppSimulator(tb require.TestingT, options AppSimulatorOptions) *AppSimulator {
	modules := &btree.Map[string, *moduleState]{}
	for module, schema := range options.AppSchema {
		require.True(tb, indexerbase.ValidateName(module))
		require.NoError(tb, schema.Validate())

		modState := &moduleState{
			ModuleSchema: schema,
			Objects:      &btree.Map[string, *objectState]{},
		}
		modules.Set(module, modState)
		for _, objectType := range schema.ObjectTypes {
			state := &btree.Map[string, indexerbase.ObjectUpdate]{}
			objState := &objectState{
				ObjectType: objectType,
				Objects:    state,
				UpdateGen:  schemagen.StatefulObjectUpdate(objectType, state),
			}
			modState.Objects.Set(objectType.Name, objState)
		}
	}

	return &AppSimulator{
		options: options,
		modules: modules,
		tb:      tb,
	}
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
		_, err := f(indexerbase.InitializationData{
			HasEventAlignedWrites: a.options.EventAlignedWrites,
		})
		require.NoError(a.tb, err)
	}
}

func (a *AppSimulator) NextBlock() {
	a.newBlockFromSeed(a.options.Seed + int(a.blockNum))
}

func (a *AppSimulator) actionNewBlock(t *rapid.T) {
	a.blockNum++

	if f := a.options.Listener.StartBlock; f != nil {
		err := f(a.blockNum)
		if err != nil {
			require.NoError(t, err)
		}
	}

	maxUpdates := a.options.MaxUpdatesPerBlock
	if maxUpdates <= 0 {
		maxUpdates = 100
	}
	numUpdates := rapid.IntRange(1, maxUpdates).Draw(t, "numUpdates")
	for i := 0; i < numUpdates; i++ {
		moduleIdx := rapid.IntRange(0, a.modules.Len()-1).Draw(t, "moduleIdx")
		keys, values := a.modules.KeyValues()
		modState := values[moduleIdx]
		objectIdx := rapid.IntRange(0, modState.Objects.Len()-1).Draw(t, "objectIdx")
		objState := modState.Objects.Values()[objectIdx]
		update := objState.UpdateGen.Draw(t, "update")
		require.NoError(t, objState.ObjectType.ValidateObjectUpdate(update))
		require.NoError(t, a.applyUpdate(keys[moduleIdx], update, objState.ObjectType.RetainDeletions))
	}

	if f := a.options.Listener.Commit; f != nil {
		err := f()
		require.NoError(t, err)
	}
}

func (a *AppSimulator) newBlockFromSeed(seed int) {
	rapid.Custom[any](func(t *rapid.T) any {
		a.actionNewBlock(t)
		return nil
	}).Example(seed)
}

func (a *AppSimulator) applyUpdate(module string, update indexerbase.ObjectUpdate, retainDeletions bool) error {
	if a.options.Listener.OnObjectUpdate != nil {
		err := a.options.Listener.OnObjectUpdate(module, update)
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
	ModuleSchema indexerbase.ModuleSchema
	Objects      *btree.Map[string, *objectState]
}

type objectState struct {
	ObjectType indexerbase.ObjectType
	Objects    *btree.Map[string, indexerbase.ObjectUpdate]
	UpdateGen  *rapid.Generator[indexerbase.ObjectUpdate]
}
