package container

import (
	reflect2 "reflect"

	"github.com/cosmos/cosmos-sdk/container/reflect"
)

type resolver interface {
	addNode(*simpleProvider, int) error
	resolve(*container, Scope, reflect.Location) (reflect2.Value, error)
}
