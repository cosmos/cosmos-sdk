package mock

import (
	"bytes"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// An sdk.Tx which is its own sdk.Msg.
type KvstoreMsg struct {
	Key   []byte
	Value []byte
}

func NewKvstoreMsg(key []byte, value []byte) KvstoreMsg {
	return KvstoreMsg{
		Key:   key,
		Value: value,
	}
}

func (msg KvstoreMsg) Type() string {
	return "kvstore"
}

func (msg KvstoreMsg) Get(key interface{}) (value interface{}) {
	switch k := key.(type) {
	case string:
		switch k {
		case "key":
			return msg.Key
		case "value":
			return msg.Value
		}
	}
	return nil
}

func (msg KvstoreMsg) GetSignBytes() []byte {
	return nil
}

// Should the app be calling this? Or only handlers?
func (msg KvstoreMsg) ValidateBasic() sdk.Error {
	return nil
}

func (msg KvstoreMsg) GetSigners() []sdk.Address {
	return nil
}

// takes raw transaction bytes and decodes them into an sdk.Tx. An sdk.Tx has
// all the signatures and can be used to authenticate.
func decodeTx(txBytes []byte) (sdk.Tx, sdk.Error) {
	var msg sdk.Msg

	split := bytes.Split(txBytes, []byte("="))
	if len(split) == 1 {
		k := split[0]
		msg = KvstoreMsg{k, k}
	} else if len(split) == 2 {
		k, v := split[0], split[1]
		msg = KvstoreMsg{k, v}
	} else {
		return sdk.Tx{}, sdk.ErrTxDecode("too many =")
	}

	tx := sdk.NewTx(msg, sdk.StdFee{}, nil)

	return tx, nil
}
