package params

import (
	"github.com/cosmos/cosmos-sdk/x/params/space"
)

// nolint - reexport
type Space = space.Space
type ReadOnlySpace = space.ReadOnlySpace
type Key = space.Key

// nolint - reexport
func NewKey(keys ...string) Key {
	return space.NewKey(keys...)
}
