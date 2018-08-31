package space

import (
	tmlibs "github.com/tendermint/tendermint/libs/common"

	"github.com/cosmos/cosmos-sdk/codec"
)

// Wrapper for key string
type Key struct {
	s string
}

// Appending two keys with '/' as separator
// Checks alpanumericity
func (k Key) Append(keys ...string) (res Key) {
	res = k

	for _, key := range keys {
		if !tmlibs.IsASCIIText(key) {
			panic("parameter key expressions can only contain alphanumeric characters")
		}
		res.s = res.s + "/" + key
	}
	return
}

// NewKey constructs a key from a list of strings
func NewKey(keys ...string) (res Key) {
	if len(keys) < 1 {
		panic("length of parameter keys must not be zero")
	}
	res = Key{keys[0]}

	return res.Append(keys[1:]...)
}

// KeyBytes make KVStore key bytes from Key
func (k Key) Bytes() []byte {
	return []byte(k.s)
}

// Human readable string
func (k Key) String() string {
	return k.s
}

// Used for associating paramstore key and field of param structs
type KeyFieldPair struct {
	Key   Key
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
		err := cdc.UnmarshalJSON(m[p.Key.String()], p.Field)
		if err != nil {
			return err
		}
	}
	return nil
}
