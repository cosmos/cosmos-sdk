package statesim

import (
	"fmt"

	"github.com/tidwall/btree"
	"pgregory.net/rapid"

	"cosmossdk.io/schema"
	schematesting "cosmossdk.io/schema/testing"
)

type ObjectCollection struct {
	objectType schema.ObjectType
	objects    *btree.Map[string, schema.ObjectUpdate]
	updateGen  *rapid.Generator[schema.ObjectUpdate]
}

func NewObjectCollection(objectType schema.ObjectType) *ObjectCollection {
	objects := &btree.Map[string, schema.ObjectUpdate]{}
	updateGen := schematesting.ObjectUpdateGen(objectType, objects)
	return &ObjectCollection{
		objectType: objectType,
		objects:    objects,
		updateGen:  updateGen,
	}
}

func (o *ObjectCollection) ApplyUpdate(update schema.ObjectUpdate) error {
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

func (o *ObjectCollection) UpdateGen() *rapid.Generator[schema.ObjectUpdate] {
	return o.updateGen
}
