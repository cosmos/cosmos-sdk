package depinject

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/token"
	"reflect"

	"github.com/pkg/errors"

	"cosmossdk.io/depinject/internal/graphviz"
	"cosmossdk.io/depinject/internal/util"
)

type container struct {
	*debugConfig

	resolvers         map[string]resolver
	interfaceBindings map[string]interfaceBinding
	invokers          []invoker

	moduleKeyContext *ModuleKeyContext

	resolveStack []resolveFrame
	callerStack  []Location
	callerMap    map[Location]bool
}

type invoker struct {
	fn     *ProviderDescriptor
	modKey *moduleKey
}

type resolveFrame struct {
	loc Location
	typ reflect.Type
}

// interfaceBinding defines a type binding for interfaceName to type implTypeName when being provided as a
// dependency to the module identified by moduleKey.  If moduleKey is nil then the type binding is applied globally,
// not module-scoped.
type interfaceBinding struct {
	interfaceName string
	implTypeName  string
	moduleKey     *moduleKey
	resolver      resolver
}

func newContainer(cfg *debugConfig) *container {
	return &container{
		debugConfig:       cfg,
		resolvers:         map[string]resolver{},
		moduleKeyContext:  &ModuleKeyContext{},
		interfaceBindings: map[string]interfaceBinding{},
		callerStack:       nil,
		callerMap:         map[Location]bool{},
	}
}

func (c *container) addNode(provider *ProviderDescriptor, key *moduleKey) (interface{}, error) {
	providerGraphNode := c.locationGraphNode(provider.Location, key)
	hasModuleKeyParam := false
	hasOwnModuleKeyParam := false
	for _, in := range provider.Inputs {
		typ := in.Type
		if typ == moduleKeyType {
			hasModuleKeyParam = true
		}

		if typ == ownModuleKeyType {
			hasOwnModuleKeyParam = true
		}

		if isManyPerContainerType(typ) {
			return nil, fmt.Errorf("many-per-container type %v can't be used as an input parameter", typ)
		} else if isOnePerModuleType(typ) {
			return nil, fmt.Errorf("one-per-module type %v can't be used as an input parameter", typ)
		}

		vr, err := c.getResolver(typ, key)
		if err != nil {
			return nil, err
		}

		var typeGraphNode *graphviz.Node
		if vr != nil {
			typeGraphNode = vr.typeGraphNode()
		} else {
			typeGraphNode = c.typeGraphNode(typ)
			if err != nil {
				return nil, err
			}
		}

		c.addGraphEdge(typeGraphNode, providerGraphNode)
	}

	if !hasModuleKeyParam {
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

			// many-per-container slices of many-per-container types
			if isManyPerContainerSliceType(typ) {
				typ = typ.Elem()
			}

			vr, err := c.getResolver(typ, key)
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

				typeGraphNode := c.typeGraphNode(typ)
				vr = &simpleResolver{
					node:        sp,
					typ:         typ,
					graphNode:   typeGraphNode,
					idxInValues: i,
				}
				c.addResolver(typ, vr)
			}

			c.addGraphEdge(providerGraphNode, vr.typeGraphNode())
		}

		return sp, nil
	} else {
		if hasOwnModuleKeyParam {
			return nil, errors.Errorf("%T and %T must not be declared as dependencies on the same provided",
				ModuleKey{}, OwnModuleKey{})
		}

		c.logf("Registering module-scoped provider: %s", provider.Location.String())
		c.indentLogger()
		defer c.dedentLogger()

		node := &moduleDepProvider{
			provider:        provider,
			calledForModule: map[*moduleKey]bool{},
			valueMap:        map[*moduleKey][]reflect.Value{},
			valueExprs:      map[*moduleKey][]ast.Expr{},
		}

		for i, out := range provider.Outputs {
			typ := out.Type

			c.logf("Registering resolver for module-scoped type %v", typ)

			existing, ok := c.resolverByType(typ)
			if ok {
				return nil, errors.Errorf("duplicate provision of type %v by module-scoped provider %s\n\talready provided by %s",
					typ, provider.Location, existing.describeLocation())
			}

			typeGraphNode := c.typeGraphNode(typ)
			c.addResolver(typ, &moduleDepResolver{
				typ:         typ,
				idxInValues: i,
				node:        node,
				valueMap:    map[*moduleKey]reflect.Value{},
				graphNode:   typeGraphNode,
			})

			c.addGraphEdge(providerGraphNode, typeGraphNode)
		}

		return node, nil
	}
}

func (c *container) supply(value reflect.Value, loc Location, codegenVar *ast.Ident) error {
	typ := value.Type()
	locGrapNode := c.locationGraphNode(loc, nil)
	markGraphNodeAsUsed(locGrapNode)
	typeGraphNode := c.typeGraphNode(typ)
	c.addGraphEdge(locGrapNode, typeGraphNode)

	if existing, ok := c.resolverByType(typ); ok {
		return duplicateDefinitionError(typ, loc, existing.describeLocation())
	}

	r := &supplyResolver{
		typ:       typ,
		value:     value,
		loc:       loc,
		graphNode: typeGraphNode,
	}

	if codegenVar != nil {
		r.codegenDef = true
		r.varIdent = codegenVar
	} else {
		r.varIdent = c.funcGen.CreateIdent(util.StringFirstLower(typ.Name()))
	}

	c.addResolver(typ, r)

	return nil
}

func (c *container) addInvoker(provider *ProviderDescriptor, key *moduleKey) error {
	// make sure there are no outputs
	if len(provider.Outputs) > 0 {
		return fmt.Errorf("invoker function %s should not return any outputs", provider.Location)
	}

	c.invokers = append(c.invokers, invoker{
		fn:     provider,
		modKey: key,
	})

	return nil
}

func (c *container) getModuleKeyExpr(key *moduleKey) ast.Expr {
	return &ast.CallExpr{
		Fun: &ast.SelectorExpr{
			X:   c.moduleKeyContextIdent,
			Sel: ast.NewIdent("For"),
		},
		Args: []ast.Expr{&ast.BasicLit{
			Kind:  token.STRING,
			Value: fmt.Sprintf("%q", key.name),
		}},
	}
}

func (c *container) resolve(in ProviderInput, moduleKey *moduleKey, caller Location) (reflect.Value, ast.Expr, error) {
	c.resolveStack = append(c.resolveStack, resolveFrame{loc: caller, typ: in.Type})

	typeGraphNode := c.typeGraphNode(in.Type)

	if in.Type == moduleKeyType {
		if moduleKey == nil {
			return reflect.Value{}, nil, errors.Errorf("trying to resolve %T for %s but not inside of any module's scope", moduleKey, caller)
		}
		c.logf("Providing ModuleKey %s", moduleKey.name)
		markGraphNodeAsUsed(typeGraphNode)
		return reflect.ValueOf(ModuleKey{moduleKey}), c.getModuleKeyExpr(moduleKey), nil
	}

	if in.Type == ownModuleKeyType {
		if moduleKey == nil {
			return reflect.Value{}, nil, errors.Errorf("trying to resolve %T for %s but not inside of any module's scope", moduleKey, caller)
		}
		c.logf("Providing OwnModuleKey %s", moduleKey.name)
		markGraphNodeAsUsed(typeGraphNode)
		e, err := c.funcGen.TypeExpr(reflect.TypeOf(OwnModuleKey{}))
		if err != nil {
			return reflect.Value{}, nil, err
		}
		return reflect.ValueOf(OwnModuleKey{moduleKey}), &ast.CallExpr{
			Fun: e,
			Args: []ast.Expr{
				c.getModuleKeyExpr(moduleKey),
			},
		}, nil
	}

	vr, err := c.getResolver(in.Type, moduleKey)
	if err != nil {
		return reflect.Value{}, nil, err
	}

	if vr == nil {
		if in.Optional {
			c.logf("Providing zero value for optional dependency %v", in.Type)
			zero := reflect.Zero(in.Type)
			zeroExpr, err := c.funcGen.ValueExpr(zero)
			return zero, zeroExpr, err
		}

		markGraphNodeAsFailed(typeGraphNode)
		return reflect.Value{}, nil, errors.Errorf("can't resolve type %v for %s:\n%s",
			fullyQualifiedTypeName(in.Type), caller, c.formatResolveStack())
	}

	res, e, err := vr.resolve(c, moduleKey, caller)
	if err != nil {
		markGraphNodeAsFailed(typeGraphNode)
		return reflect.Value{}, nil, err
	}

	markGraphNodeAsUsed(typeGraphNode)

	c.resolveStack = c.resolveStack[:len(c.resolveStack)-1]

	return res, e, nil
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

func fullyQualifiedTypeName(typ reflect.Type) string {
	pkgType := typ
	if typ.Kind() == reflect.Pointer || typ.Kind() == reflect.Slice || typ.Kind() == reflect.Map || typ.Kind() == reflect.Array {
		pkgType = typ.Elem()
	}
	return fmt.Sprintf("%s/%v", pkgType.PkgPath(), typ)
}

func bindingKeyFromTypeName(typeName string, key *moduleKey) string {
	if key == nil {
		return fmt.Sprintf("%s;", typeName)
	}
	return fmt.Sprintf("%s;%s", typeName, key.name)
}

func bindingKeyFromType(typ reflect.Type, key *moduleKey) string {
	return bindingKeyFromTypeName(fullyQualifiedTypeName(typ), key)
}

func (c *container) addBinding(p interfaceBinding) {
	c.interfaceBindings[bindingKeyFromTypeName(p.interfaceName, p.moduleKey)] = p
}

func (c *container) addResolver(typ reflect.Type, r resolver) {
	c.resolvers[fullyQualifiedTypeName(typ)] = r
}

func (c *container) resolverByType(typ reflect.Type) (resolver, bool) {
	return c.resolverByTypeName(fullyQualifiedTypeName(typ))
}

func (c *container) resolverByTypeName(typeName string) (resolver, bool) {
	res, found := c.resolvers[typeName]
	return res, found
}

func markGraphNodeAsUsed(node *graphviz.Node) {
	node.SetColor("black")
	node.SetPenWidth("1.5")
	node.SetFontColor("black")
}

func markGraphNodeAsFailed(node *graphviz.Node) {
	node.SetColor("red")
	node.SetFontColor("red")
}
