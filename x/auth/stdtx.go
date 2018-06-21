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
	Msgs       []sdk.Msg      `json:"msg"`
	Fee        StdFee         `json:"fee"`
	Signatures []StdSignature `json:"signatures"`
	Memo       string         `json:"memo"`
}

func NewStdTx(msgs []sdk.Msg, fee StdFee, sigs []StdSignature, memo string) StdTx {
	return StdTx{
		Msgs:       msgs,
		Fee:        fee,
		Signatures: sigs,
		Memo:       memo,
	}
}

//nolint
func (tx StdTx) GetMsgs() []sdk.Msg { return tx.Msgs }

// GetSigners returns the addresses that must sign the transaction.
// Addresses are returned in a determistic order.
// They are accumulated from the GetSigners method for each Msg
// in the order they appear in tx.GetMsgs().
// Duplicate addresses will be ommitted.
func (tx StdTx) GetSigners() []sdk.Address {
	seen := map[string]bool{}
	var signers []sdk.Address
	for _, msg := range tx.GetMsgs() {
		for _, addr := range msg.GetSigners() {
			if !seen[addr.String()] {
				signers = append(signers, addr)
				seen[addr.String()] = true
			}
		}
	}
	return signers
}

//nolint
func (tx StdTx) GetMemo() string { return tx.Memo }

// Signatures returns the signature of signers who signed the Msg.
// GetSignatures returns the signature of signers who signed the Msg.
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
	return tx.GetMsgs()[0].GetSigners()[0]
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
	ChainID       string `json:"chain_id"`
	AccountNumber int64  `json:"account_number"`
	Sequence      int64  `json:"sequence"`
	FeeBytes      []byte `json:"fee_bytes"`
	MsgsBytes     []byte `json:"msg_bytes"`
	Memo          string `json:"memo"`
}

// StdSignBytes returns the bytes to sign for a transaction.
// TODO: change the API to just take a chainID and StdTx ?
func StdSignBytes(chainID string, accnum int64, sequence int64, fee StdFee, msgs []sdk.Msg, memo string) []byte {
	var msgBytes []byte
	for _, msg := range msgs {
		msgBytes = append(msgBytes, msg.GetSignBytes()...)
	}

	bz, err := json.Marshal(StdSignDoc{
		ChainID:       chainID,
		AccountNumber: accnum,
		Sequence:      sequence,
		FeeBytes:      fee.Bytes(),
		MsgsBytes:     msgBytes,
		Memo:          memo,
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
	ChainID       string
	AccountNumber int64
	Sequence      int64
	Fee           StdFee
	Msgs          []sdk.Msg
	Memo          string
}

// get message bytes
func (msg StdSignMsg) Bytes() []byte {
	return StdSignBytes(msg.ChainID, msg.AccountNumber, msg.Sequence, msg.Fee, msg.Msgs, msg.Memo)
}

// Standard Signature
type StdSignature struct {
	crypto.PubKey    `json:"pub_key"` // optional
	crypto.Signature `json:"signature"`
	AccountNumber    int64 `json:"account_number"`
	Sequence         int64 `json:"sequence"`
}
