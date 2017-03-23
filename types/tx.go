package types

import (
	"bytes"
	"encoding/json"

	abci "github.com/tendermint/abci/types"
	. "github.com/tendermint/go-common"
	"github.com/tendermint/go-crypto"
	"github.com/tendermint/go-data"
	"github.com/tendermint/go-wire"
)

/*
Tx (Transaction) is an atomic operation on the ledger state.

Account Types:
 - SendTx         Send coins to address
 - AppTx         Send a msg to a contract that runs in the vm
*/
type Tx interface {
	AssertIsTx()
	SignBytes(chainID string) []byte
}

// Types of Tx implementations
const (
	// Account transactions
	TxTypeSend = byte(0x01)
	TxTypeApp  = byte(0x02)
	TxNameSend = "send"
	TxNameApp  = "app"
)

func (_ *SendTx) AssertIsTx() {}
func (_ *AppTx) AssertIsTx()  {}

var txMapper data.Mapper

// register both private key types with go-data (and thus go-wire)
func init() {
	txMapper = data.NewMapper(TxS{}).
		RegisterInterface(&SendTx{}, TxNameSend, TxTypeSend).
		RegisterInterface(&AppTx{}, TxNameApp, TxTypeApp)
}

// TxS add json serialization to Tx
type TxS struct {
	Tx
}

func (p TxS) MarshalJSON() ([]byte, error) {
	return txMapper.ToJSON(p.Tx)
}

func (p *TxS) UnmarshalJSON(data []byte) (err error) {
	parsed, err := txMapper.FromJSON(data)
	if err == nil {
		p.Tx = parsed.(Tx)
	}
	return
}

//-----------------------------------------------------------------------------

type TxInput struct {
	Address    data.Bytes        `json:"address"`   // Hash of the PubKey
	Coins      Coins             `json:"coins"`     //
	Sequence   int               `json:"sequence"`  // Must be 1 greater than the last committed TxInput
	SignatureS crypto.SignatureS `json:"signature"` // Depends on the PubKey type and the whole Tx
	PubKey     crypto.PubKeyS    `json:"pub_key"`   // Is present iff Sequence == 0
}

func (txIn TxInput) ValidateBasic() abci.Result {
	if len(txIn.Address) != 20 {
		return abci.ErrBaseInvalidInput.AppendLog("Invalid address length")
	}
	if !txIn.Coins.IsValid() {
		return abci.ErrBaseInvalidInput.AppendLog(Fmt("Invalid coins %v", txIn.Coins))
	}
	if txIn.Coins.IsZero() {
		return abci.ErrBaseInvalidInput.AppendLog("Coins cannot be zero")
	}
	if txIn.Sequence <= 0 {
		return abci.ErrBaseInvalidInput.AppendLog("Sequence must be greater than 0")
	}
	if txIn.Sequence == 1 && txIn.PubKey.Empty() {
		return abci.ErrBaseInvalidInput.AppendLog("PubKey must be present when Sequence == 1")
	}
	if txIn.Sequence > 1 && !txIn.PubKey.Empty() {
		return abci.ErrBaseInvalidInput.AppendLog("PubKey must be nil when Sequence > 1")
	}
	return abci.OK
}

func (txIn TxInput) String() string {
	return Fmt("TxInput{%X,%v,%v,%v,%v}", txIn.Address, txIn.Coins, txIn.Sequence, txIn.SignatureS, txIn.PubKey)
}

func NewTxInput(pubKey crypto.PubKey, coins Coins, sequence int) TxInput {
	input := TxInput{
		Address:  pubKey.Address(),
		Coins:    coins,
		Sequence: sequence,
	}
	if sequence == 1 {
		// safely wrap if needed
		// TODO: extract this as utility function?
		ps, ok := pubKey.(crypto.PubKeyS)
		if !ok {
			ps = crypto.PubKeyS{pubKey}
		}
		input.PubKey = ps
	}
	return input
}

//-----------------------------------------------------------------------------

type TxOutput struct {
	Address data.Bytes `json:"address"` // Hash of the PubKey
	Coins   Coins      `json:"coins"`   //
}

func (txOut TxOutput) ValidateBasic() abci.Result {
	if len(txOut.Address) != 20 {
		return abci.ErrBaseInvalidOutput.AppendLog("Invalid address length")
	}
	if !txOut.Coins.IsValid() {
		return abci.ErrBaseInvalidOutput.AppendLog(Fmt("Invalid coins %v", txOut.Coins))
	}
	if txOut.Coins.IsZero() {
		return abci.ErrBaseInvalidOutput.AppendLog("Coins cannot be zero")
	}
	return abci.OK
}

func (txOut TxOutput) String() string {
	return Fmt("TxOutput{%X,%v}", txOut.Address, txOut.Coins)
}

//-----------------------------------------------------------------------------

type SendTx struct {
	Gas     int64      `json:"gas"` // Gas
	Fee     Coin       `json:"fee"` // Fee
	Inputs  []TxInput  `json:"inputs"`
	Outputs []TxOutput `json:"outputs"`
}

func (tx *SendTx) SignBytes(chainID string) []byte {
	signBytes := wire.BinaryBytes(chainID)
	sigz := make([]crypto.Signature, len(tx.Inputs))
	for i, input := range tx.Inputs {
		sigz[i] = input.SignatureS.Signature
		tx.Inputs[i].SignatureS.Signature = nil
	}
	signBytes = append(signBytes, wire.BinaryBytes(tx)...)
	for i := range tx.Inputs {
		tx.Inputs[i].SignatureS.Signature = sigz[i]
	}
	return signBytes
}

func (tx *SendTx) SetSignature(addr []byte, sig crypto.Signature) bool {
	sigs, ok := sig.(crypto.SignatureS)
	if !ok {
		sigs = crypto.SignatureS{sig}
	}
	for i, input := range tx.Inputs {
		if bytes.Equal(input.Address, addr) {
			tx.Inputs[i].SignatureS = sigs
			return true
		}
	}
	return false
}

func (tx *SendTx) String() string {
	return Fmt("SendTx{%v/%v %v->%v}", tx.Gas, tx.Fee, tx.Inputs, tx.Outputs)
}

//-----------------------------------------------------------------------------

type AppTx struct {
	Gas   int64           `json:"gas"`   // Gas
	Fee   Coin            `json:"fee"`   // Fee
	Name  string          `json:"type"`  // Which plugin
	Input TxInput         `json:"input"` // Hmmm do we want coins?
	Data  json.RawMessage `json:"data"`
}

func (tx *AppTx) SignBytes(chainID string) []byte {
	signBytes := wire.BinaryBytes(chainID)
	sig := tx.Input.SignatureS
	tx.Input.SignatureS.Signature = nil
	signBytes = append(signBytes, wire.BinaryBytes(tx)...)
	tx.Input.SignatureS = sig
	return signBytes
}

func (tx *AppTx) SetSignature(sig crypto.Signature) bool {
	sigs, ok := sig.(crypto.SignatureS)
	if !ok {
		sigs = crypto.SignatureS{sig}
	}
	tx.Input.SignatureS = sigs
	return true
}

func (tx *AppTx) String() string {
	return Fmt("AppTx{%v/%v %v %v %X}", tx.Gas, tx.Fee, tx.Name, tx.Input, tx.Data)
}

//-----------------------------------------------------------------------------

func TxID(chainID string, tx Tx) []byte {
	signBytes := tx.SignBytes(chainID)
	return wire.BinaryRipemd160(signBytes)
}

//--------------------------------------------------------------------------------

// Contract: This function is deterministic and completely reversible.
func jsonEscape(str string) string {
	escapedBytes, err := json.Marshal(str)
	if err != nil {
		PanicSanity(Fmt("Error json-escaping a string", str))
	}
	return string(escapedBytes)
}
