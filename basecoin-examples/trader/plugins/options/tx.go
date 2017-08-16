package options

import (
	"fmt"

	abci "github.com/tendermint/abci/types"
	"github.com/tendermint/basecoin-examples/trader"
	"github.com/tendermint/basecoin-examples/trader/types"

	bc "github.com/tendermint/basecoin/types"
)

func (p Plugin) runCreateOption(store bc.KVStore,
	accts trader.Accountant,
	ctx bc.CallContext,
	tx types.CreateOptionTx) abci.Result {

	issue := types.OptionIssue{
		Issuer:     ctx.CallerAddress,
		Serial:     ctx.CallerAccount.Sequence,
		Expiration: tx.Expiration,
		Bond:       ctx.Coins,
		Trade:      tx.Trade,
	}
	data := types.OptionData{
		OptionIssue: issue,
		OptionHolder: types.OptionHolder{
			Holder: ctx.CallerAddress,
		},
	}
	if data.IsExpired(p.height) {
		accts.Refund(ctx)
		return abci.ErrEncodingError.AppendLog("Already expired")
	}
	types.StoreOptionData(store, data)
	addr := data.Address()
	return abci.NewResultOK(addr, fmt.Sprintf("new option: %X", addr))
}

func (p Plugin) runSellOption(store bc.KVStore,
	accts trader.Accountant,
	ctx bc.CallContext,
	tx types.SellOptionTx) abci.Result {

	// always return money sent, no need
	accts.Refund(ctx)

	data, err := types.LoadOptionData(store, tx.Addr)
	if err != nil {
		return abci.ErrEncodingError.AppendLog(err.Error())
	}

	// make sure we can do this
	if !data.CanSell(ctx.CallerAddress) {
		return abci.ErrUnauthorized.AppendLog("Not option holder")
	}

	data.NewHolder = tx.NewHolder
	data.Price = tx.Price
	types.StoreOptionData(store, data)
	return abci.OK
}

func (p Plugin) runBuyOption(store bc.KVStore,
	accts trader.Accountant,
	ctx bc.CallContext,
	tx types.BuyOptionTx) abci.Result {

	data, err := types.LoadOptionData(store, tx.Addr)
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
	types.StoreOptionData(store, data)
	// and refund any overpayment
	accts.Pay(ctx.CallerAddress, remain)

	return abci.OK
}

func (p Plugin) runExerciseOption(store bc.KVStore,
	accts trader.Accountant,
	ctx bc.CallContext,
	tx types.ExerciseOptionTx) abci.Result {

	data, err := types.LoadOptionData(store, tx.Addr)
	if err != nil {
		accts.Refund(ctx)
		return abci.ErrEncodingError.AppendLog(err.Error())
	}

	// make sure we can do this
	if !data.CanExercise(ctx.CallerAddress, p.height) {
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
	types.DeleteOptionData(store, data)

	return abci.OK
}

func (p Plugin) runDisolveOption(store bc.KVStore,
	accts trader.Accountant,
	ctx bc.CallContext,
	tx types.DisolveOptionTx) abci.Result {

	// no need for payments, always return
	accts.Refund(ctx)

	data, err := types.LoadOptionData(store, tx.Addr)
	if err != nil {
		return abci.ErrEncodingError.AppendLog(err.Error())
	}

	// make sure we can do this
	if !data.CanDissolve(ctx.CallerAddress, p.height) {
		return abci.ErrUnauthorized.AppendLog("Can't exercise this option")
	}

	// return bond to the issue
	accts.Pay(data.Issuer, data.Bond)
	// and remove this option from history
	types.DeleteOptionData(store, data)

	return abci.OK
}
