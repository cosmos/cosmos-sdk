package mock

import (
	"crypto/sha256"
	"encoding/json"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"

	"cosmossdk.io/server/v2/core/transaction"
)

var _ transaction.Tx = Tx{}

type Tx struct {
	Sender   []byte
	Msg      proto.Message
	GasLimit uint64
}

func (t Tx) Hash() [32]byte {
	return sha256.Sum256(t.Bytes())
}

func (t Tx) GetMessages() []transaction.Type {
	return []transaction.Type{t.Msg}
}

func (t Tx) GetSenders() []transaction.Identity {
	return []transaction.Identity{t.Sender}
}

func (t Tx) GetGasLimit() uint64 {
	return t.GasLimit
}

type encodedTx struct {
	Sender   []byte     `json:"sender"`
	Msg      *anypb.Any `json:"message"`
	GasLimit uint64     `json:"gas_limit"`
}

func (t Tx) Bytes() []byte {
	msg, err := anypb.New(t.Msg)
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
	msg, err := rawTx.Msg.UnmarshalNew()
	if err != nil {
		panic(err)
	}
	t.Msg = msg
	t.Sender = rawTx.Sender
	t.GasLimit = rawTx.GasLimit
}

type txCodec struct{}

func (txCodec) Decode(bytes []byte) (Tx, error) {
	t := new(Tx)
	t.Decode(bytes)
	return *t, nil
}
