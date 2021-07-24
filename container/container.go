package container

import (
	"fmt"
	"reflect"

	containerreflect "github.com/cosmos/cosmos-sdk/container/reflect"
)

type container struct {
	*config

	valueResolvers map[reflect.Type]valueResolver
}

func newContainer(cfg *config) *container {
	ctr := &container{
		config:         cfg,
		valueResolvers: map[reflect.Type]valueResolver{},
	}

	for typ := range cfg.autoGroupTypes {
		sliceType := reflect.SliceOf(typ)
		r := &groupValueResolver{
			typ:       typ,
			sliceType: sliceType,
		}
		ctr.valueResolvers[typ] = r
		ctr.valueResolvers[sliceType] = &sliceGroupValueResolver{r}
	}

	for typ := range cfg.onePerScopeTypes {
		mapType := reflect.MapOf(scopeType, typ)
		r := &onePerScopeValueResolver{
			typ:     typ,
			mapType: mapType,
			nodes:   map[Scope]*simpleNode{},
			idxMap:  map[Scope]int{},
		}
		ctr.valueResolvers[typ] = r
		ctr.valueResolvers[mapType] = &mapOfOnePerScopeValueResolver{r}
	}

	return ctr
}

type valueResolver interface {
	addNode(*simpleNode, int) error
	resolve(*container, Scope) (reflect.Value, error)
}

type simpleNode struct {
	ctr    *containerreflect.Constructor
	called bool
	values []reflect.Value
	scope  Scope
}

type simpleValueResolver struct {
	node        *simpleNode
	idxInValues int
	resolved    bool
	typ         reflect.Type
	value       reflect.Value
}

func (c *container) call(constructor *containerreflect.Constructor, scope Scope) ([]reflect.Value, error) {
	c.logf("Gathering dependencies for %s", constructor.Location)
	c.indentLogger()
	inVals := make([]reflect.Value, len(constructor.In))
	for i, in := range constructor.In {
		val, err := c.resolve(in, scope)
		if err != nil {
			return nil, err
		}
		inVals[i] = val
	}
	c.dedentLogger()
	c.logf("Calling %s", constructor.Location)
	return constructor.Fn(inVals), nil
}

func (s *simpleNode) resolveValues(ctr *container) ([]reflect.Value, error) {
	if !s.called {
		values, err := ctr.call(s.ctr, s.scope)
		if err != nil {
			return nil, err
		}
		s.values = values
		s.called = true
	}

	return s.values, nil
}

func (s *simpleValueResolver) resolve(c *container, _ Scope) (reflect.Value, error) {
	c.logf("Providing %v from %s", s.typ, s.node.ctr.Location)
	if s.resolved {
		return s.value, nil
	}

	values, err := s.node.resolveValues(c)
	if err != nil {
		return reflect.Value{}, err
	}

	value := values[s.idxInValues]
	s.value = value
	s.resolved = true
	return value, nil
}

func (s simpleValueResolver) addNode(*simpleNode, int) error {
	return fmt.Errorf("duplicate constructor for type %v", s.typ)
}

type groupValueResolver struct {
	typ          reflect.Type
	sliceType    reflect.Type
	idxsInValues []int
	nodes        []*simpleNode
	resolved     bool
	values       reflect.Value
}

type sliceGroupValueResolver struct {
	*groupValueResolver
}

func (g *sliceGroupValueResolver) resolve(c *container, _ Scope) (reflect.Value, error) {
	c.logf("Providing %v from:", g.sliceType)
	c.indentLogger()
	for _, node := range g.nodes {
		c.logf(node.ctr.Location.String())
	}
	c.dedentLogger()
	if !g.resolved {
		res := reflect.MakeSlice(g.sliceType, 0, 0)
		for i, node := range g.nodes {
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

func (g *groupValueResolver) resolve(_ *container, _ Scope) (reflect.Value, error) {
	return reflect.Value{}, fmt.Errorf("%v is an auto-group type and cannot be used as an input value, instead use %v", g.typ, g.sliceType)
}

func (g *groupValueResolver) addNode(n *simpleNode, i int) error {
	g.nodes = append(g.nodes, n)
	g.idxsInValues = append(g.idxsInValues, i)
	return nil
}

type onePerScopeValueResolver struct {
	typ      reflect.Type
	mapType  reflect.Type
	nodes    map[Scope]*simpleNode
	idxMap   map[Scope]int
	resolved bool
	values   reflect.Value
}

type mapOfOnePerScopeValueResolver struct {
	*onePerScopeValueResolver
}

func (o *onePerScopeValueResolver) resolve(_ *container, _ Scope) (reflect.Value, error) {
	return reflect.Value{}, fmt.Errorf("%v is a one-per-scope type and thus can't be used as an input parameter, instead use %v", o.typ, o.mapType)
}

func (o *mapOfOnePerScopeValueResolver) resolve(c *container, _ Scope) (reflect.Value, error) {
	c.logf("Providing %v from:", o.mapType)
	c.indentLogger()
	for scope, node := range o.nodes {
		c.logf("%s: %s", scope.Name(), node.ctr.Location)
	}
	c.dedentLogger()
	if !o.resolved {
		res := reflect.MakeMap(o.mapType)
		for scope, node := range o.nodes {
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

func (o *onePerScopeValueResolver) addNode(n *simpleNode, i int) error {
	if n.scope == nil {
		return fmt.Errorf("cannot define a constructor with one-per-scope dependency %v which isn't provided in a scope", o.typ)
	}

	if _, ok := o.nodes[n.scope]; ok {
		return fmt.Errorf("duplicate constructor for one-per-scope type %v in scope %s", o.typ, n.scope)
	}

	o.nodes[n.scope] = n
	o.idxMap[n.scope] = i

	return nil
}

func (o *mapOfOnePerScopeValueResolver) addNode(*simpleNode, int) error {
	return fmt.Errorf("%v is a one-per-scope type and thus %v can't be used as an output parameter", o.typ, o.mapType)
}

type scopeProviderNode struct {
	ctr            *containerreflect.Constructor
	calledForScope map[Scope]bool
	valueMap       map[Scope][]reflect.Value
}

type scopeProviderValueResolver struct {
	typ         reflect.Type
	idxInValues int
	node        *scopeProviderNode
	valueMap    map[Scope]reflect.Value
}

func (s scopeProviderValueResolver) resolve(ctr *container, scope Scope) (reflect.Value, error) {
	ctr.logf("Providing %v from %s", s.typ, s.node.ctr.Location)
	if val, ok := s.valueMap[scope]; ok {
		return val, nil
	}

	if !s.node.calledForScope[scope] {
		values, err := ctr.call(s.node.ctr, scope)
		if err != nil {
			return reflect.Value{}, err
		}

		s.node.valueMap[scope] = values
		s.node.calledForScope[scope] = true
	}

	value := s.node.valueMap[scope][s.idxInValues]
	s.valueMap[scope] = value
	return value, nil
}

func (s scopeProviderValueResolver) addNode(*simpleNode, int) error {
	return fmt.Errorf("duplicate constructor for type %v", s.typ)
}

func reflectCtr(ctr interface{}) (*containerreflect.Constructor, error) {
	rctr, ok := ctr.(containerreflect.Constructor)
	if !ok {
		val := reflect.ValueOf(ctr)
		typ := val.Type()
		if typ.Kind() != reflect.Func {
			return nil, fmt.Errorf("expected a Func type, got %T", ctr)
		}

		numIn := typ.NumIn()
		in := make([]containerreflect.Input, numIn)
		for i := 0; i < numIn; i++ {
			in[i] = containerreflect.Input{
				Type: typ.In(i),
			}
		}

		numOut := typ.NumOut()
		out := make([]containerreflect.Output, numOut)
		for i := 0; i < numOut; i++ {
			out[i] = containerreflect.Output{Type: typ.Out(i)}
		}

		rctr = containerreflect.Constructor{
			In:  in,
			Out: out,
			Fn: func(values []reflect.Value) []reflect.Value {
				return val.Call(values)
			},
			Location: containerreflect.LocationFromPC(val.Pointer()),
		}
	}

	return &rctr, nil
}

func (c *container) addNode(constructor *containerreflect.Constructor, scope Scope) (interface{}, error) {
	hasScopeParam := len(constructor.In) > 0 && constructor.In[0].Type == scopeType
	if scope != nil || !hasScopeParam {
		node := &simpleNode{
			ctr:   constructor,
			scope: scope,
		}

		for i, out := range constructor.Out {
			typ := out.Type
			// auto-group slices of auto-group types
			if typ.Kind() == reflect.Slice && c.autoGroupTypes[typ.Elem()] {
				typ = typ.Elem()
			}

			vr, ok := c.valueResolvers[typ]
			if ok {
				err := vr.addNode(node, i)
				if err != nil {
					return nil, err
				}
			} else {
				c.valueResolvers[typ] = &simpleValueResolver{
					node: node,
					typ:  typ,
				}
			}
		}

		return node, nil
	} else {
		node := &scopeProviderNode{
			ctr:            constructor,
			calledForScope: map[Scope]bool{},
			valueMap:       map[Scope][]reflect.Value{},
		}

		for i, out := range constructor.Out {
			typ := out.Type
			_, ok := c.valueResolvers[typ]
			if ok {
				return nil, fmt.Errorf("duplicate constructor for type %v", typ)
			}
			c.valueResolvers[typ] = &scopeProviderValueResolver{
				typ:         typ,
				idxInValues: i,
				node:        node,
				valueMap:    map[Scope]reflect.Value{},
			}
		}

		return node, nil
	}
}

func (c *container) resolve(in containerreflect.Input, scope Scope) (reflect.Value, error) {
	if in.Type == scopeType {
		if scope == nil {
			return reflect.Value{}, fmt.Errorf("expected scope but got nil")
		}
		c.logf("Providing Scope %s", scope.Name())
		return reflect.ValueOf(scope), nil
	}

	vr, ok := c.valueResolvers[in.Type]
	if !ok {
		if in.Optional {
			c.logf("Providing zero value for optional dependency %v", in.Type)
			return reflect.Zero(in.Type), nil
		}

		return reflect.Value{}, fmt.Errorf("no constructor for type %v", in.Type)
	}

	return vr.resolve(c, scope)
}

func (c *container) run(invoker interface{}) error {
	rctr, err := reflectCtr(invoker)
	if err != nil {
		return err
	}

	node, err := c.addNode(rctr, nil)
	if err != nil {
		return err
	}

	sn, ok := node.(*simpleNode)
	if !ok {
		return fmt.Errorf("cannot run scoped provider as an invoker")
	}

	_, err = sn.resolveValues(c)
	return err
}
