package options

import (
	"fmt"

	abci "github.com/tendermint/abci/types"

	"github.com/tendermint/basecoin/types"
)

// CreateOptionTx is used to create an option in the first place
type CreateOptionTx struct {
	Expiration uint64      // height when the offer expires
	Trade      types.Coins // this is the money that can exercise the option
}

func (tx CreateOptionTx) Apply(store types.KVStore,
	accts Accountant,
	ctx types.CallContext,
	height uint64) abci.Result {

	issue := OptionIssue{
		Issuer:     ctx.CallerAddress,
		Serial:     ctx.CallerAccount.Sequence,
		Expiration: tx.Expiration,
		Bond:       ctx.Coins,
		Trade:      tx.Trade,
	}
	data := OptionData{
		OptionIssue: issue,
		OptionHolder: OptionHolder{
			Holder: ctx.CallerAddress,
		},
	}
	if data.IsExpired(height) {
		accts.Refund(ctx)
		return abci.ErrEncodingError.AppendLog("Already expired")
	}
	StoreData(store, data)
	addr := data.Address()
	return abci.NewResultOK(addr, fmt.Sprintf("new option: %X", addr))
}

// SellOptionTx is used to offer the option for sale
type SellOptionTx struct {
	Addr      []byte      // address of the refered option
	Price     types.Coins // required payment to transfer ownership
	NewHolder []byte      // set to allow for only one buyer, empty for any buyer
}

func (tx SellOptionTx) Apply(store types.KVStore,
	accts Accountant,
	ctx types.CallContext,
	height uint64) abci.Result {

	// always return money sent, no need
	accts.Refund(ctx)

	data, err := LoadData(store, tx.Addr)
	if err != nil {
		return abci.ErrEncodingError.AppendLog(err.Error())
	}

	// make sure we can do this
	if !data.CanSell(ctx.CallerAddress) {
		return abci.ErrUnauthorized.AppendLog("Not option holder")
	}

	data.NewHolder = tx.NewHolder
	data.Price = tx.Price
	StoreData(store, data)
	return abci.OK
}

// BuyOptionTx is used to purchase the right to exercise the option
type BuyOptionTx struct {
	Addr []byte // address of the refered option
}

func (tx BuyOptionTx) Apply(store types.KVStore,
	accts Accountant,
	ctx types.CallContext,
	height uint64) abci.Result {

	data, err := LoadData(store, tx.Addr)
	if err != nil {
		accts.Refund(ctx)
		return abci.ErrEncodingError.AppendLog(err.Error())
	}

	// make sure we can do this
	if !data.CanBuy(ctx.CallerAddress) {
		accts.Refund(ctx)
		return abci.ErrUnauthorized.AppendLog("Can't buy this option")
	}

	// make sure there is enough money to buy
	remain := ctx.Coins.Minus(data.Price)
	if !remain.IsNonnegative() {
		accts.Refund(ctx)
		return abci.ErrInsufficientFunds.AppendLog("Must pay more for the option")
	}

	// send the money to the seller
	accts.Pay(data.Holder, data.Price)
	// transfer ownership
	data.Holder = ctx.CallerAddress
	data.NewHolder = nil
	data.Price = nil
	StoreData(store, data)
	// and refund any overpayment
	accts.Pay(ctx.CallerAddress, remain)

	return abci.OK
}

// ExerciseOptionTx must send Trade and recieve Bond
type ExerciseOptionTx struct {
	Addr []byte // address of the refered option
}

func (tx ExerciseOptionTx) Apply(store types.KVStore,
	accts Accountant,
	ctx types.CallContext,
	height uint64) abci.Result {

	data, err := LoadData(store, tx.Addr)
	if err != nil {
		accts.Refund(ctx)
		return abci.ErrEncodingError.AppendLog(err.Error())
	}

	// make sure we can do this
	if !data.CanExercise(ctx.CallerAddress, height) {
		accts.Refund(ctx)
		return abci.ErrUnauthorized.AppendLog("Can't exercise this option")
	}

	// make sure there is enough money to trade
	remain := ctx.Coins.Minus(data.Trade)
	if !remain.IsNonnegative() {
		accts.Refund(ctx)
		return abci.ErrInsufficientFunds.AppendLog("Option requires higher trade value")
	}

	// pay back caller over-payment and the bond value
	accts.Pay(ctx.CallerAddress, remain)
	accts.Pay(ctx.CallerAddress, data.Bond)
	// the trade value goes to the original issuer
	accts.Pay(data.Issuer, data.Trade)

	// and remove this option from history
	DeleteData(store, data)

	return abci.OK
}

// DisolveOptionTx returns Bond to issue if expired or unpurchased
type DisolveOptionTx struct {
	Addr []byte // address of the refered option
}

func (tx DisolveOptionTx) Apply(store types.KVStore,
	accts Accountant,
	ctx types.CallContext,
	height uint64) abci.Result {

	// no need for payments, always return
	accts.Refund(ctx)

	data, err := LoadData(store, tx.Addr)
	if err != nil {
		return abci.ErrEncodingError.AppendLog(err.Error())
	}

	// make sure we can do this
	if !data.CanDissolve(ctx.CallerAddress, height) {
		return abci.ErrUnauthorized.AppendLog("Can't exercise this option")
	}

	// return bond to the issue
	accts.Pay(data.Issuer, data.Bond)
	// and remove this option from history
	DeleteData(store, data)

	return abci.OK
}
