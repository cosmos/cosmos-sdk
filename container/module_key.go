package container

import (
	"reflect"
)

// ModuleKey is a special type used to scope a provider to a "module".
//
// Special module-scoped constructors can be used with Provide by declaring a
// constructor with an input parameter of type ModuleKey. These constructors
// may construct a unique value of a dependency for each module and will
// be called at most once per module.
//
// Constructors passed to ProvideInModule can also declare an input parameter
// of type ModuleKey to retrieve their module key but these constructors will be
// called at most once.
type ModuleKey struct {
	*moduleKey
}

type moduleKey struct {
	name string
}

func (k ModuleKey) Name() string {
	return k.name
}

var moduleKeyType = reflect.TypeOf(ModuleKey{})

var stringType = reflect.TypeOf("")
