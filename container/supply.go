package container

import (
	"reflect"

	"github.com/goccy/go-graphviz/cgraph"
)

type supplyResolver struct {
	typ       reflect.Type
	value     reflect.Value
	loc       Location
	graphNode *cgraph.Node
}

func (s supplyResolver) describeLocation() string {
	return s.loc.String()
}

func (s supplyResolver) addNode(provider *simpleProvider, _ int) error {
	return duplicateDefinitionError(s.typ, provider.provider.Location, s.loc.String())
}

func (s supplyResolver) resolve(c *container, _ Scope, caller Location) (reflect.Value, error) {
	c.logf("Supplying %v from %s to %s", s.typ, s.loc, caller.Name())
	return s.value, nil
}

func (s supplyResolver) typeGraphNode() *cgraph.Node {
	return s.graphNode
}
