package types

import (
	"github.com/tendermint/go-crypto"
	"github.com/tendermint/go-wire"
	gov "github.com/tendermint/governmint/types"
)

type Input struct {
	PubKey    crypto.PubKey
	Amount    uint64
	Sequence  uint
	Signature crypto.Signature
}

type Output struct {
	PubKey crypto.PubKey
	Amount uint64
}

type SendTx struct {
	Inputs  []Input
	Outputs []Output
}

func (tx *SendTx) SignBytes() []byte {
	sigs := make([]crypto.Signature, len(tx.Inputs))
	for i, input := range tx.Inputs {
		sigs[i] = input.Signature
		input.Signature = nil
		tx.Inputs[i] = input
	}
	signBytes := wire.BinaryBytes(tx)
	for i := range tx.Inputs {
		tx.Inputs[i].Signature = sigs[i]
	}
	return signBytes
}

func (tx *SendTx) GetInputs() []Input   { return tx.Inputs }
func (tx *SendTx) GetOutputs() []Output { return tx.Outputs }

type GovTx struct {
	Input Input
	Tx    gov.Tx
}

func (tx *GovTx) SignBytes() []byte {
	sig := tx.Input.Signature
	tx.Input.Signature = nil
	signBytes := wire.BinaryBytes(tx)
	tx.Input.Signature = sig
	return signBytes
}

func (tx *GovTx) GetInputs() []Input   { return []Input{tx.Input} }
func (tx *GovTx) GetOutputs() []Output { return nil }

type Tx interface {
	AssertIsTx()
	SignBytes() []byte
	GetInputs() []Input
	GetOutputs() []Output
}

func (_ *SendTx) AssertIsTx() {}
func (_ *GovTx) AssertIsTx()  {}

const (
	TxTypeSend = byte(0x01)
	TxTypeGov  = byte(0x02)
)

var _ = wire.RegisterInterface(
	struct{ Tx }{},
	wire.ConcreteType{&SendTx{}, TxTypeSend},
	wire.ConcreteType{&GovTx{}, TxTypeGov},
)

//----------------------------------------

type Account struct {
	Sequence uint
	Balance  uint64
}

type PubAccount struct {
	crypto.PubKey
	Account
}

type PrivAccount struct {
	crypto.PubKey
	crypto.PrivKey
	Account
}

type GenesisState struct {
	Accounts []PubAccount
}
