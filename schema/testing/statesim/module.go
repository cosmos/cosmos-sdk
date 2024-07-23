package statesim

import (
	"fmt"

	"github.com/stretchr/testify/require"
	"github.com/tidwall/btree"
	"pgregory.net/rapid"

	"cosmossdk.io/schema"
)

// Module is a collection of object collections corresponding to a module's schema for testing purposes.
type Module struct {
	moduleSchema      schema.ModuleSchema
	objectCollections *btree.Map[string, *ObjectCollection]
	updateGen         *rapid.Generator[schema.ObjectUpdate]
}

// NewModule creates a new Module for the given module schema.
func NewModule(moduleSchema schema.ModuleSchema, options Options) *Module {
	objectCollections := &btree.Map[string, *ObjectCollection]{}
	var objectTypeNames []string
	for typ := range moduleSchema.Types {
		objectType, ok := typ.(schema.ObjectType)
		if !ok {
			continue
		}
		objectCollection := NewObjectCollection(objectType, options)
		objectCollections.Set(objectType.Name, objectCollection)
		objectTypeNames = append(objectTypeNames, objectType.Name)
	}

	objectTypeSelector := rapid.SampledFrom(objectTypeNames)

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

// ApplyUpdate applies the given object update to the module.
func (o *Module) ApplyUpdate(update schema.ObjectUpdate) error {
	objState, ok := o.objectCollections.Get(update.TypeName)
	if !ok {
		return fmt.Errorf("object type %s not found in module", update.TypeName)
	}

	return objState.ApplyUpdate(update)
}

// UpdateGen returns a generator for object updates. The generator is stateful and returns
// a certain number of updates and deletes of existing objects in the module.
func (o *Module) UpdateGen() *rapid.Generator[schema.ObjectUpdate] {
	return o.updateGen
}

// ModuleSchema returns the module schema for the module.
func (o *Module) ModuleSchema() schema.ModuleSchema {
	return o.moduleSchema
}

// GetObjectCollection returns the object collection for the given object type.
func (o *Module) GetObjectCollection(objectType string) (*ObjectCollection, bool) {
	return o.objectCollections.Get(objectType)
}

// ObjectCollections iterates over all object collections in the module.
func (o *Module) ObjectCollections(f func(value *ObjectCollection) bool) {
	o.objectCollections.Scan(func(key string, value *ObjectCollection) bool {
		return f(value)
	})
}
