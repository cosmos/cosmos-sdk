package main

import (
	"bytes"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	crypto "github.com/tendermint/go-crypto"
)

type dummyTx struct {
	key   []byte
	value []byte
	bytes []byte
}

func (tx dummyTx) Get(key interface{}) (value interface{}) {
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

func (tx dummyTx) Type() string {
	return "dummy"
}

func (tx dummyTx) GetSignBytes() []byte {
	return tx.bytes
}

// Should the app be calling this? Or only handlers?
func (tx dummyTx) ValidateBasic() error {
	return nil
}

func (tx dummyTx) GetSigners() []crypto.Address {
	return nil
}

func (tx dummyTx) GetSignatures() []sdk.StdSignature {
	return nil
}

func (tx dummyTx) GetFeePayer() crypto.Address {
	return nil
}

func decodeTx(txBytes []byte) (sdk.Tx, error) {
	var tx sdk.Tx

	split := bytes.Split(txBytes, []byte("="))
	if len(split) == 1 {
		k := split[0]
		tx = dummyTx{k, k, txBytes}
	} else if len(split) == 2 {
		k, v := split[0], split[1]
		tx = dummyTx{k, v, txBytes}
	} else {
		return nil, fmt.Errorf("too many =")
	}

	return tx, nil
}
