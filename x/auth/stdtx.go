package auth

import (
	"encoding/json"

	sdk "github.com/cosmos/cosmos-sdk/types"
	crypto "github.com/tendermint/go-crypto"
)

var _ sdk.Tx = (*StdTx)(nil)

// StdTx is a standard way to wrap a Msg with Fee and Signatures.
// NOTE: the first signature is the FeePayer (Signatures must not be nil).
type StdTx struct {
	Msg        sdk.Msg        `json:"msg"`
	Fee        StdFee         `json:"fee"`
	Signatures []StdSignature `json:"signatures"`
}

func NewStdTx(msg sdk.Msg, fee StdFee, sigs []StdSignature) StdTx {
	return StdTx{
		Msg:        msg,
		Fee:        fee,
		Signatures: sigs,
	}
}

//nolint
func (tx StdTx) GetMsg() sdk.Msg { return tx.Msg }

// Signatures returns the signature of signers who signed the Msg.
// CONTRACT: Length returned is same as length of
// pubkeys returned from MsgKeySigners, and the order
// matches.
// CONTRACT: If the signature is missing (ie the Msg is
// invalid), then the corresponding signature is
// .Empty().
func (tx StdTx) GetSignatures() []StdSignature { return tx.Signatures }

// FeePayer returns the address responsible for paying the fees
// for the transactions. It's the first address returned by msg.GetSigners().
// If GetSigners() is empty, this panics.
func FeePayer(tx sdk.Tx) sdk.Address {
	return tx.GetMsg().GetSigners()[0]
}

//__________________________________________________________

// StdFee includes the amount of coins paid in fees and the maximum
// gas to be used by the transaction. The ratio yields an effective "gasprice",
// which must be above some miminum to be accepted into the mempool.
type StdFee struct {
	Amount sdk.Coins `json:"amount"`
	Gas    int64     `json:"gas"`
}

func NewStdFee(gas int64, amount ...sdk.Coin) StdFee {
	return StdFee{
		Amount: amount,
		Gas:    gas,
	}
}

// fee bytes for signing later
func (fee StdFee) Bytes() []byte {
	// normalize. XXX
	// this is a sign of something ugly
	// (in the lcd_test, client side its null,
	// server side its [])
	if len(fee.Amount) == 0 {
		fee.Amount = sdk.Coins{}
	}
	bz, err := msgCdc.MarshalJSON(fee) // TODO
	if err != nil {
		panic(err)
	}
	return bz
}

//__________________________________________________________

// StdSignDoc is replay-prevention structure.
// It includes the result of msg.GetSignBytes(),
// as well as the ChainID (prevent cross chain replay)
// and the Sequence numbers for each signature (prevent
// inchain replay and enforce tx ordering per account).
type StdSignDoc struct {
	ChainID        string  `json:"chain_id"`
	AccountNumbers []int64 `json:"account_numbers"`
	Sequences      []int64 `json:"sequences"`
	FeeBytes       []byte  `json:"fee_bytes"`
	MsgBytes       []byte  `json:"msg_bytes"`
	AltBytes       []byte  `json:"alt_bytes"`
}

// StdSignBytes returns the bytes to sign for a transaction.
// TODO: change the API to just take a chainID and StdTx ?
func StdSignBytes(chainID string, accnums []int64, sequences []int64, fee StdFee, msg sdk.Msg) []byte {
	bz, err := json.Marshal(StdSignDoc{
		ChainID:        chainID,
		AccountNumbers: accnums,
		Sequences:      sequences,
		FeeBytes:       fee.Bytes(),
		MsgBytes:       msg.GetSignBytes(),
	})
	if err != nil {
		panic(err)
	}
	return bz
}

// StdSignMsg is a convenience structure for passing along
// a Msg with the other requirements for a StdSignDoc before
// it is signed. For use in the CLI.
type StdSignMsg struct {
	ChainID        string
	AccountNumbers []int64
	Sequences      []int64
	Fee            StdFee
	Msg            sdk.Msg
	// XXX: Alt
}

// get message bytes
func (msg StdSignMsg) Bytes() []byte {
	return StdSignBytes(msg.ChainID, msg.AccountNumbers, msg.Sequences, msg.Fee, msg.Msg)
}

// Standard Signature
type StdSignature struct {
	crypto.PubKey    `json:"pub_key"` // optional
	crypto.Signature `json:"signature"`
	AccountNumber    int64 `json:"account_number"`
	Sequence         int64 `json:"sequence"`
}
