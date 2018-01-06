package eyes

import (
	"github.com/tendermint/go-wire/data"
)

// DO NOT USE THIS INTERFACE.
// You probably want to use EyesTx
// +gen wrapper:"EyesTx,Impl[SetTx,RemoveTx],set,remove"
type EyesTxInner interface {
	ValidateBasic() error
}

// func init() {
// 	sdk.TxMapper.
// 		RegisterImplementation(SetTx{}, TypeSet, ByteSet).
// 		RegisterImplementation(RemoveTx{}, TypeRemove, ByteRemove)
// }

// SetTx sets a key-value pair
type SetTx struct {
	Key   data.Bytes `json:"key"`
	Value data.Bytes `json:"value"`
}

func NewSetTx(key, value []byte) EyesTx {
	return SetTx{Key: key, Value: value}.Wrap()
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

func NewRemoveTx(key []byte) EyesTx {
	return RemoveTx{Key: key}.Wrap()
}

// ValidateBasic makes sure it is valid
func (t RemoveTx) ValidateBasic() error {
	if len(t.Key) == 0 {
		return ErrMissingData()
	}
	return nil
}
