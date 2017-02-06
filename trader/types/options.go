package types

import (
	"bytes"
	"fmt"

	"github.com/tendermint/basecoin/types"
	wire "github.com/tendermint/go-wire"
	"golang.org/x/crypto/ripemd160"
)

func init() {
	// register tx implementations with gowire
	wire.RegisterInterface(
		optionswrap{},
		wire.ConcreteType{O: CreateOptionTx{}, Byte: 0x01},
		wire.ConcreteType{O: SellOptionTx{}, Byte: 0x02},
		wire.ConcreteType{O: BuyOptionTx{}, Byte: 0x03},
		wire.ConcreteType{O: ExerciseOptionTx{}, Byte: 0x04},
		wire.ConcreteType{O: DisolveOptionTx{}, Byte: 0x05},
	)
}

type OptionsTx interface{}

type optionswrap struct {
	OptionsTx
}

func ParseOptionsTx(data []byte) (OptionsTx, error) {
	holder := optionswrap{}
	err := wire.ReadBinaryBytes(data, &holder)
	return holder.OptionsTx, err
}

func OptionsTxBytes(tx OptionsTx) []byte {
	return wire.BinaryBytes(optionswrap{tx})
}

// CreateOptionTx is used to create an option in the first place
type CreateOptionTx struct {
	Expiration uint64      // height when the offer expires
	Trade      types.Coins // this is the money that can exercise the option
}

// SellOptionTx is used to offer the option for sale
type SellOptionTx struct {
	Addr      []byte      // address of the refered option
	Price     types.Coins // required payment to transfer ownership
	NewHolder []byte      // set to allow for only one buyer, empty for any buyer
}

// BuyOptionTx is used to purchase the right to exercise the option
type BuyOptionTx struct {
	Addr []byte // address of the refered option
}

// ExerciseOptionTx must send Trade and recieve Bond
type ExerciseOptionTx struct {
	Addr []byte // address of the refered option
}

// DisolveOptionTx returns Bond to issue if expired or unpurchased
type DisolveOptionTx struct {
	Addr []byte // address of the refered option
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
	Serial     int         // this sequence number is from the apptx that created it
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

func (i OptionIssue) IsExpired(h uint64) bool {
	return (i.Expiration != 0 && h > i.Expiration)
}

// Address is the ripemd160 hash of the constant part of the option
func (i OptionIssue) Address() []byte {
	hasher := ripemd160.New()
	hasher.Write(i.Bytes())
	return hasher.Sum(nil)
}

func (i OptionIssue) Bytes() []byte {
	return wire.BinaryBytes(i)
}

func (d OptionData) Bytes() []byte {
	return wire.BinaryBytes(d)
}

// To buy, this option must be for sale, and the buyer must be
// listed (or an open sale)
func (d OptionData) CanBuy(buyer []byte) bool {
	return !d.Price.IsZero() &&
		(len(d.NewHolder) == 0 || bytes.Equal(buyer, d.NewHolder))
}

func (d OptionData) CanSell(buyer []byte) bool {
	return bytes.Equal(buyer, d.Holder)
}

func (d OptionData) CanExercise(addr []byte, h uint64) bool {
	return bytes.Equal(addr, d.Holder) && !d.IsExpired(h)
}

// CanDissolve if it is expired, or the holder, issue, and caller are the same
func (d OptionData) CanDissolve(addr []byte, h uint64) bool {
	return d.IsExpired(h) ||
		(bytes.Equal(addr, d.Holder) && bytes.Equal(d.Holder, d.Issuer))
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

func StoreData(store types.KVStore, data OptionData) {
	addr := data.Address()
	store.Set(addr, data.Bytes())
}

func DeleteData(store types.KVStore, data OptionData) {
	addr := data.Address()
	store.Set(addr, nil)
}
