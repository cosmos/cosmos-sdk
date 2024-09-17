package statesim

import (
	"fmt"

	"github.com/stretchr/testify/require"
	"github.com/tidwall/btree"
	"pgregory.net/rapid"

	"cosmossdk.io/schema"
	"cosmossdk.io/schema/view"
)

// Module is a collection of object collections corresponding to a module's schema for testing purposes.
type Module struct {
	name              string
	moduleSchema      schema.ModuleSchema
	objectCollections *btree.Map[string, *ObjectCollection]
	updateGen         *rapid.Generator[schema.StateObjectUpdate]
}

// NewModule creates a new Module for the given module schema.
func NewModule(name string, moduleSchema schema.ModuleSchema, options Options) *Module {
	objectCollections := &btree.Map[string, *ObjectCollection]{}
	var objectTypeNames []string

	moduleSchema.StateObjectTypes(func(objectType schema.StateObjectType) bool {
		objectCollection := NewObjectCollection(objectType, options, moduleSchema)
		objectCollections.Set(objectType.Name, objectCollection)
		objectTypeNames = append(objectTypeNames, objectType.Name)
		return true
	})

	objectTypeSelector := rapid.SampledFrom(objectTypeNames)

	updateGen := rapid.Custom(func(t *rapid.T) schema.StateObjectUpdate {
		objectType := objectTypeSelector.Draw(t, "objectType")
		objectColl, ok := objectCollections.Get(objectType)
		require.True(t, ok)
		return objectColl.UpdateGen().Draw(t, "update")
	})

	return &Module{
		name:              name,
		moduleSchema:      moduleSchema,
		updateGen:         updateGen,
		objectCollections: objectCollections,
	}
}

// ApplyUpdate applies the given object update to the module.
func (o *Module) ApplyUpdate(update schema.StateObjectUpdate) error {
	objState, ok := o.objectCollections.Get(update.TypeName)
	if !ok {
		return fmt.Errorf("object type %s not found in module", update.TypeName)
	}

	return objState.ApplyUpdate(update)
}

// UpdateGen returns a generator for object updates. The generator is stateful and returns
// a certain number of updates and deletes of existing objects in the module.
func (o *Module) UpdateGen() *rapid.Generator[schema.StateObjectUpdate] {
	return o.updateGen
}

// ModuleName returns the name of the module.
func (o *Module) ModuleName() string {
	return o.name
}

// ModuleSchema returns the module schema for the module.
func (o *Module) ModuleSchema() schema.ModuleSchema {
	return o.moduleSchema
}

// GetObjectCollection returns the object collection for the given object type.
func (o *Module) GetObjectCollection(objectType string) (view.ObjectCollection, error) {
	obj, ok := o.objectCollections.Get(objectType)
	if !ok {
		return nil, nil
	}
	return obj, nil
}

// ObjectCollections iterates over all object collections in the module.
func (o *Module) ObjectCollections(f func(value view.ObjectCollection, err error) bool) {
	o.objectCollections.Scan(func(key string, value *ObjectCollection) bool {
		return f(value, nil)
	})
}

// NumObjectCollections returns the number of object collections in the module.
func (o *Module) NumObjectCollections() (int, error) {
	return o.objectCollections.Len(), nil
}
