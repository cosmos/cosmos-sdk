package container

import (
	"fmt"
	"reflect"

	"github.com/goccy/go-graphviz/cgraph"

	"github.com/pkg/errors"
)

// AutoGroupType marks a type which automatically gets grouped together. For an AutoGroupType T,
// T and []T can be declared as output parameters for constructors as many times within the container
// as desired. All of the provided values for T can be retrieved by declaring an
// []T input parameter.
type AutoGroupType interface {
	// IsAutoGroupType is a marker function which just indicates that this is a auto-group type.
	IsAutoGroupType()
}

var autoGroupTypeType = reflect.TypeOf((*AutoGroupType)(nil)).Elem()

func isAutoGroupType(t reflect.Type) bool {
	return t.Implements(autoGroupTypeType)
}

func isAutoGroupSliceType(typ reflect.Type) bool {
	return typ.Kind() == reflect.Slice && isAutoGroupType(typ.Elem())
}

type groupResolver struct {
	typ          reflect.Type
	sliceType    reflect.Type
	idxsInValues []int
	providers    []*simpleProvider
	resolved     bool
	values       reflect.Value
	graphNode    *cgraph.Node
}

type sliceGroupResolver struct {
	*groupResolver
}

func (g *groupResolver) describeLocation() string {
	return fmt.Sprintf("auto-group type %v", g.typ)
}

func (g *sliceGroupResolver) resolve(c *container, _ Scope, caller Location) (reflect.Value, error) {
	// Log
	c.logf("Providing auto-group type slice %v to %s from:", g.sliceType, caller.Name())
	c.indentLogger()
	for _, node := range g.providers {
		c.logf(node.provider.Location.String())
	}
	c.dedentLogger()

	// Resolve
	if !g.resolved {
		res := reflect.MakeSlice(g.sliceType, 0, 0)
		for i, node := range g.providers {
			values, err := node.resolveValues(c)
			if err != nil {
				return reflect.Value{}, err
			}
			value := values[g.idxsInValues[i]]
			if value.Kind() == reflect.Slice {
				n := value.Len()
				for j := 0; j < n; j++ {
					res = reflect.Append(res, value.Index(j))
				}
			} else {
				res = reflect.Append(res, value)
			}
		}
		g.values = res
		g.resolved = true
	}

	return g.values, nil
}

func (g *groupResolver) resolve(_ *container, _ Scope, _ Location) (reflect.Value, error) {
	return reflect.Value{}, errors.Errorf("%v is an auto-group type and cannot be used as an input value, instead use %v", g.typ, g.sliceType)
}

func (g *groupResolver) addNode(n *simpleProvider, i int) error {
	g.providers = append(g.providers, n)
	g.idxsInValues = append(g.idxsInValues, i)
	return nil
}

func (g groupResolver) typeGraphNode() *cgraph.Node {
	return g.graphNode
}
