package container

import (
	"fmt"
	"reflect"

	containerreflect "github.com/cosmos/cosmos-sdk/container/reflect"
)

type scopeDepProvider struct {
	ctr            *containerreflect.Constructor
	calledForScope map[Scope]bool
	valueMap       map[Scope][]reflect.Value
}

type scopeDepResolver struct {
	typ         reflect.Type
	idxInValues int
	node        *scopeDepProvider
	valueMap    map[Scope]reflect.Value
}

func (s scopeDepResolver) resolve(ctr *container, scope Scope, caller containerreflect.Location) (reflect.Value, error) {
	// Log
	ctr.logf("Providing %v from %s to %s", s.typ, s.node.ctr.Location, caller.Name())

	// Graph
	typeGraphNode, err := ctr.typeGraphNode(s.typ)
	markGraphNodeAsUsed(typeGraphNode)
	if err != nil {
		return reflect.Value{}, err
	}

	to, err := ctr.locationGraphNode(caller)
	if err != nil {
		return reflect.Value{}, err
	}

	err = ctr.addGraphEdge(typeGraphNode, to, s.typ.Name())
	if err != nil {
		return reflect.Value{}, err
	}

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

func (s scopeDepResolver) addNode(*simpleProvider, int, *container) error {
	return fmt.Errorf("duplicate constructor for type %v", s.typ)
}
