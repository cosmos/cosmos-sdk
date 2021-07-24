package container

import (
	"fmt"
	"reflect"

	containerreflect "github.com/cosmos/cosmos-sdk/container/reflect"
)

type scopeProvider struct {
	ctr            *containerreflect.Constructor
	calledForScope map[Scope]bool
	valueMap       map[Scope][]reflect.Value
}

type scopeProviderResolver struct {
	typ         reflect.Type
	idxInValues int
	node        *scopeProvider
	valueMap    map[Scope]reflect.Value
}

func (s scopeProviderResolver) resolve(ctr *container, scope Scope, resolver containerreflect.Location) (reflect.Value, error) {
	ctr.logf("Providing %v from %s to %s", s.typ, s.node.ctr.Location, resolver.Name())
	err := ctr.addGraphEdge(s.node.ctr.Location, resolver)
	if err != nil {
		return reflect.Value{}, err
	}

	if val, ok := s.valueMap[scope]; ok {
		return val, nil
	}

	if !s.node.calledForScope[scope] {
		values, err := ctr.call(s.node.ctr, scope)
		if err != nil {
			return reflect.Value{}, err
		}

		s.node.valueMap[scope] = values
		s.node.calledForScope[scope] = true
	}

	value := s.node.valueMap[scope][s.idxInValues]
	s.valueMap[scope] = value
	return value, nil
}

func (s scopeProviderResolver) addNode(*simpleProvider, int) error {
	return fmt.Errorf("duplicate constructor for type %v", s.typ)
}
