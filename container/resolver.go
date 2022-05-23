package container

import (
	"reflect"

	"cosmossdk.io/container/internal/graphviz"
)

type resolver interface {
	addNode(*simpleProvider, int) error
	resolve(*container, *moduleKey, Location) (reflect.Value, error)
	describeLocation() string
	typeGraphNode() *graphviz.Node
}
