package mock

import (
	"crypto/sha256"
	"encoding/json"
	"errors"

	gogoproto "github.com/cosmos/gogoproto/types"

	"cosmossdk.io/core/transaction"
)

var _ transaction.Tx = Tx{}

type Tx struct {
	Sender   []byte
	Msg      transaction.Msg
	GasLimit uint64
}

func (t Tx) Hash() [32]byte {
	return sha256.Sum256(t.Bytes())
}

func (t Tx) GetMessages() ([]transaction.Msg, error) {
	return []transaction.Msg{t.Msg}, nil
}

func (t Tx) GetSenders() ([]transaction.Identity, error) {
	if t.Sender == nil {
		return nil, errors.New("senders not available or are nil")
	}
	return []transaction.Identity{t.Sender}, nil
}

func (t Tx) GetGasLimit() (uint64, error) {
	return t.GasLimit, nil
}

type encodedTx struct {
	Sender   []byte         `json:"sender"`
	Msg      *gogoproto.Any `json:"message"`
	GasLimit uint64         `json:"gas_limit"`
}

func (t Tx) Bytes() []byte {
	v2Msg := t.Msg
	msg, err := gogoproto.MarshalAny(v2Msg)
	if err != nil {
		panic(err)
	}
	tx, err := json.Marshal(encodedTx{
		Sender:   t.Sender,
		Msg:      msg,
		GasLimit: t.GasLimit,
	})
	if err != nil {
		panic(err)
	}
	return tx
}

func (t *Tx) Decode(b []byte) {
	rawTx := new(encodedTx)
	err := json.Unmarshal(b, rawTx)
	if err != nil {
		panic(err)
	}
	var msg transaction.Msg
	if err := gogoproto.UnmarshalAny(rawTx.Msg, msg); err != nil {
		panic(err)
	}
	t.Msg = msg
	t.Sender = rawTx.Sender
	t.GasLimit = rawTx.GasLimit
}

func (t *Tx) DecodeJSON(b []byte) {
	rawTx := new(encodedTx)
	err := json.Unmarshal(b, rawTx)
	if err != nil {
		panic(err)
	}
	var msg transaction.Msg
	if err := gogoproto.UnmarshalAny(rawTx.Msg, msg); err != nil {
		panic(err)
	}
	t.Msg = msg
	t.Sender = rawTx.Sender
	t.GasLimit = rawTx.GasLimit
}

type TxCodec struct{}

func (TxCodec) Decode(bytes []byte) (Tx, error) {
	t := new(Tx)
	t.Decode(bytes)
	return *t, nil
}

func (TxCodec) DecodeJSON(bytes []byte) (Tx, error) {
	t := new(Tx)
	t.DecodeJSON(bytes)
	return *t, nil
}
