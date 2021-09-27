package container

import (
	"fmt"
	"reflect"

	"github.com/goccy/go-graphviz/cgraph"

	"github.com/pkg/errors"
)

// OnePerScopeType marks a type which
// can have up to one value per scope. All of the values for a one-per-scope type T
// and their respective scopes, can be retrieved by declaring an input parameter map[string]T.
type OnePerScopeType interface {
	// IsOnePerScopeType is a marker function just indicates that this is a one-per-scope type.
	IsOnePerScopeType()
}

var onePerScopeTypeType = reflect.TypeOf((*OnePerScopeType)(nil)).Elem()

func isOnePerScopeType(t reflect.Type) bool {
	return t.Implements(onePerScopeTypeType)
}

func isOnePerScopeMapType(typ reflect.Type) bool {
	return typ.Kind() == reflect.Map && isOnePerScopeType(typ.Elem()) && typ.Key().Kind() == reflect.String
}

type onePerScopeResolver struct {
	typ       reflect.Type
	mapType   reflect.Type
	providers map[Scope]*simpleProvider
	idxMap    map[Scope]int
	resolved  bool
	values    reflect.Value
	graphNode *cgraph.Node
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
	c.logf("Providing one-per-scope type map %v to %s from:", o.mapType, caller.Name())
	c.indentLogger()
	for scope, node := range o.providers {
		c.logf("%s: %s", scope.Name(), node.provider.Location)
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

func (o *onePerScopeResolver) addNode(n *simpleProvider, i int) error {
	if n.scope == nil {
		return errors.Errorf("cannot define a constructor with one-per-scope dependency %v which isn't provided in a scope", o.typ)
	}

	if existing, ok := o.providers[n.scope]; ok {
		return errors.Errorf("duplicate provision for one-per-scope type %v in scope %s: %s\n\talready provided by %s",
			o.typ, n.scope.Name(), n.provider.Location, existing.provider.Location)
	}

	o.providers[n.scope] = n
	o.idxMap[n.scope] = i

	return nil
}

func (o *mapOfOnePerScopeResolver) addNode(s *simpleProvider, _ int) error {
	return errors.Errorf("%v is a one-per-scope type and thus %v can't be used as an output parameter in %s", o.typ, o.mapType, s.provider.Location)
}

func (o onePerScopeResolver) typeGraphNode() *cgraph.Node {
	return o.graphNode
}
