package main

import (
	"bytes"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// An sdk.Tx which is its own sdk.Msg.
type kvstoreTx struct {
	key   []byte
	value []byte
	bytes []byte
}

func (tx kvstoreTx) Get(key interface{}) (value interface{}) {
	switch k := key.(type) {
	case string:
		switch k {
		case "key":
			return tx.key
		case "value":
			return tx.value
		}
	}
	return nil
}

func (tx kvstoreTx) Type() string {
	return "kvstore"
}

func (tx kvstoreTx) GetMsg() sdk.Msg {
	return tx
}

func (tx kvstoreTx) GetSignBytes() []byte {
	return tx.bytes
}

// Should the app be calling this? Or only handlers?
func (tx kvstoreTx) ValidateBasic() sdk.Error {
	return nil
}

func (tx kvstoreTx) GetSigners() []sdk.Address {
	return nil
}

func (tx kvstoreTx) GetSignatures() []sdk.StdSignature {
	return nil
}

// takes raw transaction bytes and decodes them into an sdk.Tx. An sdk.Tx has
// all the signatures and can be used to authenticate.
func decodeTx(txBytes []byte) (sdk.Tx, sdk.Error) {
	var tx sdk.Tx

	split := bytes.Split(txBytes, []byte("="))
	if len(split) == 1 {
		k := split[0]
		tx = kvstoreTx{k, k, txBytes}
	} else if len(split) == 2 {
		k, v := split[0], split[1]
		tx = kvstoreTx{k, v, txBytes}
	} else {
		return nil, sdk.ErrTxParse("too many =")
	}

	return tx, nil
}
