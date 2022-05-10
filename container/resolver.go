package container

import (
	"reflect"

	"github.com/emicklei/dot"
)

type resolver interface {
	addNode(*simpleProvider, int) error
	resolve(*container, *moduleKey, Location) (reflect.Value, error)
	describeLocation() string
	typeGraphNode() dot.Node
}
