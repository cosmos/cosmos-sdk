package space

import (
	"github.com/cosmos/cosmos-sdk/codec"
)

// Wrapper for key string
type Key struct {
	s []byte
}

// Appending two keys with '/' as separator
// Checks alphanumericity
func (k Key) Append(keys ...[]byte) (res Key) {
	res.s = make([]byte, len(k.s))
	copy(res.s, k.s)

	for _, key := range keys {
		for _, b := range key {
			if !(32 <= b && b <= 126) {
				panic("parameter key expressions can only contain alphanumeric characters")
			}
		}
		res.s = append(append(res.s, byte('/')), key...)
	}
	return
}

// NewKey constructs a key from a list of strings
func NewKey(keys ...[]byte) (res Key) {
	if len(keys) < 1 {
		panic("length of parameter keys must not be zero")
	}
	res = Key{[]byte(keys[0])}

	return res.Append(keys[1:]...)
}

// KeyBytes make KVStore key bytes from Key
func (k Key) Bytes() []byte {
	return k.s
}

// Human readable string
func (k Key) String() string {
	return string(k.s)
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
