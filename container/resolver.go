package container

import (
	"reflect"

	"github.com/cosmos/cosmos-sdk/container/internal/graphviz"
)

type resolver interface {
	addNode(*simpleProvider, int) error
	resolve(*container, *moduleKey, Location) (reflect.Value, error)
	describeLocation() string
	typeGraphNode() *graphviz.Node
}
