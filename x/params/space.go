package params

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/params/space"
)

// nolint - reexport
type Space = space.Space
type ReadOnlySpace = space.ReadOnlySpace
type Key = space.Key
type KeyFieldPair = space.KeyFieldPair
type KeyFieldPairs = space.KeyFieldPairs
type ParamStruct = space.ParamStruct

// nolint - reexport
func NewKey(keys ...[]byte) Key {
	return space.NewKey(keys...)
}
func UnmarshalParamsFromMap(m map[string][]byte, cdc *codec.Codec, ps space.ParamStruct) error {
	return space.UnmarshalParamsFromMap(m, cdc, ps)
}
