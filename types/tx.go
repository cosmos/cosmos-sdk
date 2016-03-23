package types

import (
	"bytes"
	"encoding/json"

	. "github.com/tendermint/go-common"
	"github.com/tendermint/go-crypto"
	"github.com/tendermint/go-wire"
	tmsp "github.com/tendermint/tmsp/types"
	"golang.org/x/crypto/ripemd160"
)

/*
Tx (Transaction) is an atomic operation on the ledger state.

Account Types:
 - SendTx         Send coins to address
 - CallTx         Send a msg to a contract that runs in the vm
*/

type Tx interface {
	SignBytes(chainID string) []byte
}

// Types of Tx implementations
const (
	// Account transactions
	TxTypeSend = byte(0x01)
	TxTypeCall = byte(0x02)
)

var _ = wire.RegisterInterface(
	struct{ Tx }{},
	wire.ConcreteType{&SendTx{}, TxTypeSend},
	wire.ConcreteType{&CallTx{}, TxTypeCall},
)

//-----------------------------------------------------------------------------

type TxInput struct {
	Address   []byte           `json:"address"`   // Hash of the PubKey
	Amount    int64            `json:"amount"`    // Must not exceed account balance
	Sequence  int              `json:"sequence"`  // Must be 1 greater than the last committed TxInput
	Signature crypto.Signature `json:"signature"` // Depends on the PubKey type and the whole Tx
	PubKey    crypto.PubKey    `json:"pub_key"`   // May be nil
}

func (txIn TxInput) ValidateBasic() tmsp.Result {
	if len(txIn.Address) != 20 {
		return tmsp.ErrBaseInvalidAddress
	}
	if txIn.Amount == 0 {
		return tmsp.ErrBaseInvalidAmount
	}
	return tmsp.OK
}

func (txIn TxInput) SignBytes() []byte {
	return []byte(Fmt(`{"address":"%X","amount":%v,"sequence":%v}`,
		txIn.Address, txIn.Amount, txIn.Sequence))
}

func (txIn TxInput) String() string {
	return Fmt("TxInput{%X,%v,%v,%v,%v}", txIn.Address, txIn.Amount, txIn.Sequence, txIn.Signature, txIn.PubKey)
}

//-----------------------------------------------------------------------------

type TxOutput struct {
	Address []byte `json:"address"` // Hash of the PubKey
	Amount  int64  `json:"amount"`  // The sum of all outputs must not exceed the inputs.
}

func (txOut TxOutput) ValidateBasic() tmsp.Result {
	if len(txOut.Address) != 20 {
		return tmsp.ErrBaseInvalidAddress
	}
	if txOut.Amount == 0 {
		return tmsp.ErrBaseInvalidAmount
	}
	return tmsp.OK
}

func (txOut TxOutput) SignBytes() []byte {
	return []byte(Fmt(`{"address":"%X","amount":%v}`,
		txOut.Address, txOut.Amount))
}

func (txOut TxOutput) String() string {
	return Fmt("TxOutput{%X,%v}", txOut.Address, txOut.Amount)
}

//-----------------------------------------------------------------------------

type SendTx struct {
	Inputs  []TxInput  `json:"inputs"`
	Outputs []TxOutput `json:"outputs"`
}

func (tx *SendTx) SignBytes(chainID string) []byte {
	var buf = new(bytes.Buffer)
	buf.Write([]byte(Fmt(`{"chain_id":%s`, jsonEscape(chainID))))
	buf.Write([]byte(Fmt(`,"tx":[%v,{"inputs":[`, TxTypeSend)))
	for i, in := range tx.Inputs {
		buf.Write(in.SignBytes())
		if i != len(tx.Inputs)-1 {
			buf.Write([]byte(","))
		}
	}
	buf.Write([]byte(`],"outputs":[`))
	for i, out := range tx.Outputs {
		buf.Write(out.SignBytes())
		if i != len(tx.Outputs)-1 {
			buf.Write([]byte(","))
		}
	}
	buf.Write([]byte(`]}]}`))
	return buf.Bytes()
}

func (tx *SendTx) String() string {
	return Fmt("SendTx{%v -> %v}", tx.Inputs, tx.Outputs)
}

//-----------------------------------------------------------------------------

type CallTx struct {
	Input    TxInput `json:"input"`
	Address  []byte  `json:"address"`
	GasLimit int64   `json:"gas_limit"`
	Fee      int64   `json:"fee"`
	Data     []byte  `json:"data"`
}

func (tx *CallTx) SignBytes(chainID string) []byte {
	var buf = new(bytes.Buffer)
	buf.Write([]byte(Fmt(`{"chain_id":%s`, jsonEscape(chainID))))
	buf.Write([]byte(Fmt(`,"tx":[%v,{"address":"%X","data":"%X"`, TxTypeCall, tx.Address, tx.Data)))
	buf.Write([]byte(Fmt(`,"fee":%v,"gas_limit":%v,"input":`, tx.Fee, tx.GasLimit)))
	buf.Write(tx.Input.SignBytes())
	buf.Write([]byte(`}]}`))
	return buf.Bytes()
}

func (tx *CallTx) String() string {
	return Fmt("CallTx{%v -> %x: %x}", tx.Input, tx.Address, tx.Data)
}

func NewContractAddress(caller []byte, nonce int) []byte {
	temp := make([]byte, 32+8)
	copy(temp, caller)
	PutInt64BE(temp[32:], int64(nonce))
	hasher := ripemd160.New()
	hasher.Write(temp) // does not error
	return hasher.Sum(nil)
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
