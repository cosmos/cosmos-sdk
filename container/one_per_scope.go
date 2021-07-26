package container

import (
	"fmt"
	"reflect"

	"github.com/pkg/errors"
)

type onePerScopeResolver struct {
	typ       reflect.Type
	mapType   reflect.Type
	providers map[Scope]*simpleProvider
	idxMap    map[Scope]int
	resolved  bool
	values    reflect.Value
}

type mapOfOnePerScopeResolver struct {
	*onePerScopeResolver
}

func (o *onePerScopeResolver) resolve(_ *container, _ Scope, _ Location) (reflect.Value, error) {
	return reflect.Value{}, errors.Errorf("%v is a one-per-scope type and thus can't be used as an input parameter, instead use %v", o.typ, o.mapType)
}

func (o *onePerScopeResolver) describeLocation() string {
	return fmt.Sprintf("one-per-scope type %v", o.typ)
}

func (o *mapOfOnePerScopeResolver) resolve(c *container, _ Scope, caller Location) (reflect.Value, error) {
	// Log
	c.logf("Providing %v to %s from:", o.mapType, caller.Name())
	c.indentLogger()
	for scope, node := range o.providers {
		c.logf("%s: %s", scope.Name(), node.ctr.Location)
	}
	c.dedentLogger()

	// Resolve
	if !o.resolved {
		res := reflect.MakeMap(o.mapType)
		for scope, node := range o.providers {
			values, err := node.resolveValues(c)
			if err != nil {
				return reflect.Value{}, err
			}
			idx := o.idxMap[scope]
			if len(values) < idx {
				return reflect.Value{}, errors.Errorf("expected value of type %T at index %d", o.typ, idx)
			}
			value := values[idx]
			res.SetMapIndex(reflect.ValueOf(scope.Name()), value)
		}

		o.values = res
		o.resolved = true
	}

	return o.values, nil
}

func (o *onePerScopeResolver) addNode(n *simpleProvider, i int, c *container) error {
	if n.scope == nil {
		return errors.Errorf("cannot define a constructor with one-per-scope dependency %v which isn't provided in a scope", o.typ)
	}

	if _, ok := o.providers[n.scope]; ok {
		return errors.Errorf("Duplicate constructor for one-per-scope type %v in scope %s: %s",
			o.typ, n.scope.Name(), n.ctr.Location)
	}

	o.providers[n.scope] = n
	o.idxMap[n.scope] = i

	constructorGraphNode, err := c.locationGraphNode(n.ctr.Location, n.scope)
	if err != nil {
		return err
	}

	typeGraphNode, err := c.typeGraphNode(o.mapType)
	if err != nil {
		return err
	}

	c.addGraphEdge(constructorGraphNode, typeGraphNode)
	return nil
}

func (o *mapOfOnePerScopeResolver) addNode(*simpleProvider, int, *container) error {
	return errors.Errorf("%v is a one-per-scope type and thus %v can't be used as an output parameter", o.typ, o.mapType)
}
