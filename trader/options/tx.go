package options

import (
	abci "github.com/tendermint/abci/types"

	"github.com/tendermint/basecoin/types"
)

// CreateOptionTx is used to create an option in the first place
type CreateOptionTx struct {
	Expiration uint64      // height when the offer expires
	Trade      types.Coins // this is the money that can exercise the option
}

func (tx CreateOptionTx) Apply(store types.KVStore,
	accts types.AccountGetterSetter,
	ctx types.CallContext,
	height uint64) abci.Result {
	return abci.ErrUnknownRequest
}

// SellOptionTx is used to offer the option for sale
type SellOptionTx struct {
	Addr      []byte      // address of the refered option
	Price     types.Coins // required payment to transfer ownership
	NewHolder []byte      // set to allow for only one buyer, empty for any buyer
}

func (tx SellOptionTx) Apply(store types.KVStore,
	accts types.AccountGetterSetter,
	ctx types.CallContext,
	height uint64) abci.Result {
	return abci.ErrUnknownRequest
}

// BuyOptionTx is used to purchase the right to exercise the option
type BuyOptionTx struct {
	Addr []byte // address of the refered option
}

func (tx BuyOptionTx) Apply(store types.KVStore,
	accts types.AccountGetterSetter,
	ctx types.CallContext,
	height uint64) abci.Result {
	return abci.ErrUnknownRequest
}

// ExerciseOptionTx must send Trade and recieve Bond
type ExerciseOptionTx struct {
	Addr []byte // address of the refered option
}

func (tx ExerciseOptionTx) Apply(store types.KVStore,
	accts types.AccountGetterSetter,
	ctx types.CallContext,
	height uint64) abci.Result {
	return abci.ErrUnknownRequest
}

// DisolveOptionTx returns Bond to issue if expired or unpurchased
type DisolveOptionTx struct {
	Addr []byte // address of the refered option
}

func (tx DisolveOptionTx) Apply(store types.KVStore,
	accts types.AccountGetterSetter,
	ctx types.CallContext,
	height uint64) abci.Result {
	return abci.ErrUnknownRequest
}
