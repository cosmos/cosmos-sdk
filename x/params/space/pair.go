package space

import (
	"github.com/cosmos/cosmos-sdk/codec"
)

// Used for associating paramstore key and field of param structs
type KeyFieldPair struct {
	Key   string
	Field interface{}
}

// Slice of KeyFieldPair
type KeyFieldPairs []KeyFieldPair

// Interface for structs containing parameters for a module
type ParamStruct interface {
	KeyFieldPairs() KeyFieldPairs
}

// Takes a map from key string to byte slice and
// unmarshalles it to ParamStruct
func UnmarshalParamsFromMap(m map[string][]byte, cdc *codec.Codec, ps ParamStruct) error {
	for _, p := range ps.KeyFieldPairs() {
		err := cdc.UnmarshalJSON(m[p.Key], p.Field)
		if err != nil {
			return err
		}
	}
	return nil
}
