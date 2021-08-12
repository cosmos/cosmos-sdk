package container

import (
	"reflect"
)

type scopeDepProvider struct {
	ctr            *ProviderDescriptor
	calledForScope map[Scope]bool
	valueMap       map[Scope][]reflect.Value
}

type scopeDepResolver struct {
	typ         reflect.Type
	idxInValues int
	node        *scopeDepProvider
	valueMap    map[Scope]reflect.Value
}

func (s scopeDepResolver) describeLocation() string {
	return s.node.ctr.Location.String()
}

func (s scopeDepResolver) resolve(ctr *container, scope Scope, caller Location) (reflect.Value, error) {
	// Log
	ctr.logf("Providing %v from %s to %s", s.typ, s.node.ctr.Location, caller.Name())

	// Resolve
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

func (s scopeDepResolver) addNode(p *simpleProvider, _ int, _ *container) error {
	return duplicateDefinitionError(s.typ, p.ctr.Location, s.node.ctr.Location.String())
}
