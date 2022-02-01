package container

import (
	"reflect"
)

// ModuleKey is a special type used to scope a provider to a "module".
//
// Special module-scoped providers can be used with Provide by declaring a
// provider with an input parameter of type ModuleKey. These providers
// may construct a unique value of a dependency for each module and will
// be called at most once per module.
//
// Providers passed to ProvideInModule can also declare an input parameter
// of type ModuleKey to retrieve their module key but these providers will be
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
