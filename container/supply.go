package container

import (
	"reflect"
)

type supplyResolver struct {
	typ   reflect.Type
	value reflect.Value
	loc   Location
}

func (s supplyResolver) addNode(provider *simpleProvider, _ int, _ *container) error {
	return duplicateConstructorError(provider.ctr.Location, s.typ)
}

func (s supplyResolver) resolve(c *container, _ Scope, caller Location) (reflect.Value, error) {
	c.logf("Supplying %v from %s to %s", s.typ, s.loc, caller.Name())
	return s.value, nil
}
