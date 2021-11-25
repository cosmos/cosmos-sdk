//nolint
package mock

import (
	"fmt"
	"math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/middleware"
)

// kvstoreTx defines a tx for mock purposes. The `key` and `value` fields will
// set those bytes in the kvstore, and the `bytes` field represents its
// GetSignBytes value.
type kvstoreTx struct {
	key   []byte
	value []byte
	bytes []byte
}

// dummy implementation of proto.Message
func (msg *kvstoreTx) Reset()         {}
func (msg *kvstoreTx) String() string { return "TODO" }
func (msg *kvstoreTx) ProtoMessage()  {}

var _ sdk.Tx = &kvstoreTx{}
var _ sdk.Msg = &kvstoreTx{}
var _ middleware.GasTx = &kvstoreTx{}

func NewTx(key, value string) kvstoreTx {
	bytes := fmt.Sprintf("%s=%s", key, value)
	return kvstoreTx{
		key:   []byte(key),
		value: []byte(value),
		bytes: []byte(bytes),
	}
}

func (tx kvstoreTx) Route() string {
	return "kvstore"
}

func (tx kvstoreTx) Type() string {
	return "kvstore_tx"
}

func (tx *kvstoreTx) GetMsgs() []sdk.Msg {
	return []sdk.Msg{tx}
}

func (tx kvstoreTx) GetMemo() string {
	return ""
}

func (tx kvstoreTx) GetSignBytes() []byte {
	return tx.bytes
}

// Should the app be calling this? Or only handlers?
func (tx kvstoreTx) ValidateBasic() error {
	return nil
}

func (tx kvstoreTx) GetSigners() []sdk.AccAddress {
	return nil
}

func (tx kvstoreTx) GetGas() uint64 {
	return math.MaxUint64
}
