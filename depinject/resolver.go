package depinject

import (
	"reflect"

	"github.com/cosmos/cosmos-sdk/depinject/v2/internal/graphviz"
)

type resolver interface {
	addNode(*simpleProvider, int) error
	resolve(*container, *moduleKey, Location) (reflect.Value, error)
	describeLocation() string
	typeGraphNode() *graphviz.Node
	getType() reflect.Type
}
