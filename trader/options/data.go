package options

import (
	"fmt"

	abci "github.com/tendermint/abci/types"
	"github.com/tendermint/basecoin/types"
	wire "github.com/tendermint/go-wire"
	"golang.org/x/crypto/ripemd160"
)

func init() {
	// register tx implementations with gowire
	wire.RegisterInterface(
		txwrap{},
		wire.ConcreteType{O: CreateOptionTx{}, Byte: 0x01},
		wire.ConcreteType{O: SellOptionTx{}, Byte: 0x02},
		wire.ConcreteType{O: BuyOptionTx{}, Byte: 0x03},
		wire.ConcreteType{O: ExerciseOptionTx{}, Byte: 0x04},
		wire.ConcreteType{O: DisolveOptionTx{}, Byte: 0x05},
	)
}

// OptionData is our principal data structure in the db
type OptionData struct {
	OptionIssue
	OptionHolder
}

// OptionIssue is the constant part, created wth the option, never changes
type OptionIssue struct {
	// this is for the normal option functionality
	Issuer     []byte
	Serial     int64       // this serial number is from the apptx that created it
	Expiration uint64      // height when the offer expires (0 = never)
	Bond       types.Coins // this is stored upon creation of the option
	Trade      types.Coins // this is the money that can exercise the option
}

// OptionHolder is the dynamic section of who can excercise the options
type OptionHolder struct {
	// this is for buying/selling the option (should be a separate struct?)
	Holder    []byte
	NewHolder []byte      // set to allow for only one buyer, empty for any buyer
	Price     types.Coins // required payment to transfer ownership
}

func (d OptionData) IsExpired(h uint64) bool {
	return (d.Expiration != 0 && h > d.Expiration)
}

// Address is the ripemd160 hash of the constant part of the option
func (d OptionData) Address() []byte {
	hasher := ripemd160.New()
	hasher.Write(d.OptionIssue.Bytes())
	return hasher.Sum(nil)
}

func (d OptionData) Bytes() []byte {
	return wire.BinaryBytes(d)
}

func (i OptionIssue) Bytes() []byte {
	return wire.BinaryBytes(i)
}

func ParseData(data []byte) (OptionData, error) {
	d := OptionData{}
	err := wire.ReadBinaryBytes(data, &d)
	return d, err
}

func LoadData(store types.KVStore, addr []byte) (OptionData, error) {
	data := store.Get(addr)
	if len(data) == 0 {
		return OptionData{}, fmt.Errorf("No option at: %X", addr)
	}
	return ParseData(data)
}

// // Payback is used to signal who to send the money to
// type Payback struct {
// 	Addr   []byte
// 	Amount types.Coins
// }

// func paybackCtx(ctx types.CallContext) Payback {
// 	return Payback{
// 		Addr:   ctx.CallerAddress,
// 		Amount: ctx.Coins,
// 	}
// }

// // Pay is used to return money back to one person after the transaction
// // this could refund the fees, or pay out escrow, or anything else....
// func (p Payback) Pay(store types.KVStore) {
// 	if len(p.Addr) == 20 {
// 		acct := state.GetAccount(store, p.Addr)
// 		if acct == nil {
// 			acct = &types.Account{}
// 		}
// 		acct.Balance = acct.Balance.Plus(p.Amount)
// 		state.SetAccount(store, p.Addr, acct)
// 	}
// }

type Tx interface {
	// store is the prefixed store for options
	// accts lets us access all accounts
	// ctx and height come from the calling block
	Apply(store types.KVStore, accts types.AccountGetterSetter,
		ctx types.CallContext, height uint64) abci.Result
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
