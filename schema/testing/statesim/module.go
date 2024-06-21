package statesim

import (
	"fmt"

	"github.com/stretchr/testify/require"
	"github.com/tidwall/btree"
	"pgregory.net/rapid"

	"cosmossdk.io/schema"
)

type Module struct {
	moduleSchema      schema.ModuleSchema
	objectCollections *btree.Map[string, *ObjectCollection]
	updateGen         *rapid.Generator[schema.ObjectUpdate]
}

func NewModule(moduleSchema schema.ModuleSchema) *Module {
	objectCollections := &btree.Map[string, *ObjectCollection]{}
	var objectTypeNames []string
	for _, objectType := range moduleSchema.ObjectTypes {
		objectCollection := NewObjectCollection(objectType)
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

	return &Module{
		moduleSchema:      moduleSchema,
		updateGen:         updateGen,
		objectCollections: objectCollections,
	}
}

func (o *Module) ApplyUpdate(update schema.ObjectUpdate) error {
	objState, ok := o.objectCollections.Get(update.TypeName)
	if !ok {
		return fmt.Errorf("object type %s not found in module", update.TypeName)
	}

	return objState.ApplyUpdate(update)
}

func (o *Module) UpdateGen() *rapid.Generator[schema.ObjectUpdate] {
	return o.updateGen
}

func (o *Module) GetObjectCollection(objectType string) (*ObjectCollection, bool) {
	return o.objectCollections.Get(objectType)
}

func (o *Module) ScanState(f func(schema.ObjectUpdate) bool) {
	o.objectCollections.Scan(func(key string, value *ObjectCollection) bool {
		keepGoing := true
		value.ScanState(func(update schema.ObjectUpdate) bool {
			if !f(update) {
				keepGoing = false
				return false
			}
			return true
		})
		return keepGoing
	})
}
