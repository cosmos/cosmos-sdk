package params

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/params/store"
)

// re-export types from store
type (
	Store         = store.Store
	ReadOnlyStore = store.ReadOnlyStore
	ParamStruct   = store.ParamStruct
	KeyFieldPairs = store.KeyFieldPairs
)

// UnmarshalParamsFromMap deserializes parameters from a given map. It returns
// an error upon failure.
func UnmarshalParamsFromMap(m map[string][]byte, cdc *codec.Codec, ps store.ParamStruct) error {
	return store.UnmarshalParamsFromMap(m, cdc, ps)
}
