package types

import (
	"reflect"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Subspace struct {
	cdc  *codec.Codec
	key  sdk.StoreKey // []byte -> []byte, stores parameter
	tkey sdk.StoreKey // []byte -> bool, stores parameter change

	name []byte

	table KeyTable
}

type attribute struct {
	ty reflect.Type
}

type KeyTable struct {
	m map[string]attribute
}
