package container

import (
	reflect2 "reflect"
)

type resolver interface {
	addNode(*simpleProvider, int, *container) error
	resolve(*container, Scope, Location) (reflect2.Value, error)
}
