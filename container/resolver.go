package container

import (
	"reflect"
)

type resolver interface {
	addNode(*simpleProvider, int, *container) error
	resolve(*container, Scope, Location) (reflect.Value, error)
	describeLocation() string
}
