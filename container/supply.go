package container

import (
	"reflect"

	reflect2 "github.com/cosmos/cosmos-sdk/container/reflect"
)

type supplyResolver struct {
	typ   reflect.Type
	value reflect.Value
	loc   reflect2.Location
}

func (s supplyResolver) addNode(provider *simpleProvider, _ int, _ *container) error {
	return duplicateConstructorError(provider.ctr.Location, s.typ)
}

func (s supplyResolver) resolve(c *container, s2 Scope, caller reflect2.Location) (reflect.Value, error) {
	c.logf("Providing %v from %s to %s", s.typ, s.loc, caller.Name())
	return s.value, nil
}
