package container

import (
	"bytes"
	"fmt"
	"reflect"

	"github.com/goccy/go-graphviz/cgraph"
	"github.com/pkg/errors"
)

type container struct {
	*debugConfig

	resolvers map[reflect.Type]resolver

	scopes map[string]Scope

	resolveStack []resolveFrame
	callerStack  []Location
	callerMap    map[Location]bool
}

type resolveFrame struct {
	loc Location
	typ reflect.Type
}

func newContainer(cfg *debugConfig) *container {
	return &container{
		debugConfig: cfg,
		resolvers:   map[reflect.Type]resolver{},
		scopes:      map[string]Scope{},
		callerStack: nil,
		callerMap:   map[Location]bool{},
	}
}

func (c *container) call(constructor *ProviderDescriptor, scope Scope) ([]reflect.Value, error) {
	loc := constructor.Location
	graphNode, err := c.locationGraphNode(loc, scope)
	if err != nil {
		return nil, err
	}
	markGraphNodeAsFailed(graphNode)

	if c.callerMap[loc] {
		return nil, errors.Errorf("cyclic dependency: %s -> %s", loc.Name(), loc.Name())
	}

	c.callerMap[loc] = true
	c.callerStack = append(c.callerStack, loc)

	c.logf("Resolving dependencies for %s", loc)
	c.indentLogger()
	inVals := make([]reflect.Value, len(constructor.Inputs))
	for i, in := range constructor.Inputs {
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

func (c *container) getResolver(typ reflect.Type) (resolver, error) {
	if vr, ok := c.resolvers[typ]; ok {
		return vr, nil
	}

	elemType := typ
	if isAutoGroupSliceType(elemType) || isOnePerScopeMapType(elemType) {
		elemType = elemType.Elem()
	}

	var typeGraphNode *cgraph.Node
	var err error

	if isAutoGroupType(elemType) {
		c.logf("Registering resolver for auto-group type %v", elemType)
		sliceType := reflect.SliceOf(elemType)

		typeGraphNode, err = c.typeGraphNode(sliceType)
		if err != nil {
			return nil, err
		}
		typeGraphNode.SetComment("auto-group")

		r := &groupResolver{
			typ:       elemType,
			sliceType: sliceType,
			graphNode: typeGraphNode,
		}

		c.resolvers[elemType] = r
		c.resolvers[sliceType] = &sliceGroupResolver{r}
	} else if isOnePerScopeType(elemType) {
		c.logf("Registering resolver for one-per-scope type %v", elemType)
		mapType := reflect.MapOf(stringType, elemType)

		typeGraphNode, err = c.typeGraphNode(mapType)
		if err != nil {
			return nil, err
		}
		typeGraphNode.SetComment("one-per-scope")

		r := &onePerScopeResolver{
			typ:       elemType,
			mapType:   mapType,
			providers: map[Scope]*simpleProvider{},
			idxMap:    map[Scope]int{},
			graphNode: typeGraphNode,
		}

		c.resolvers[elemType] = r
		c.resolvers[mapType] = &mapOfOnePerScopeResolver{r}
	}

	return c.resolvers[typ], nil
}

func (c *container) addNode(constructor *ProviderDescriptor, scope Scope) (interface{}, error) {
	constructorGraphNode, err := c.locationGraphNode(constructor.Location, scope)
	if err != nil {
		return nil, err
	}

	hasScopeParam := false
	for _, in := range constructor.Inputs {
		typ := in.Type
		if typ == scopeType {
			hasScopeParam = true
		}

		if isAutoGroupType(typ) {
			return nil, fmt.Errorf("auto-group type %v can't be used as an input parameter", typ)
		} else if isOnePerScopeType(typ) {
			return nil, fmt.Errorf("one-per-scope type %v can't be used as an input parameter", typ)
		}

		vr, err := c.getResolver(typ)
		if err != nil {
			return nil, err
		}

		var typeGraphNode *cgraph.Node
		if vr != nil {
			typeGraphNode = vr.typeGraphNode()
		} else {
			typeGraphNode, err = c.typeGraphNode(typ)
			if err != nil {
				return nil, err
			}
		}

		c.addGraphEdge(typeGraphNode, constructorGraphNode)
	}

	if scope != nil || !hasScopeParam {
		c.logf("Registering %s", constructor.Location.String())
		c.indentLogger()
		defer c.dedentLogger()

		sp := &simpleProvider{
			provider: constructor,
			scope:    scope,
		}

		for i, out := range constructor.Outputs {
			typ := out.Type

			// one-per-scope maps can't be used as a return type
			if isOnePerScopeMapType(typ) {
				return nil, fmt.Errorf("%v cannot be used as a return type because %v is a one-per-scope type",
					typ, typ.Elem())
			}

			// auto-group slices of auto-group types
			if isAutoGroupSliceType(typ) {
				typ = typ.Elem()
			}

			vr, err := c.getResolver(typ)
			if err != nil {
				return nil, err
			}

			if vr != nil {
				c.logf("Found resolver for %v: %T", typ, vr)
				err := vr.addNode(sp, i)
				if err != nil {
					return nil, err
				}
			} else {
				c.logf("Registering resolver for simple type %v", typ)

				typeGraphNode, err := c.typeGraphNode(typ)
				if err != nil {
					return nil, err
				}

				vr = &simpleResolver{
					node:      sp,
					typ:       typ,
					graphNode: typeGraphNode,
				}
				c.resolvers[typ] = vr
			}

			c.addGraphEdge(constructorGraphNode, vr.typeGraphNode())
		}

		return sp, nil
	} else {
		c.logf("Registering scope provider: %s", constructor.Location.String())
		c.indentLogger()
		defer c.dedentLogger()

		node := &scopeDepProvider{
			provider:       constructor,
			calledForScope: map[Scope]bool{},
			valueMap:       map[Scope][]reflect.Value{},
		}

		for i, out := range constructor.Outputs {
			typ := out.Type

			c.logf("Registering resolver for scoped type %v", typ)

			existing, ok := c.resolvers[typ]
			if ok {
				return nil, errors.Errorf("duplicate provision of type %v by scoped provider %s\n\talready provided by %s",
					typ, constructor.Location, existing.describeLocation())
			}

			typeGraphNode, err := c.typeGraphNode(typ)
			if err != nil {
				return reflect.Value{}, err
			}

			c.resolvers[typ] = &scopeDepResolver{
				typ:         typ,
				idxInValues: i,
				node:        node,
				valueMap:    map[Scope]reflect.Value{},
				graphNode:   typeGraphNode,
			}

			c.addGraphEdge(constructorGraphNode, typeGraphNode)
		}

		return node, nil
	}
}

func (c *container) supply(value reflect.Value, location Location) error {
	typ := value.Type()
	locGrapNode, err := c.locationGraphNode(location, nil)
	if err != nil {
		return err
	}
	markGraphNodeAsUsed(locGrapNode)

	typeGraphNode, err := c.typeGraphNode(typ)
	if err != nil {
		return err
	}

	c.addGraphEdge(locGrapNode, typeGraphNode)

	if existing, ok := c.resolvers[typ]; ok {
		return duplicateDefinitionError(typ, location, existing.describeLocation())
	}

	c.resolvers[typ] = &supplyResolver{
		typ:       typ,
		value:     value,
		loc:       location,
		graphNode: typeGraphNode,
	}

	return nil
}

func (c *container) resolve(in ProviderInput, scope Scope, caller Location) (reflect.Value, error) {
	c.resolveStack = append(c.resolveStack, resolveFrame{loc: caller, typ: in.Type})

	typeGraphNode, err := c.typeGraphNode(in.Type)
	if err != nil {
		return reflect.Value{}, err
	}

	if in.Type == scopeType {
		if scope == nil {
			return reflect.Value{}, errors.Errorf("trying to resolve %T for %s but not inside of any scope", scope, caller)
		}
		c.logf("Providing Scope %s", scope.Name())
		markGraphNodeAsUsed(typeGraphNode)
		return reflect.ValueOf(scope), nil
	}

	vr, err := c.getResolver(in.Type)
	if err != nil {
		return reflect.Value{}, err
	}

	if vr == nil {
		if in.Optional {
			c.logf("Providing zero value for optional dependency %v", in.Type)
			return reflect.Zero(in.Type), nil
		}

		markGraphNodeAsFailed(typeGraphNode)
		return reflect.Value{}, errors.Errorf("can't resolve type %v for %s:\n%s",
			in.Type, caller, c.formatResolveStack())
	}

	res, err := vr.resolve(c, scope, caller)
	if err != nil {
		markGraphNodeAsFailed(typeGraphNode)
		return reflect.Value{}, err
	}

	markGraphNodeAsUsed(typeGraphNode)

	c.resolveStack = c.resolveStack[:len(c.resolveStack)-1]

	return res, nil
}

func (c *container) run(invoker interface{}) error {
	rctr, err := ExtractProviderDescriptor(invoker)
	if err != nil {
		return err
	}

	if len(rctr.Outputs) > 0 {
		return errors.Errorf("invoker function cannot have return values other than error: %s", rctr.Location)
	}

	c.logf("Registering invoker")
	c.indentLogger()

	node, err := c.addNode(&rctr, nil)
	if err != nil {
		return err
	}

	c.dedentLogger()

	sn, ok := node.(*simpleProvider)
	if !ok {
		return errors.Errorf("cannot run scoped provider as an invoker")
	}

	c.logf("Building container")
	_, err = sn.resolveValues(c)
	if err != nil {
		return err
	}
	c.logf("Done building container")

	return nil
}

func (c container) createOrGetScope(name string) Scope {
	if s, ok := c.scopes[name]; ok {
		return s
	}
	s := newScope(name)
	c.scopes[name] = s
	return s
}

func (c container) formatResolveStack() string {
	buf := &bytes.Buffer{}
	_, _ = fmt.Fprintf(buf, "\twhile resolving:\n")
	n := len(c.resolveStack)
	for i := n - 1; i >= 0; i-- {
		rk := c.resolveStack[i]
		_, _ = fmt.Fprintf(buf, "\t\t%v for %s\n", rk.typ, rk.loc)
	}
	return buf.String()
}

func markGraphNodeAsUsed(node *cgraph.Node) {
	node.SetColor("black")
}

func markGraphNodeAsFailed(node *cgraph.Node) {
	node.SetColor("red")
}
