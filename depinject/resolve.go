package depinject

import (
	"reflect"

	"github.com/cosmos/cosmos-sdk/depinject/internal/graphviz"
)

func (c *container) getResolver(typ reflect.Type, key *moduleKey) (resolver, error) {
	pr, err := c.getExplicitResolver(typ, key)
	if err != nil {
		return nil, err
	}
	if pr != nil {
		return pr, nil
	}

	if vr, ok := c.resolverByType(typ); ok {
		return vr, nil
	}

	elemType := typ
	if isManyPerContainerSliceType(elemType) || isOnePerModuleMapType(elemType) {
		elemType = elemType.Elem()
	}

	var typeGraphNode *graphviz.Node

	if isManyPerContainerType(elemType) {
		c.logf("Registering resolver for many-per-container type %v", elemType)
		sliceType := reflect.SliceOf(elemType)

		typeGraphNode = c.typeGraphNode(sliceType)
		typeGraphNode.SetComment("many-per-container")

		r := &groupResolver{
			typ:       elemType,
			sliceType: sliceType,
			graphNode: typeGraphNode,
		}

		c.addResolver(elemType, r)
		c.addResolver(sliceType, &sliceGroupResolver{groupResolver: r})
	} else if isOnePerModuleType(elemType) {
		c.logf("Registering resolver for one-per-module type %v", elemType)
		mapType := reflect.MapOf(stringType, elemType)

		typeGraphNode = c.typeGraphNode(mapType)
		typeGraphNode.SetComment("one-per-module")

		r := &onePerModuleResolver{
			typ:       elemType,
			mapType:   mapType,
			providers: map[*moduleKey]*simpleProvider{},
			idxMap:    map[*moduleKey]int{},
			graphNode: typeGraphNode,
		}

		c.addResolver(elemType, r)
		c.addResolver(mapType, &mapOfOnePerModuleResolver{r})
	}

	res, found := c.resolverByType(typ)

	if !found && typ.Kind() == reflect.Interface {
		matches := map[reflect.Type]reflect.Type{}
		var resolverType reflect.Type
		for _, r := range c.resolvers {
			if r.getType().Kind() != reflect.Interface && r.getType().Implements(typ) {
				resolverType = r.getType()
				matches[resolverType] = resolverType
			}
		}

		if len(matches) == 1 {
			res, _ = c.resolverByType(resolverType)
			c.logf("Implicitly registering resolver %v for interface type %v", resolverType, typ)
			c.addResolver(typ, res)
		} else if len(matches) > 1 {
			return nil, newErrMultipleImplicitInterfaceBindings(typ, matches)
		}
	}

	return res, nil
}

func (c *container) getExplicitResolver(typ reflect.Type, key *moduleKey) (resolver, error) {
	var pref interfaceBinding
	var found bool

	// module scoped binding takes precedence
	pref, found = c.interfaceBindings[bindingKeyFromType(typ, key)]

	// fallback to global scope binding
	if !found {
		pref, found = c.interfaceBindings[bindingKeyFromType(typ, nil)]
	}

	if !found {
		return nil, nil
	}

	if pref.resolver != nil {
		return pref.resolver, nil
	}

	res, ok := c.resolverByTypeName(pref.implTypeName)
	if ok {
		c.logf("Registering resolver %v for interface type %v by explicit binding", res.getType(), typ)
		pref.resolver = res
		return res, nil

	}

	return nil, newErrNoTypeForExplicitBindingFound(pref)
}

var stringType = reflect.TypeOf("")
