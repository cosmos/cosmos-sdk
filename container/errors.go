package container

import (
	"reflect"

	"github.com/pkg/errors"

	reflect2 "github.com/cosmos/cosmos-sdk/container/reflect"
)

func duplicateConstructorError(loc reflect2.Location, typ reflect.Type) error {
	return errors.Errorf("Duplicate constructor for type %v: %s", typ, loc)
}
