package indexerbase

import "fmt"

// ModuleSchema represents the logical schema of a module for purposes of indexing and querying.
type ModuleSchema struct {
	// ObjectTypes describe the types of objects that are part of the module's schema.
	ObjectTypes map[string]ObjectType
}

func (m ModuleSchema) ValidateObjectUpdate(update ObjectUpdate) error {
	objectType, ok := m.ObjectTypes[update.TypeName]
	if !ok {
		return fmt.Errorf("unknown object type %q", update.TypeName)
	}

	if err := objectType.ValidateKey(update.Key); err != nil {
		return fmt.Errorf("invalid key for object type %q: %w", update.TypeName, err)
	}

	if update.Delete {
		return nil
	}

	return objectType.ValidateValue(update.Value)
}
