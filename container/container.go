package container

import (
	"fmt"
	"reflect"

	"github.com/goccy/go-graphviz/cgraph"

	containerreflect "github.com/cosmos/cosmos-sdk/container/reflect"
)

/*
TODO:
circular dependencies
error resolve traces
StructArgs
check all errors
show errors on graph - call, resolve
*/

type container struct {
	*config

	resolvers map[reflect.Type]resolver

	callerStack []containerreflect.Location
	callerMap   map[containerreflect.Location]bool
}

func newContainer(cfg *config) *container {
	ctr := &container{
		config:      cfg,
		resolvers:   map[reflect.Type]resolver{},
		callerStack: nil,
		callerMap:   map[containerreflect.Location]bool{},
	}

	for typ := range cfg.autoGroupTypes {
		sliceType := reflect.SliceOf(typ)
		r := &groupResolver{
			typ:       typ,
			sliceType: sliceType,
		}
		ctr.resolvers[typ] = r
		ctr.resolvers[sliceType] = &sliceGroupValueResolver{r}
	}

	for typ := range cfg.onePerScopeTypes {
		mapType := reflect.MapOf(scopeType, typ)
		r := &onePerScopeResolver{
			typ:       typ,
			mapType:   mapType,
			providers: map[Scope]*simpleProvider{},
			idxMap:    map[Scope]int{},
		}
		ctr.resolvers[typ] = r
		ctr.resolvers[mapType] = &mapOfOnePerScopeResolver{r}
	}

	return ctr
}

func (c *container) call(constructor *containerreflect.Constructor, scope Scope) ([]reflect.Value, error) {
	loc := constructor.Location
	graphNode, err := c.locationGraphNode(loc)
	if err != nil {
		return nil, err
	}
	c.logf("Marking %s as failed", loc.Name())
	markGraphNodeAsFailed(graphNode)

	if c.callerMap[loc] {
		return nil, fmt.Errorf("cyclic dependency: %s -> %s", loc.Name(), loc.Name())
	}

	c.callerMap[loc] = true
	c.callerStack = append(c.callerStack, loc)

	c.logf("Resolving dependencies for %s", loc)
	c.indentLogger()
	inVals := make([]reflect.Value, len(constructor.In))
	for i, in := range constructor.In {
		val, err := c.resolve(in, scope, loc)
		if err != nil {
			return nil, err
		}
		inVals[i] = val
	}
	c.dedentLogger()
	c.logf("Calling %s", loc)

	delete(c.callerMap, loc)
	c.callerStack = c.callerStack[0 : len(c.callerStack)-1]

	out := constructor.Fn(inVals)
	c.logf("Marking %s as used", loc.Name())
	markGraphNodeAsUsed(graphNode)

	return out, nil
}

func (c *container) addNode(constructor *containerreflect.Constructor, scope Scope, noLog bool) (interface{}, error) {
	hasScopeParam := len(constructor.In) > 0 && constructor.In[0].Type == scopeType
	if scope != nil || !hasScopeParam {
		if !noLog {
			c.logf("Registering provider: %s", constructor.Location.String())
		}
		node := &simpleProvider{
			ctr:   constructor,
			scope: scope,
		}

		constructorGraphNode, err := c.locationGraphNode(constructor.Location)
		if err != nil {
			return reflect.Value{}, err
		}

		for i, out := range constructor.Out {
			typ := out.Type
			// auto-group slices of auto-group types
			if typ.Kind() == reflect.Slice && c.autoGroupTypes[typ.Elem()] {
				typ = typ.Elem()
			}

			vr, ok := c.resolvers[typ]
			if ok {
				err := vr.addNode(node, i, c)
				if err != nil {
					return nil, err
				}
			} else {
				c.resolvers[typ] = &simpleResolver{
					node: node,
					typ:  typ,
				}

				typeGraphNode, err := c.typeGraphNode(typ)
				if err != nil {
					return reflect.Value{}, err
				}

				c.addGraphEdge(constructorGraphNode, typeGraphNode)
			}
		}

		return node, nil
	} else {
		if !noLog {
			c.logf("Registering scope provider: %s", constructor.Location.String())
		}
		node := &scopeDepProvider{
			ctr:            constructor,
			calledForScope: map[Scope]bool{},
			valueMap:       map[Scope][]reflect.Value{},
		}

		constructorGraphNode, err := c.locationGraphNode(constructor.Location)
		if err != nil {
			return reflect.Value{}, err
		}

		for i, out := range constructor.Out {
			typ := out.Type
			_, ok := c.resolvers[typ]
			if ok {
				return nil, fmt.Errorf("duplicate constructor for type %v", typ)
			}
			c.resolvers[typ] = &scopeDepResolver{
				typ:         typ,
				idxInValues: i,
				node:        node,
				valueMap:    map[Scope]reflect.Value{},
			}

			typeGraphNode, err := c.typeGraphNode(typ)
			if err != nil {
				return reflect.Value{}, err
			}

			c.addGraphEdge(constructorGraphNode, typeGraphNode)
		}

		return node, nil
	}
}

func (c *container) resolve(in containerreflect.Input, scope Scope, caller containerreflect.Location) (reflect.Value, error) {
	typeGraphNode, err := c.typeGraphNode(in.Type)
	if err != nil {
		return reflect.Value{}, err
	}

	to, err := c.locationGraphNode(caller)
	if err != nil {
		return reflect.Value{}, err
	}

	c.addGraphEdge(typeGraphNode, to)

	if in.Type == scopeType {
		if scope == nil {
			return reflect.Value{}, fmt.Errorf("expected scope but got nil")
		}
		c.logf("Providing Scope %s", scope.Name())
		markGraphNodeAsUsed(typeGraphNode)
		return reflect.ValueOf(scope), nil
	}

	vr, ok := c.resolvers[in.Type]
	if !ok {
		if in.Optional {
			c.logf("Providing zero value for optional dependency %v", in.Type)
			return reflect.Zero(in.Type), nil
		}

		markGraphNodeAsFailed(typeGraphNode)
		return reflect.Value{}, fmt.Errorf("no constructor for type %v", in.Type)
	}

	res, err := vr.resolve(c, scope, caller)
	if err != nil {
		markGraphNodeAsFailed(typeGraphNode)
		return reflect.Value{}, err
	}

	markGraphNodeAsUsed(typeGraphNode)
	return res, nil
}

func (c *container) run(invoker interface{}) error {
	rctr, err := makeReflectConstructor(invoker)
	if err != nil {
		return err
	}

	c.logf("Registering invoker %s", rctr.Location)

	node, err := c.addNode(rctr, nil, true)
	if err != nil {
		return err
	}

	sn, ok := node.(*simpleProvider)
	if !ok {
		return fmt.Errorf("cannot run scoped provider as an invoker")
	}

	c.logf("Building container")
	_, err = sn.resolveValues(c)
	if err != nil {
		return err
	}
	c.logf("Done building container")

	return nil
}

func (c *container) locationGraphNode(location containerreflect.Location) (*cgraph.Node, error) {
	node, found, err := c.findOrCreateGraphNode(location.Name())
	if err != nil {
		return nil, err
	}

	if found {
		return node, nil
	}

	node = node.SetShape(cgraph.BoxShape)
	node.SetColor("lightgrey")
	return node, nil
}

func (c *container) typeGraphNode(typ reflect.Type) (*cgraph.Node, error) {
	node, found, err := c.findOrCreateGraphNode(typ.String())
	if err != nil {
		return nil, err
	}

	if found {
		return node, nil
	}

	node.SetColor("lightgrey")
	return node, err
}

func (c *container) findOrCreateGraphNode(name string) (node *cgraph.Node, found bool, err error) {
	node, err = c.graph.Node(name)
	if err != nil {
		return nil, false, err
	}

	if node != nil {
		return node, true, nil
	}

	node, err = c.graph.CreateNode(name)
	return node, false, err
}

func (c *container) addGraphEdge(from *cgraph.Node, to *cgraph.Node) {
	_, err := c.graph.CreateEdge("", from, to)
	if err != nil {
		c.logf("error creating graph edge")
	}
}

func markGraphNodeAsUsed(node *cgraph.Node) {
	node.SetColor("black")
}

func markGraphNodeAsFailed(node *cgraph.Node) {
	node.SetColor("red")
}
