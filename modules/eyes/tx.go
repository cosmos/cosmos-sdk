package eyes

import (
	sdk "github.com/cosmos/cosmos-sdk"
	"github.com/tendermint/go-wire/data"
)

// nolint
const (
	TypeSet    = Name + "/set"
	TypeRemove = Name + "/remove"

	ByteSet    = 0xF4
	ByteRemove = 0xF5
)

func init() {
	sdk.TxMapper.
		RegisterImplementation(SetTx{}, TypeSet, ByteSet).
		RegisterImplementation(RemoveTx{}, TypeRemove, ByteRemove)
}

// SetTx sets a key-value pair
type SetTx struct {
	Key   data.Bytes `json:"key"`
	Value data.Bytes `json:"value"`
}

func NewSetTx(key, value []byte) sdk.Tx {
	return SetTx{Key: key, Value: value}.Wrap()
}

// Wrap - fulfills TxInner interface
func (t SetTx) Wrap() sdk.Tx {
	return sdk.Tx{t}
}

// ValidateBasic makes sure it is valid
func (t SetTx) ValidateBasic() error {
	if len(t.Key) == 0 || len(t.Value) == 0 {
		return ErrMissingData()
	}
	return nil
}

// RemoveTx deletes the value at this key, returns old value
type RemoveTx struct {
	Key data.Bytes `json:"key"`
}

func NewRemoveTx(key []byte) sdk.Tx {
	return RemoveTx{Key: key}.Wrap()
}

// Wrap - fulfills TxInner interface
func (t RemoveTx) Wrap() sdk.Tx {
	return sdk.Tx{t}
}

// ValidateBasic makes sure it is valid
func (t RemoveTx) ValidateBasic() error {
	if len(t.Key) == 0 {
		return ErrMissingData()
	}
	return nil
}
