package container

import (
	"fmt"
	"reflect"
)

// Container is a low-level dependency injection container which manages dependencies
// based on scopes and security policies. All providers can be run in a scope which
// may provide certain dependencies specifically for that scope or provide/deny access
// to dependencies based on that scope.
type Container struct {
	providers      map[Key]*node
	scopeProviders map[Key]*scopeNode
	nodes          []*node
	scopeNodes     []*scopeNode

	values          map[Key]secureValue
	scopedValues    map[Scope]map[Key]reflect.Value
	securityContext func(scope Scope, tag string) error
}

func NewContainer() *Container {
	return &Container{
		providers:      map[Key]*node{},
		scopeProviders: map[Key]*scopeNode{},
		nodes:          nil,
		scopeNodes:     nil,
		values:         map[Key]secureValue{},
		scopedValues:   map[Scope]map[Key]reflect.Value{},
	}
}

type Input struct {
	Key
	Optional bool
}

type Output struct {
	Key
	SecurityChecker SecurityChecker
}

type Key struct {
	Type reflect.Type
}

type Scope string

type node struct {
	Provider
	called bool
	values []reflect.Value
	err    error
}

// Provider is a general dependency provider. Its scope parameter is used
// to receive scoped dependencies and gain access to general dependencies within
// its security policy. Access to dependencies provided by this provider can optionally
// be restricted to certain scopes based on SecurityCheckers.
type Provider struct {
	// Constructor provides the dependencies
	Constructor func(deps []reflect.Value, scope Scope) ([]reflect.Value, error)

	// Needs are the keys for dependencies the constructor needs
	Needs []Input

	// Needs are the keys for dependencies the constructor provides
	Provides []Output

	// Scope is the scope within which the constructor runs
	Scope Scope

	IsScopeProvider bool
}

type scopeNode struct {
	Provider
	calledForScope map[Scope]bool
	valuesForScope map[Scope][]reflect.Value
	errsForScope   map[Scope]error
}

// ScopeProvider provides scoped dependencies. Its constructor function will provide
// dependencies specific to the scope parameter. Instead of providing general dependencies
// with restricted access based on security checkers, ScopeProvider provides potentially different
// dependency instances to different scopes. It is assumed that a scoped provider
// can provide a dependency for any valid scope passed to it, although it can return an error
// to deny access.
type ScopeProvider struct {

	// Constructor provides dependencies for the provided scope
	Constructor func(scope Scope, deps []reflect.Value) ([]reflect.Value, error)

	// Needs are the keys for dependencies the constructor needs
	Needs []Input

	// Needs are the keys for dependencies the constructor provides
	Provides []Key

	// Scope is the scope within which the constructor runs, if it is left empty,
	// the constructor runs in the scope it was called with (this only applies to ScopeProvider).
	Scope Scope
}

type secureValue struct {
	value           reflect.Value
	securityChecker SecurityChecker
}

type SecurityChecker func(scope Scope) error

func (c *Container) RegisterProvider(provider Provider) error {
	if !provider.IsScopeProvider {
		n := &node{
			Provider: provider,
			called:   false,
		}

		c.nodes = append(c.nodes, n)

		for _, key := range provider.Provides {
			if c.providers[key.Key] != nil {
				return fmt.Errorf("TODO")
			}

			if c.scopeProviders[key.Key] != nil {
				return fmt.Errorf("TODO")
			}

			c.providers[key.Key] = n
		}
	} else {
		n := &scopeNode{
			Provider:       provider,
			calledForScope: map[Scope]bool{},
			valuesForScope: map[Scope][]reflect.Value{},
			errsForScope:   map[Scope]error{},
		}

		c.scopeNodes = append(c.scopeNodes, n)

		for _, key := range provider.Provides {
			if c.providers[key.Key] != nil {
				return fmt.Errorf("TODO")
			}

			if c.scopeProviders[key.Key] != nil {
				return fmt.Errorf("TODO")
			}

			c.scopeProviders[key.Key] = n
		}

		return nil
	}

	return nil
}

//func (c *Container) RegisterScopeProvider(provider *ScopeProvider) error {
//	n := &scopeNode{
//		ScopeProvider:  provider,
//		calledForScope: map[Scope]bool{},
//		valuesForScope: map[Scope][]reflect.Value{},
//		errsForScope:   map[Scope]error{},
//	}
//
//	c.scopeNodes = append(c.scopeNodes, n)
//
//	for _, key := range provider.Provides {
//		if c.scopeProviders[key] != nil {
//			return fmt.Errorf("TODO")
//		}
//
//		c.scopeProviders[key] = n
//	}
//
//	return nil
//}

func (c *Container) resolve(scope Scope, input Input, stack map[interface{}]bool) (reflect.Value, error) {
	if scope != "" {
		if val, ok := c.scopedValues[scope][input.Key]; ok {
			return val, nil
		}

		if provider, ok := c.scopeProviders[input.Key]; ok {
			if stack[provider] {
				return reflect.Value{}, fmt.Errorf("fatal: cycle detected")
			}

			if provider.calledForScope[scope] {
				return reflect.Value{}, fmt.Errorf("error: %v", provider.errsForScope[scope])
			}

			var deps []reflect.Value
			for _, need := range provider.Needs {
				subScope := provider.Scope
				// for ScopeProvider we default to the calling scope
				if subScope == "" {
					subScope = scope
				}
				stack[provider] = true
				res, err := c.resolve(subScope, need, stack)
				delete(stack, provider)

				if err != nil {
					return reflect.Value{}, err
				}

				deps = append(deps, res)
			}

			res, err := provider.Constructor(deps, scope)
			provider.calledForScope[scope] = true
			if err != nil {
				provider.errsForScope[scope] = err
				return reflect.Value{}, err
			}

			provider.valuesForScope[scope] = res

			for i, val := range res {
				p := provider.Provides[i]
				if _, ok := c.scopedValues[scope][p.Key]; ok {
					return reflect.Value{}, fmt.Errorf("value provided twice")
				}

				if c.scopedValues[scope] == nil {
					c.scopedValues[scope] = map[Key]reflect.Value{}
				}
				c.scopedValues[scope][p.Key] = val
			}

			val, ok := c.scopedValues[scope][input.Key]
			if !ok {
				return reflect.Value{}, fmt.Errorf("internal error: bug")
			}

			return val, nil
		}
	}

	if val, ok, err := c.getValue(scope, input.Key); ok {
		if err != nil {
			return reflect.Value{}, err
		}

		return val, nil
	}

	if provider, ok := c.providers[input.Key]; ok {
		if stack[provider] {
			return reflect.Value{}, fmt.Errorf("fatal: cycle detected")
		}

		if provider.called {
			return reflect.Value{}, fmt.Errorf("error: %v", provider.err)
		}

		err := c.execNode(provider, stack)
		if err != nil {
			return reflect.Value{}, err
		}

		val, ok, err := c.getValue(scope, input.Key)
		if !ok {
			return reflect.Value{}, fmt.Errorf("internal error: bug")
		}

		return val, err
	}

	if input.Optional {
		return reflect.Zero(input.Type), nil
	}

	return reflect.Value{}, fmt.Errorf("no provider")
}

func (c *Container) execNode(provider *node, stack map[interface{}]bool) error {
	if provider.called {
		return provider.err
	}

	var deps []reflect.Value
	for _, need := range provider.Needs {
		stack[provider] = true
		res, err := c.resolve(provider.Scope, need, stack)
		delete(stack, provider)

		if err != nil {
			return err
		}

		deps = append(deps, res)
	}

	res, err := provider.Constructor(deps, "")
	provider.called = true
	if err != nil {
		provider.err = err
		return err
	}

	provider.values = res

	for i, val := range res {
		p := provider.Provides[i]
		if _, ok := c.values[p.Key]; ok {
			return fmt.Errorf("value provided twice")
		}

		c.values[p.Key] = secureValue{
			value:           val,
			securityChecker: p.SecurityChecker,
		}
	}

	return nil
}

func (c *Container) getValue(scope Scope, key Key) (reflect.Value, bool, error) {
	if val, ok := c.values[key]; ok {
		if val.securityChecker != nil {
			if err := val.securityChecker(scope); err != nil {
				return reflect.Value{}, true, err
			}
		}

		return val.value, true, nil
	}

	return reflect.Value{}, false, nil
}

func (c *Container) Resolve(scope Scope, key Key) (reflect.Value, error) {
	val, err := c.resolve(scope, Input{
		Key:      key,
		Optional: false,
	}, map[interface{}]bool{})
	if err != nil {
		return reflect.Value{}, err
	}
	return val, nil
}

// InitializeAll attempts to call all providers instantiating the dependencies they provide
func (c *Container) InitializeAll() error {
	for _, node := range c.nodes {
		err := c.execNode(node, map[interface{}]bool{})
		if err != nil {
			return err
		}
	}
	return nil
}

type StructArgs struct{}

func (StructArgs) isStructArgs() {}

type isStructArgs interface{ isStructArgs() }

var structArgsType = reflect.TypeOf(StructArgs{})

var isStructArgsTyp = reflect.TypeOf((*isStructArgs)(nil)).Elem()

var scopeTyp = reflect.TypeOf(Scope(""))

type InMarshaler func([]reflect.Value) reflect.Value

type inFieldMarshaler struct {
	n           int
	inMarshaler InMarshaler
}

type OutMarshaler func(reflect.Value) []reflect.Value

func TypeToInput(typ reflect.Type) ([]Input, InMarshaler, error) {
	if typ.AssignableTo(isStructArgsTyp) && typ.Kind() == reflect.Struct {
		nFields := typ.NumField()
		var res []Input

		var marshalers []inFieldMarshaler

		for i := 0; i < nFields; i++ {
			field := typ.Field(i)
			if field.Type == structArgsType {
				marshalers = append(marshalers, inFieldMarshaler{
					n: 0,
					inMarshaler: func(values []reflect.Value) reflect.Value {
						return reflect.ValueOf(StructArgs{})
					},
				})
			} else {
				fieldInputs, m, err := TypeToInput(field.Type)
				if err != nil {
					return nil, nil, err
				}

				optionalTag, ok := field.Tag.Lookup("optional")
				if ok {
					if len(fieldInputs) == 1 {
						if optionalTag != "true" {
							return nil, nil, fmt.Errorf("true is the only valid value for the optional tag, got %s", optionalTag)
						}
						fieldInputs[0].Optional = true
					} else if len(fieldInputs) > 1 {
						return nil, nil, fmt.Errorf("optional tag cannot be applied to nested StructArgs")
					}
				}

				res = append(res, fieldInputs...)
				marshalers = append(marshalers, inFieldMarshaler{
					n:           len(fieldInputs),
					inMarshaler: m,
				})
			}
		}

		return res, structMarshaler(typ, marshalers), nil
	} else if typ == scopeTyp {
		return nil, nil, fmt.Errorf("can't convert type %T to %T", Scope(""), Input{})
	} else {
		return []Input{{
				Key: Key{
					Type: typ,
				},
			}}, func(values []reflect.Value) reflect.Value {
				return values[0]
			}, nil
	}
}

func TypeToOutput(typ reflect.Type, securityContext func(scope Scope, tag string) error) ([]Output, OutMarshaler, error) {
	if typ.AssignableTo(isStructArgsTyp) && typ.Kind() == reflect.Struct {
		nFields := typ.NumField()
		var res []Output
		var marshalers []OutMarshaler

		for i := 0; i < nFields; i++ {
			field := typ.Field(i)
			fieldOutputs, fieldMarshaler, err := TypeToOutput(field.Type, securityContext)
			if err != nil {
				return nil, nil, err
			}

			securityTag, ok := field.Tag.Lookup("security")
			if ok {
				if len(fieldOutputs) == 1 {
					if securityContext == nil {
						return nil, nil, fmt.Errorf("security tag is invalid in this context")
					}
					fieldOutputs[0].SecurityChecker = func(scope Scope) error {
						return securityContext(scope, securityTag)
					}
				} else if len(fieldOutputs) > 1 {
					return nil, nil, fmt.Errorf("security tag cannot be applied to nested StructArgs")
				}
			}

			res = append(res, fieldOutputs...)
			marshalers = append(marshalers, fieldMarshaler)
		}
		return res, func(value reflect.Value) []reflect.Value {
			var vals []reflect.Value
			for i := 0; i < nFields; i++ {
				val := value.Field(i)
				vals = append(vals, marshalers[i](val)...)
			}
			return vals
		}, nil
	} else if typ == scopeTyp {
		return nil, nil, fmt.Errorf("can't convert type %T to %T", Scope(""), Input{})
	} else {
		return []Output{{
				Key: Key{
					Type: typ,
				},
			}}, func(val reflect.Value) []reflect.Value {
				return []reflect.Value{val}
			}, nil
	}
}

func structMarshaler(typ reflect.Type, marshalers []inFieldMarshaler) func([]reflect.Value) reflect.Value {
	return func(values []reflect.Value) reflect.Value {
		structInst := reflect.Zero(typ)

		for i, m := range marshalers {
			val := m.inMarshaler(values[:m.n])
			structInst.Field(i).Set(val)
			values = values[m.n:]
		}

		return structInst
	}
}

func (c *Container) Provide(constructor interface{}) error {
	return c.ProvideWithScope(constructor, "")
}

func (c *Container) ProvideWithScope(constructor interface{}, scope Scope) error {
	p, err := ConstructorToProvider(constructor, scope, c.securityContext)
	if err != nil {
		return err
	}

	return c.RegisterProvider(p)
}

func ConstructorToProvider(constructor interface{}, scope Scope, securityContext func(scope Scope, tag string) error) (Provider, error) {
	ctrTyp := reflect.TypeOf(constructor)
	if ctrTyp.Kind() != reflect.Func {
		return Provider{}, fmt.Errorf("expected function got %T", constructor)
	}

	numIn := ctrTyp.NumIn()
	numOut := ctrTyp.NumOut()

	var scopeProvider bool
	i := 0
	if numIn >= 1 {
		if in0 := ctrTyp.In(0); in0 == scopeTyp {
			scopeProvider = true
			i = 1
		}
	}

	var inputs []Input
	var inMarshalers []inFieldMarshaler
	for ; i < numIn; i++ {
		in, inMarshaler, err := TypeToInput(ctrTyp.In(i))
		if err != nil {
			return Provider{}, err
		}
		inputs = append(inputs, in...)
		inMarshalers = append(inMarshalers, inFieldMarshaler{
			n:           len(in),
			inMarshaler: inMarshaler,
		})
	}

	var outputs []Output
	var outMarshalers []OutMarshaler
	for i := 0; i < numOut; i++ {
		out, outMarshaler, err := TypeToOutput(ctrTyp.Out(i), securityContext)
		if err != nil {
			return Provider{}, err
		}
		outputs = append(outputs, out...)
		outMarshalers = append(outMarshalers, outMarshaler)
	}

	ctrVal := reflect.ValueOf(constructor)
	provideCtr := func(deps []reflect.Value, scope Scope) ([]reflect.Value, error) {
		var inVals []reflect.Value

		if scopeProvider {
			inVals = append(inVals, reflect.ValueOf(scope))
		}

		nInMarshalers := len(inMarshalers)
		for i = 0; i < nInMarshalers; i++ {
			m := inMarshalers[i]
			inVals = append(inVals, m.inMarshaler(deps[:m.n]))
			deps = deps[m.n:]
		}

		outVals := ctrVal.Call(inVals)

		var provides []reflect.Value
		for i := 0; i < numOut; i++ {
			provides = append(provides, outMarshalers[i](outVals[i])...)
		}

		return outVals, nil
	}

	return Provider{
		Constructor:     provideCtr,
		Needs:           inputs,
		Provides:        outputs,
		Scope:           scope,
		IsScopeProvider: scopeProvider,
	}, nil
}

func (c *Container) Invoke(fn interface{}) error {
	fnTyp := reflect.TypeOf(fn)
	if fnTyp.Kind() != reflect.Func {
		return fmt.Errorf("expected function got %T", fn)
	}

	numIn := fnTyp.NumIn()
	in := make([]reflect.Value, numIn)
	for i := 0; i < numIn; i++ {
		val, err := c.Resolve("", Key{Type: fnTyp.In(i)})
		if err != nil {
			return err
		}
		in[i] = val
	}

	_ = reflect.ValueOf(fn).Call(in)

	return nil
}
