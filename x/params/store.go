package params

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/params/store"
)

// nolint - reexport
type Store = store.Store
type ReadOnlyStore = store.ReadOnlyStore
type ParamStruct = store.ParamStruct
type KeyFieldPairs = store.KeyFieldPairs

// nolint - reexport
func UnmarshalParamsFromMap(m map[string][]byte, cdc *codec.Codec, ps store.ParamStruct) error {
	return store.UnmarshalParamsFromMap(m, cdc, ps)
}
