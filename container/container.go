package container

import (
	"fmt"
	"reflect"

	"github.com/goccy/go-graphviz/cgraph"
	"github.com/pkg/errors"

	containerreflect "github.com/cosmos/cosmos-sdk/container/reflect"
)

/*
TODO:
error resolve traces
StructArgs
review all errors return
error return args
scope in wrong position
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

	out, err := constructor.Fn(inVals)
	if err != nil {
		return nil, errors.Wrapf(err, "error calling constructor %s", loc)
	}
	markGraphNodeAsUsed(graphNode)

	return out, nil
}

func (c *container) addNode(constructor *containerreflect.Constructor, scope Scope, noLog bool) (interface{}, error) {
	constructorGraphNode, err := c.locationGraphNode(constructor.Location)
	if err != nil {
		return reflect.Value{}, err
	}

	for _, in := range constructor.In {
		typeGraphNode, err := c.typeGraphNode(in.Type)
		if err != nil {
			return reflect.Value{}, err
		}

		c.addGraphEdge(typeGraphNode, constructorGraphNode)
	}

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

		for i, out := range constructor.Out {
			typ := out.Type
			_, ok := c.resolvers[typ]
			if ok {
				return nil, &duplicateConstructorError{
					loc: constructor.Location,
					typ: typ,
				}
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

func markGraphNodeAsUsed(node *cgraph.Node) {
	node.SetColor("black")
}

func markGraphNodeAsFailed(node *cgraph.Node) {
	node.SetColor("red")
}
