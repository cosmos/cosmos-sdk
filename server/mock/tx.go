package mock

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// An sdk.Tx which is its own sdk.Msg.
type kvstoreTx struct {
	key   []byte
	value []byte
	bytes []byte
}

// dummy implementation of proto.Message
func (msg *kvstoreTx) Reset()         {}
func (msg *kvstoreTx) String() string { return "TODO" }
func (msg *kvstoreTx) ProtoMessage()  {}

var (
	_ sdk.Tx  = &kvstoreTx{}
	_ sdk.Msg = &kvstoreTx{}
)

func NewTx(key, value string) *kvstoreTx {
	bytes := fmt.Sprintf("%s=%s", key, value)
	return &kvstoreTx{
		key:   []byte(key),
		value: []byte(value),
		bytes: []byte(bytes),
	}
}

func (tx *kvstoreTx) Type() string {
	return "kvstore_tx"
}

func (tx *kvstoreTx) GetMsgs() []sdk.Msg {
	return []sdk.Msg{tx}
}

func (tx *kvstoreTx) GetSignBytes() []byte {
	return tx.bytes
}

// Should the app be calling this? Or only handlers?
func (tx *kvstoreTx) ValidateBasic() error {
	return nil
}

func (tx *kvstoreTx) GetSigners() []sdk.AccAddress {
	return nil
}
