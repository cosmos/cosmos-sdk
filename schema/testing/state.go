package schematesting

import (
	"fmt"

	"github.com/stretchr/testify/require"
	"github.com/tidwall/btree"
	"pgregory.net/rapid"

	"cosmossdk.io/schema"
	"cosmossdk.io/schema/listener"
)

type TestAppState struct {
	moduleStates *btree.Map[string, *TestModuleState]
	updateGen    *rapid.Generator[listener.ObjectUpdateData]
}

func NewTestAppState(appSchema map[string]schema.ModuleSchema) *TestAppState {
	moduleStates := &btree.Map[string, *TestModuleState]{}
	var moduleNames []string

	for moduleName, moduleSchema := range appSchema {
		moduleState := NewTestModuleState(moduleSchema)
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

	return &TestAppState{
		moduleStates: moduleStates,
		updateGen:    updateGen,
	}
}

func (s *TestAppState) ApplyUpdate(moduleName string, update schema.ObjectUpdate) error {
	moduleState, ok := s.moduleStates.Get(moduleName)
	if !ok {
		return fmt.Errorf("module %s not found", moduleName)
	}

	return moduleState.ApplyUpdate(update)
}

func (o *TestAppState) UpdateGen() *rapid.Generator[listener.ObjectUpdateData] {
	return o.updateGen
}

type TestModuleState struct {
	moduleSchema      schema.ModuleSchema
	objectCollections *btree.Map[string, *TestObjectCollectionState]
	updateGen         *rapid.Generator[schema.ObjectUpdate]
}

func NewTestModuleState(moduleSchema schema.ModuleSchema) *TestModuleState {
	objectCollections := &btree.Map[string, *TestObjectCollectionState]{}
	var objectTypeNames []string
	for _, objectType := range moduleSchema.ObjectTypes {
		objectCollection := NewTestObjectCollectionState(objectType)
		objectCollections.Set(objectType.Name, objectCollection)
		objectTypeNames = append(objectTypeNames, objectType.Name)
	}

	objectTypeSelector := rapid.Map(rapid.IntRange(0, len(objectTypeNames)), func(u int) string {
		return objectTypeNames[u]
	})

	updateGen := rapid.Custom(func(t *rapid.T) schema.ObjectUpdate {
		objectType := objectTypeSelector.Draw(t, "objectType")
		objectColl, ok := objectCollections.Get(objectType)
		require.True(t, ok)
		return objectColl.UpdateGen().Draw(t, "update")
	})

	return &TestModuleState{
		moduleSchema: moduleSchema,
		updateGen:    updateGen,
	}
}

func (o *TestModuleState) ApplyUpdate(update schema.ObjectUpdate) error {
	objState, ok := o.objectCollections.Get(update.TypeName)
	if !ok {
		return fmt.Errorf("object type %s not found in module", update.TypeName)
	}

	return objState.ApplyUpdate(update)
}

func (o *TestModuleState) UpdateGen() *rapid.Generator[schema.ObjectUpdate] {
	return o.updateGen
}

type TestObjectCollectionState struct {
	objectType schema.ObjectType
	objects    *btree.Map[string, schema.ObjectUpdate]
	updateGen  *rapid.Generator[schema.ObjectUpdate]
}

func NewTestObjectCollectionState(objectType schema.ObjectType) *TestObjectCollectionState {
	objects := &btree.Map[string, schema.ObjectUpdate]{}
	updateGen := ObjectUpdateGen(objectType, objects)
	return &TestObjectCollectionState{
		objectType: objectType,
		objects:    objects,
		updateGen:  updateGen,
	}
}

func (o *TestObjectCollectionState) ApplyUpdate(update schema.ObjectUpdate) error {
	if update.TypeName != o.objectType.Name {
		return fmt.Errorf("update type name %q does not match object type name %q", update.TypeName, o.objectType.Name)
	}

	err := o.objectType.ValidateObjectUpdate(update)
	if err != nil {
		return err
	}

	keyStr := fmt.Sprintf("%v", update.Key)
	if update.Delete {
		if o.objectType.RetainDeletions {
			cur, ok := o.objects.Get(keyStr)
			if !ok {
				return fmt.Errorf("object not found for deletion: %v", update.Key)
			}

			cur.Delete = true
			o.objects.Set(keyStr, cur)
		} else {
			o.objects.Delete(keyStr)
		}
	} else {
		o.objects.Set(keyStr, update)
	}

	return nil
}

func (o *TestObjectCollectionState) UpdateGen() *rapid.Generator[schema.ObjectUpdate] {
	return o.updateGen
}
