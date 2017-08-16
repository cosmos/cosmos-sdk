package types

import (
	"fmt"

	"github.com/tendermint/basecoin/types"
	wire "github.com/tendermint/go-wire"
	"golang.org/x/crypto/ripemd160"
)

func init() {
	// register tx implementations with gowire
	wire.RegisterInterface(
		escrowwrap{},
		wire.ConcreteType{O: CreateEscrowTx{}, Byte: 0x01},
		wire.ConcreteType{O: ResolveEscrowTx{}, Byte: 0x02},
		wire.ConcreteType{O: ExpireEscrowTx{}, Byte: 0x03},
	)
}

type EscrowTx interface{}

type escrowwrap struct {
	EscrowTx
}

func ParseEscrowTx(data []byte) (EscrowTx, error) {
	holder := escrowwrap{}
	err := wire.ReadBinaryBytes(data, &holder)
	return holder.EscrowTx, err
}

func EscrowTxBytes(tx EscrowTx) []byte {
	return wire.BinaryBytes(escrowwrap{tx})
}

// CreateEscrowTx is used to create an escrow in the first place
type CreateEscrowTx struct {
	Recipient  []byte
	Arbiter    []byte
	Expiration uint64 // height when the offer expires
	// Sender and Amount come from the basecoin context
}

// ResolveEscrowTx must be signed by the Arbiter and resolves the escrow
// by sending the money to Sender or Recipient as specified
type ResolveEscrowTx struct {
	Escrow []byte
	Payout bool // if true, to Recipient, else back to Sender
}

// ExpireEscrowTx can be signed by anyone, and only succeeds if the
// Expiration height has passed.  All coins go back to the Sender
// (Intended to be used by the sender to recover old payments)
type ExpireEscrowTx struct {
	Escrow []byte
}

// EscrowData is our principal data structure in the db
type EscrowData struct {
	Sender     []byte
	Recipient  []byte
	Arbiter    []byte
	Expiration uint64 // height when the offer expires (0 = never)
	Amount     types.Coins
}

func (d EscrowData) IsExpired(h uint64) bool {
	return (d.Expiration != 0 && h > d.Expiration)
}

// Address is the ripemd160 hash of the escrow contents, which is constant
func (d EscrowData) Address() []byte {
	hasher := ripemd160.New()
	hasher.Write(d.Bytes())
	return hasher.Sum(nil)
}

func (d EscrowData) Bytes() []byte {
	return wire.BinaryBytes(d)
}

func ParseEscrow(data []byte) (EscrowData, error) {
	d := EscrowData{}
	err := wire.ReadBinaryBytes(data, &d)
	return d, err
}

func LoadEscrow(store types.KVStore, addr []byte) (EscrowData, error) {
	data := store.Get(addr)
	if len(data) == 0 {
		return EscrowData{}, fmt.Errorf("No escrow at: %X", addr)
	}
	return ParseEscrow(data)
}
