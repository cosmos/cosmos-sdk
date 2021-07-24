package container

import (
	"fmt"
	"reflect"

	"github.com/awalterschulze/gographviz"

	containerreflect "github.com/cosmos/cosmos-sdk/container/reflect"
)

/*
TODO:
error traces
circular dependencies
StructArgs
*/

type container struct {
	*config

	resolvers map[reflect.Type]resolver
	graph     *gographviz.Graph
}

func newContainer(cfg *config) *container {
	graph := gographviz.NewGraph()
	err := graph.SetName("G")
	if err != nil {
		panic(err)
	}

	ctr := &container{
		config:    cfg,
		resolvers: map[reflect.Type]resolver{},
		graph:     graph,
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
			typ:     typ,
			mapType: mapType,
			nodes:   map[Scope]*simpleProvider{},
			idxMap:  map[Scope]int{},
		}
		ctr.resolvers[typ] = r
		ctr.resolvers[mapType] = &mapOfOnePerScopeResolver{r}
	}

	return ctr
}

func (c *container) call(constructor *containerreflect.Constructor, scope Scope) ([]reflect.Value, error) {
	c.logf("Resolving dependencies for %s", constructor.Location)
	c.indentLogger()
	inVals := make([]reflect.Value, len(constructor.In))
	for i, in := range constructor.In {
		val, err := c.resolve(in, scope, constructor.Location)
		if err != nil {
			return nil, err
		}
		inVals[i] = val
	}
	c.dedentLogger()
	c.logf("Calling %s", constructor.Location)
	return constructor.Fn(inVals), nil
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

		for i, out := range constructor.Out {
			typ := out.Type
			// auto-group slices of auto-group types
			if typ.Kind() == reflect.Slice && c.autoGroupTypes[typ.Elem()] {
				typ = typ.Elem()
			}

			vr, ok := c.resolvers[typ]
			if ok {
				err := vr.addNode(node, i)
				if err != nil {
					return nil, err
				}
			} else {
				c.resolvers[typ] = &simpleResolver{
					node: node,
					typ:  typ,
				}
			}
		}

		return node, nil
	} else {
		if !noLog {
			c.logf("Registering scope provider: %s", constructor.Location.String())
		}
		node := &scopeProvider{
			ctr:            constructor,
			calledForScope: map[Scope]bool{},
			valueMap:       map[Scope][]reflect.Value{},
		}

		for i, out := range constructor.Out {
			typ := out.Type
			_, ok := c.resolvers[typ]
			if ok {
				return nil, fmt.Errorf("duplicate constructor for type %v", typ)
			}
			c.resolvers[typ] = &scopeProviderResolver{
				typ:         typ,
				idxInValues: i,
				node:        node,
				valueMap:    map[Scope]reflect.Value{},
			}
		}

		return node, nil
	}
}

func (c *container) addGraphNode(loc containerreflect.Location) error {
	str := loc.String()
	if !c.graph.IsNode(str) {
		return c.graph.AddNode("G", str, nil)
	}
	return nil
}

func (c *container) addGraphEdge(from containerreflect.Location, to containerreflect.Location) error {
	err := c.addGraphNode(from)
	if err != nil {
		return err
	}

	err = c.addGraphNode(to)
	if err != nil {
		return err
	}

	return c.graph.AddEdge(from.String(), to.String(), true, nil)
}

func (c *container) resolve(in containerreflect.Input, scope Scope, caller containerreflect.Location) (reflect.Value, error) {
	if in.Type == scopeType {
		if scope == nil {
			return reflect.Value{}, fmt.Errorf("expected scope but got nil")
		}
		c.logf("Providing Scope %s", scope.Name())
		return reflect.ValueOf(scope), nil
	}

	vr, ok := c.resolvers[in.Type]
	if !ok {
		if in.Optional {
			c.logf("Providing zero value for optional dependency %v", in.Type)
			return reflect.Zero(in.Type), nil
		}

		return reflect.Value{}, fmt.Errorf("no constructor for type %v", in.Type)
	}

	return vr.resolve(c, scope, caller)
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
	c.logf("Done")
	c.logf("Graph: %s", c.graph.String())
	return nil
}
