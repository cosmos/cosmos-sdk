package container

import (
	"fmt"
	"reflect"

	containerreflect "github.com/cosmos/cosmos-sdk/container/reflect"
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

func (o *onePerScopeResolver) resolve(_ *container, _ Scope, _ containerreflect.Location) (reflect.Value, error) {
	return reflect.Value{}, fmt.Errorf("%v is a one-per-scope type and thus can't be used as an input parameter, instead use %v", o.typ, o.mapType)
}

func (o *mapOfOnePerScopeResolver) resolve(c *container, _ Scope, resolver containerreflect.Location) (reflect.Value, error) {
	c.logf("Providing %v to %s from:", o.mapType, resolver.Name())
	c.indentLogger()
	for scope, node := range o.providers {
		c.logf("%s: %s", scope.Name(), node.ctr.Location)
		err := c.addGraphEdge(node.ctr.Location, resolver)
		if err != nil {
			return reflect.Value{}, err
		}
	}
	c.dedentLogger()
	if !o.resolved {
		res := reflect.MakeMap(o.mapType)
		for scope, node := range o.providers {
			values, err := node.resolveValues(c)
			if err != nil {
				return reflect.Value{}, err
			}
			idx := o.idxMap[scope]
			if len(values) < idx {
				return reflect.Value{}, fmt.Errorf("expected value of type %T at index %d", o.typ, idx)
			}
			value := values[idx]
			res.SetMapIndex(reflect.ValueOf(scope), value)
		}

		o.values = res
		o.resolved = true
	}

	return o.values, nil
}

func (o *onePerScopeResolver) addNode(n *simpleProvider, i int) error {
	if n.scope == nil {
		return fmt.Errorf("cannot define a constructor with one-per-scope dependency %v which isn't provided in a scope", o.typ)
	}

	if _, ok := o.providers[n.scope]; ok {
		return fmt.Errorf("duplicate constructor for one-per-scope type %v in scope %s", o.typ, n.scope)
	}

	o.providers[n.scope] = n
	o.idxMap[n.scope] = i

	return nil
}

func (o *mapOfOnePerScopeResolver) addNode(*simpleProvider, int) error {
	return fmt.Errorf("%v is a one-per-scope type and thus %v can't be used as an output parameter", o.typ, o.mapType)
}
