package container

import (
	"reflect"

	"github.com/pkg/errors"
)

func duplicateConstructorError(loc Location, typ reflect.Type) error {
	return errors.Errorf("Duplicate constructor for type %v: %s", typ, loc)
}
