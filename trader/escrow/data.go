package escrow

import (
	"fmt"

	abci "github.com/tendermint/abci/types"
	"github.com/tendermint/basecoin/state"
	"github.com/tendermint/basecoin/types"
	wire "github.com/tendermint/go-wire"
	"golang.org/x/crypto/ripemd160"
)

func init() {
	// register tx implementations with gowire
	wire.RegisterInterface(
		txwrap{},
		wire.ConcreteType{O: CreateEscrowTx{}, Byte: 0x01},
		wire.ConcreteType{O: ResolveEscrowTx{}, Byte: 0x02},
		wire.ConcreteType{O: ExpireEscrowTx{}, Byte: 0x03},
	)
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

func ParseData(data []byte) (EscrowData, error) {
	d := EscrowData{}
	err := wire.ReadBinaryBytes(data, &d)
	return d, err
}

func LoadData(store types.KVStore, addr []byte) (EscrowData, error) {
	data := store.Get(addr)
	if len(data) == 0 {
		return EscrowData{}, fmt.Errorf("No escrow at: %X", addr)
	}
	return ParseData(data)
}

// Payback is used to signal who to send the money to
type Payback struct {
	Addr   []byte
	Amount types.Coins
}

func paybackCtx(ctx types.CallContext) Payback {
	return Payback{
		Addr:   ctx.CallerAddress,
		Amount: ctx.Coins,
	}
}

// Pay is used to return money back to one person after the transaction
// this could refund the fees, or pay out escrow, or anything else....
func (p Payback) Pay(store types.KVStore) {
	if len(p.Addr) == 20 {
		acct := state.GetAccount(store, p.Addr)
		if acct == nil {
			acct = &types.Account{}
		}
		acct.Balance = acct.Balance.Plus(p.Amount)
		state.SetAccount(store, p.Addr, acct)
	}
}

type Tx interface {
	Apply(store types.KVStore, ctx types.CallContext, height uint64) (abci.Result, Payback)
}

type txwrap struct {
	Tx
}

func ParseTx(data []byte) (Tx, error) {
	holder := txwrap{}
	err := wire.ReadBinaryBytes(data, &holder)
	return holder.Tx, err
}

func TxBytes(tx Tx) []byte {
	return wire.BinaryBytes(txwrap{tx})
}
