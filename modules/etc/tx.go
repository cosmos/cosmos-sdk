package etc

import (
	"github.com/tendermint/basecoin"
	"github.com/tendermint/go-wire/data"
)

// nolint
const (
	TypeSet    = Name + "/set"
	TypeGet    = Name + "/get"
	TypeRemove = Name + "/remove"

	ByteSet    = 0xF0
	ByteGet    = 0xF1
	ByteRemove = 0xF2
)

func init() {
	basecoin.TxMapper.
		RegisterImplementation(SetTx{}, TypeSet, ByteSet)
}

// SetTx sets a key-value pair
type SetTx struct {
	Key   data.Bytes `json:"key"`
	Value data.Bytes `json:"value"`
}

// Wrap - fulfills TxInner interface
func (t SetTx) Wrap() basecoin.Tx {
	return basecoin.Tx{t}
}

// ValidateBasic makes sure it is valid
func (t SetTx) ValidateBasic() error {
	if len(t.Key) == 0 || len(t.Value) == 0 {
		return ErrMissingData()
	}
	return nil
}
