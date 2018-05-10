package auth

import (
	"encoding/hex"
	"encoding/json"
	"errors"

	bam "github.com/cosmos/cosmos-sdk/baseapp"
	crypto "github.com/tendermint/go-crypto"
	cmn "github.com/tendermint/tmlibs/common"
)

// Address in go-crypto style
type Address = cmn.HexBytes

// create an Address from a string
func GetAddress(address string) (addr Address, err error) {
	if len(address) == 0 {
		return addr, errors.New("must use provide address")
	}
	bz, err := hex.DecodeString(address)
	if err != nil {
		return nil, err
	}
	return Address(bz), nil
}

// Account is a standard account using a sequence number for replay protection
// and a pubkey for authentication.
type Account interface {
	GetAddress() Address
	SetAddress(Address) error // errors if already set.

	GetPubKey() crypto.PubKey // can return nil.
	SetPubKey(crypto.PubKey) error

	GetSequence() int64
	SetSequence(int64) error

	GetCoins() bam.Coins
	SetCoins(bam.Coins) error
}

// AccountDecoder unmarshals account bytes
type AccountDecoder func(accountBytes []byte) (Account, error)

// Standard Signature
type StdSignature struct {
	crypto.PubKey    `json:"pub_key"` // optional
	crypto.Signature `json:"signature"`
	Sequence         int64 `json:"sequence"`
}

var _ bam.Tx = (*StdTx)(nil)

// StdTx is a standard way to wrap a Msg with Fee and Signatures.
// NOTE: the first signature is the FeePayer (Signatures must not be nil).
type StdTx struct {
	Msg bam.Msg `json:"msg"`
	Fee StdFee  `json:"fee"`

	// Signatures returns the signature of signers who signed the Msg.
	// CONTRACT: Length returned is same as length of
	// pubkeys returned from MsgKeySigners, and the order
	// matches.
	// CONTRACT: If the signature is missing (ie the Msg is
	// invalid), then the corresponding signature is
	// .Empty().
	Signatures []StdSignature `json:"signatures"`
}

func NewStdTx(msg bam.Msg, fee StdFee, sigs []StdSignature) StdTx {
	return StdTx{
		Msg:        msg,
		Fee:        fee,
		Signatures: sigs,
	}
}

//nolintx
func (tx StdTx) GetMsg() bam.Msg               { return tx.Msg }
func (tx StdTx) GetSignatures() []StdSignature { return tx.Signatures }

// FeePayer returns the address responsible for paying the fees
// for the transactions. It's the first address returned by msg.GetSigners().
// If GetSigners() is empty, this panics.
func FeePayer(tx StdTx) bam.Address {
	return tx.GetMsg().GetSigners()[0]
}

//__________________________________________________________

// StdFee includes the amount of coins paid in fees and the maximum
// gas to be used by the transaction. The ratio yields an effective "gasprice",
// which must be above some miminum to be accepted into the mempool.
type StdFee struct {
	Amount bam.Coins `json:"amount"`
	Gas    int64     `json:"gas"`
}

func NewStdFee(gas int64, amount ...bam.Coin) StdFee {
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
	if fee.Amount.Len() == 0 {
		fee.Amount = bam.Coins{}
	}
	bz, err := json.Marshal(fee) // TODO
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
	ChainID   string  `json:"chain_id"`
	Sequences []int64 `json:"sequences"`
	FeeBytes  []byte  `json:"fee_bytes"`
	MsgBytes  []byte  `json:"msg_bytes"`
	AltBytes  []byte  `json:"alt_bytes"`
}

// StdSignBytes returns the bytes to sign for a transaction.
// TODO: change the API to just take a chainID and StdTx ?
func StdSignBytes(chainID string, sequences []int64, fee StdFee, msg bam.Msg) []byte {
	bz, err := json.Marshal(StdSignDoc{
		ChainID:   chainID,
		Sequences: sequences,
		FeeBytes:  fee.Bytes(),
		MsgBytes:  msg.GetSignBytes(),
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
	ChainID   string
	Sequences []int64
	Fee       StdFee
	Msg       bam.Msg
	// XXX: Alt
}

// get message bytes
func (msg StdSignMsg) Bytes() []byte {
	return StdSignBytes(msg.ChainID, msg.Sequences, msg.Fee, msg.Msg)
}
