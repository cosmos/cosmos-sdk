package container

import (
	"reflect"

	"github.com/goccy/go-graphviz/cgraph"
)

type resolver interface {
	addNode(*simpleProvider, int) error
	resolve(*container, *moduleKey, Location) (reflect.Value, error)
	describeLocation() string
	typeGraphNode() *cgraph.Node
}
