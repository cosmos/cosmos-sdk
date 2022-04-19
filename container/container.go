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

	moduleKeys map[string]*moduleKey

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
		moduleKeys:  map[string]*moduleKey{},
		callerStack: nil,
		callerMap:   map[Location]bool{},
	}
}

func (c *container) call(provider *ProviderDescriptor, moduleKey *moduleKey) ([]reflect.Value, error) {
	loc := provider.Location
	graphNode, err := c.locationGraphNode(loc, moduleKey)
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
	inVals := make([]reflect.Value, len(provider.Inputs))
	for i, in := range provider.Inputs {
		val, err := c.resolve(in, moduleKey, loc)
		if err != nil {
			return nil, err
		}
		inVals[i] = val
	}
	c.dedentLogger()
	c.logf("Calling %s", loc)

	delete(c.callerMap, loc)
	c.callerStack = c.callerStack[0 : len(c.callerStack)-1]

	out, err := provider.Fn(inVals)
	if err != nil {
		return nil, errors.Wrapf(err, "error calling provider %s", loc)
	}

	markGraphNodeAsUsed(graphNode)

	return out, nil
}

func (c *container) getResolver(typ reflect.Type) (resolver, error) {
	if vr, ok := c.resolvers[typ]; ok {
		return vr, nil
	}

	elemType := typ
	if isAutoGroupSliceType(elemType) || isOnePerModuleMapType(elemType) {
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
	} else if isOnePerModuleType(elemType) {
		c.logf("Registering resolver for one-per-module type %v", elemType)
		mapType := reflect.MapOf(stringType, elemType)

		typeGraphNode, err = c.typeGraphNode(mapType)
		if err != nil {
			return nil, err
		}
		typeGraphNode.SetComment("one-per-module")

		r := &onePerModuleResolver{
			typ:       elemType,
			mapType:   mapType,
			providers: map[*moduleKey]*simpleProvider{},
			idxMap:    map[*moduleKey]int{},
			graphNode: typeGraphNode,
		}

		c.resolvers[elemType] = r
		c.resolvers[mapType] = &mapOfOnePerModuleResolver{r}
	}

	return c.resolvers[typ], nil
}

func (c *container) addNode(provider *ProviderDescriptor, key *moduleKey) (interface{}, error) {
	providerGraphNode, err := c.locationGraphNode(provider.Location, key)
	if err != nil {
		return nil, err
	}

	hasModuleKeyParam := false
	for _, in := range provider.Inputs {
		typ := in.Type
		if typ == moduleKeyType {
			hasModuleKeyParam = true
		}

		if isAutoGroupType(typ) {
			return nil, fmt.Errorf("auto-group type %v can't be used as an input parameter", typ)
		} else if isOnePerModuleType(typ) {
			return nil, fmt.Errorf("one-per-module type %v can't be used as an input parameter", typ)
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

		c.addGraphEdge(typeGraphNode, providerGraphNode)
	}

	if key != nil || !hasModuleKeyParam {
		c.logf("Registering %s", provider.Location.String())
		c.indentLogger()
		defer c.dedentLogger()

		sp := &simpleProvider{
			provider:  provider,
			moduleKey: key,
		}

		for i, out := range provider.Outputs {
			typ := out.Type

			// one-per-module maps can't be used as a return type
			if isOnePerModuleMapType(typ) {
				return nil, fmt.Errorf("%v cannot be used as a return type because %v is a one-per-module type",
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

			c.addGraphEdge(providerGraphNode, vr.typeGraphNode())
		}

		return sp, nil
	} else {
		c.logf("Registering module-scoped provider: %s", provider.Location.String())
		c.indentLogger()
		defer c.dedentLogger()

		node := &moduleDepProvider{
			provider:        provider,
			calledForModule: map[*moduleKey]bool{},
			valueMap:        map[*moduleKey][]reflect.Value{},
		}

		for i, out := range provider.Outputs {
			typ := out.Type

			c.logf("Registering resolver for module-scoped type %v", typ)

			existing, ok := c.resolvers[typ]
			if ok {
				return nil, errors.Errorf("duplicate provision of type %v by module-scoped provider %s\n\talready provided by %s",
					typ, provider.Location, existing.describeLocation())
			}

			typeGraphNode, err := c.typeGraphNode(typ)
			if err != nil {
				return reflect.Value{}, err
			}

			c.resolvers[typ] = &moduleDepResolver{
				typ:         typ,
				idxInValues: i,
				node:        node,
				valueMap:    map[*moduleKey]reflect.Value{},
				graphNode:   typeGraphNode,
			}

			c.addGraphEdge(providerGraphNode, typeGraphNode)
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

func (c *container) resolve(in ProviderInput, moduleKey *moduleKey, caller Location) (reflect.Value, error) {
	c.resolveStack = append(c.resolveStack, resolveFrame{loc: caller, typ: in.Type})

	typeGraphNode, err := c.typeGraphNode(in.Type)
	if err != nil {
		return reflect.Value{}, err
	}

	if in.Type == moduleKeyType {
		if moduleKey == nil {
			return reflect.Value{}, errors.Errorf("trying to resolve %T for %s but not inside of any module's scope", moduleKey, caller)
		}
		c.logf("Providing ModuleKey %s", moduleKey.name)
		markGraphNodeAsUsed(typeGraphNode)
		return reflect.ValueOf(ModuleKey{moduleKey}), nil
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

	res, err := vr.resolve(c, moduleKey, caller)
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
		return errors.Errorf("cannot run module-scoped provider as an invoker")
	}

	c.logf("Building container")
	_, err = sn.resolveValues(c)
	if err != nil {
		return err
	}
	c.logf("Done building container")

	return nil
}

func (c container) createOrGetModuleKey(name string) *moduleKey {
	if s, ok := c.moduleKeys[name]; ok {
		return s
	}
	s := &moduleKey{name}
	c.moduleKeys[name] = s
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
