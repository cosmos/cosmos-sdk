package params

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/params/space"
)

// nolint - reexport
type Space = space.Space
type ReadOnlySpace = space.ReadOnlySpace
type ParamStruct = space.ParamStruct
type KeyFieldPairs = space.KeyFieldPairs

// nolint - reexport
func UnmarshalParamsFromMap(m map[string][]byte, cdc *codec.Codec, ps space.ParamStruct) error {
	return space.UnmarshalParamsFromMap(m, cdc, ps)
}
