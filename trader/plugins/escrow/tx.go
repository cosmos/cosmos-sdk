package escrow

import (
	"bytes"

	abci "github.com/tendermint/abci/types"
	"github.com/tendermint/basecoin-examples/trader"
	"github.com/tendermint/basecoin-examples/trader/types"

	bc "github.com/tendermint/basecoin/types"
)

func (p Plugin) runCreateEscrow(store bc.KVStore,
	accts trader.Accountant,
	ctx bc.CallContext,
	tx types.CreateEscrowTx) abci.Result {
	// TODO: require fees? limit the size of the escrow?

	data := types.EscrowData{
		Sender:     ctx.CallerAddress,
		Recipient:  tx.Recipient,
		Arbiter:    tx.Arbiter,
		Expiration: tx.Expiration,
		Amount:     ctx.Coins,
	}
	// make sure all settings are valid, if not abort and return money
	if data.IsExpired(p.height) {
		accts.Refund(ctx)
		return abci.NewError(abci.CodeType_BaseInvalidInput, "Escrow already expired")
	}
	if len(data.Recipient) != 20 {
		accts.Refund(ctx)
		return abci.ErrBaseInvalidInput.AppendLog("Invalid recipient address")
	}
	if len(data.Arbiter) != 20 {
		accts.Refund(ctx)
		return abci.ErrBaseInvalidInput.AppendLog("Invalid arbiter address")
	}

	// create the escrow contract
	addr := data.Address()
	store.Set(addr, data.Bytes())
	return abci.NewResultOK(addr, "Created Escrow")
}

func (p Plugin) runResolveEscrow(store bc.KVStore,
	accts trader.Accountant,
	ctx bc.CallContext,
	tx types.ResolveEscrowTx) abci.Result {
	// first load the data
	data := store.Get(tx.Escrow)
	if len(data) == 0 { // nil and []byte{}
		accts.Refund(ctx)
		return abci.ErrBaseUnknownAddress
	}

	esc, err := types.ParseEscrow(data)
	if err != nil {
		accts.Refund(ctx)
		return abci.NewError(abci.CodeType_BaseEncodingError, "Cannot parse data at location")
	}

	// only the Arbiter can resolve
	if !bytes.Equal(ctx.CallerAddress, esc.Arbiter) {
		accts.Refund(ctx)
		return abci.ErrUnauthorized
	}

	// Okay, now let's resolve this transaction!
	if tx.Payout {
		accts.Pay(esc.Recipient, esc.Amount)
	} else {
		accts.Pay(esc.Sender, esc.Amount)
	}

	// wipe out the escrow and return the payment
	store.Set(tx.Escrow, nil)
	return abci.OK.AppendLog("Escrow settled")
}

func (p Plugin) runExpireEscrow(store bc.KVStore,
	accts trader.Accountant,
	ctx bc.CallContext,
	tx types.ExpireEscrowTx) abci.Result {
	// first load the data
	data := store.Get(tx.Escrow)
	if len(data) == 0 { // nil and []byte{}
		accts.Refund(ctx)
		return abci.ErrBaseUnknownAddress
	}

	esc, err := types.ParseEscrow(data)
	if err != nil {
		accts.Refund(ctx)
		return abci.NewError(abci.CodeType_BaseEncodingError, "Cannot parse data at location")
	}

	// only resolve if expired
	if !esc.IsExpired(p.height) {
		accts.Refund(ctx)
		return abci.ErrUnauthorized
	}

	// wipe out the escrow and return the payment to sender
	accts.Pay(esc.Sender, esc.Amount)
	store.Set(tx.Escrow, nil)
	return abci.OK.AppendLog("Escrow settled")
}
