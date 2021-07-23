package container

import (
	"fmt"
	"reflect"
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
		ctr.valueResolvers[typ] = &groupValueResolver{
			typ: typ,
		}
	}

	for typ := range cfg.onePerScopeTypes {
		ctr.valueResolvers[typ] = &onePerScopeValueResolver{
			typ:   typ,
			nodes: map[Scope]*simpleNode{},
		}
	}

	return ctr
}

type valueResolver interface {
	addNode(*simpleNode) error
	resolve(*container, Scope) (reflect.Value, error)
}

type simpleNode struct {
	ctr    *ReflectConstructor
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

func (c *container) call(constructor *ReflectConstructor, scope Scope) ([]reflect.Value, error) {
	inVals := make([]reflect.Value, len(constructor.In))
	for i, in := range constructor.In {
		val, err := c.resolve(in, scope)
		if err != nil {
			return nil, err
		}
		inVals[i] = val
	}
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

func (s simpleValueResolver) addNode(n *simpleNode) error {
	return fmt.Errorf("duplicate constructor for type %v", s.typ)
}

type groupValueResolver struct {
	typ          reflect.Type
	idxsInValues []int
	nodes        []*simpleNode
	resolved     bool
	values       reflect.Value
}

func (g *groupValueResolver) resolve(c *container, s Scope) (reflect.Value, error) {
	if !g.resolved {
		res := reflect.MakeSlice(g.typ, 0, 0)
		for i, node := range g.nodes {
			values, err := node.resolveValues(c)
			if err != nil {
				return reflect.Value{}, err
			}
			value := values[i]
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

func (g *groupValueResolver) addNode(n *simpleNode) error {
	g.nodes = append(g.nodes, n)
	return nil
}

type onePerScopeValueResolver struct {
	typ      reflect.Type
	nodes    map[Scope]*simpleNode
	resolved bool
	values   reflect.Value
}

func (o *onePerScopeValueResolver) resolve(c *container, s Scope) (reflect.Value, error) {
	if !o.resolved {
		panic("TODO")
	}

	return o.values, nil
}

func (o *onePerScopeValueResolver) addNode(n *simpleNode) error {
	if n.scope == nil {
		return fmt.Errorf("cannot define a constructor with one-per-scope dependency %T which isn't provided in a scope", o.typ)
	}
	panic("implement me")
	return nil
}

type scopeProviderNode struct {
	ctr            *ReflectConstructor
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
	if val, ok := s.valueMap[scope]; ok {
		return val, nil
	}

	if s.node.calledForScope[scope] {
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

func (s scopeProviderValueResolver) addNode(*simpleNode) error {
	return fmt.Errorf("duplicate constructor for type %v", s.typ)
}

func reflectCtr(ctr interface{}) (*ReflectConstructor, error) {
	reflectCtr, ok := ctr.(ReflectConstructor)
	if !ok {
		val := reflect.ValueOf(ctr)
		typ := val.Type()
		if typ.Kind() != reflect.Func {
			return nil, fmt.Errorf("expected a Func type, got %T", ctr)
		}

		numIn := typ.NumIn()
		in := make([]reflect.Type, numIn)
		for i := 0; i < numIn; i++ {
			in[i] = typ.In(i)
		}

		numOut := typ.NumOut()
		out := make([]reflect.Type, numOut)
		for i := 0; i < numOut; i++ {
			out[i] = typ.Out(i)
		}

		reflectCtr = ReflectConstructor{
			In:  in,
			Out: out,
			Fn: func(values []reflect.Value) []reflect.Value {
				return val.Call(values)
			},
			Location: nil,
		}
	}

	return &reflectCtr, nil
}

func (ctr *container) addNode(constructor *ReflectConstructor, scope Scope) error {
	hasScopeParam := len(constructor.In) > 0 && constructor.In[0] == scopeTyp
	if scope != nil || !hasScopeParam {
		node := &simpleNode{
			ctr:   constructor,
			scope: scope,
		}

		for _, typ := range constructor.Out {
			// auto-group slices of auto-group types
			if typ.Kind() == reflect.Slice && ctr.autoGroupTypes[typ.Elem()] {
				typ = typ.Elem()
			}

			vr, ok := ctr.valueResolvers[typ]
			if ok {
				err := vr.addNode(node)
				if err != nil {
					return err
				}
			} else {
				ctr.valueResolvers[typ] = &simpleValueResolver{
					node: node,
					typ:  typ,
				}
			}
		}
	} else {
		node := &scopeProviderNode{
			ctr:            constructor,
			calledForScope: map[Scope]bool{},
			valueMap:       map[Scope][]reflect.Value{},
		}

		for i, typ := range constructor.Out {
			_, ok := ctr.valueResolvers[typ]
			if ok {
				return fmt.Errorf("duplicate constructor for type %v", typ)
			}
			ctr.valueResolvers[typ] = &scopeProviderValueResolver{
				typ:         typ,
				idxInValues: i,
				node:        node,
				valueMap:    nil,
			}
		}
	}

	return nil
}

func (ctr *container) resolve(typ reflect.Type, scope Scope) (reflect.Value, error) {
	vr, ok := ctr.valueResolvers[typ]
	if !ok {
		return reflect.Value{}, fmt.Errorf("no constructor for type %v", typ)
	}

	return vr.resolve(ctr, scope)
}
