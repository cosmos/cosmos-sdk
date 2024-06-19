package state

import (
	"fmt"

	"github.com/stretchr/testify/require"
	"github.com/tidwall/btree"
	"pgregory.net/rapid"

	indexerbase "cosmossdk.io/indexer/base"
)

type Entry struct {
	Key   any
	Value any
}

type AppStateMachine struct {
	CheckFn           func(*rapid.T)
	MaxUpdatePerBlock int
	Modules           *btree.Map[string, *ModuleState]
	Listener          indexerbase.Listener
}

type ModuleState struct {
	ModuleSchema indexerbase.ModuleSchema
	Objects      *btree.Map[string, *ObjectState]
}

type ObjectState struct {
	ObjectType indexerbase.ObjectType
	Objects    *btree.Map[string, *Entry]
	UpdateGen  *rapid.Generator[indexerbase.ObjectUpdate]
}

func (s *AppStateMachine) Check(t *rapid.T) {
	if s.CheckFn != nil {
		s.CheckFn(t)
	}
}

func (s *AppStateMachine) ActionNewBlock(t *rapid.T) {
	maxUpdates := s.MaxUpdatePerBlock
	if maxUpdates <= 0 {
		maxUpdates = 100
	}
	numUpdates := rapid.IntRange(1, s.MaxUpdatePerBlock).Draw(t, "numUpdates")
	for i := 0; i < numUpdates; i++ {
		moduleIdx := rapid.IntRange(0, s.Modules.Len()).Draw(t, "moduleIdx")
		keys, values := s.Modules.KeyValues()
		modState := values[moduleIdx]
		objectIdx := rapid.IntRange(0, modState.Objects.Len()).Draw(t, "objectIdx")
		objState := modState.Objects.Values()[objectIdx]
		update := objState.UpdateGen.Draw(t, "update")
		require.NoError(t, objState.ObjectType.ValidateObjectUpdate(update))
		require.NoError(t, s.ApplyUpdate(keys[moduleIdx], update))
	}
}

func (s *AppStateMachine) NewBlockFromSeed(seed int) {
	rapid.Custom[any](func(t *rapid.T) any {
		s.ActionNewBlock(t)
		return nil
	}).Example(seed)
}

func (s *AppStateMachine) ApplyUpdate(module string, update indexerbase.ObjectUpdate) error {
	modState, ok := s.Modules.Get(module)
	if !ok {
		return fmt.Errorf("module %v not found", module)
	}

	objState, ok := modState.Objects.Get(update.TypeName)
	if !ok {
		return fmt.Errorf("object type %v not found in module %v", update.TypeName, module)
	}

	keyStr := fmt.Sprintf("%v", update.Key)
	if update.Delete {
		objState.Objects.Delete(keyStr)
	} else {
		objState.Objects.Set(fmt.Sprintf("%v", update.Key), &Entry{Key: update.Key, Value: update.Value})
	}

	if s.Listener.OnObjectUpdate != nil {
		err := s.Listener.OnObjectUpdate(module, update)
		return err
	}
	return nil
}
