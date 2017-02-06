package escrow

import (
	"bytes"

	abci "github.com/tendermint/abci/types"
	"github.com/tendermint/basecoin-examples/trader/types"

	bc "github.com/tendermint/basecoin/types"
)

func (p Plugin) runCreateEscrow(store bc.KVStore, ctx bc.CallContext, tx types.CreateEscrowTx) (abci.Result, Payback) {
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
		return abci.NewError(abci.CodeType_BaseInvalidInput, "Escrow already expired"), paybackCtx(ctx)
	}
	if len(data.Recipient) != 20 {
		return abci.ErrBaseInvalidInput.AppendLog("Invalid recipient address"), paybackCtx(ctx)
	}
	if len(data.Arbiter) != 20 {
		return abci.ErrBaseInvalidInput.AppendLog("Invalid arbiter address"), paybackCtx(ctx)
	}

	// create the escrow contract
	addr := data.Address()
	store.Set(addr, data.Bytes())
	return abci.NewResultOK(addr, "Created Escrow"), Payback{}
}

func (p Plugin) runResolveEscrow(store bc.KVStore, ctx bc.CallContext, tx types.ResolveEscrowTx) (abci.Result, Payback) {
	// first load the data
	data := store.Get(tx.Escrow)
	if len(data) == 0 { // nil and []byte{}
		return abci.ErrBaseUnknownAddress, paybackCtx(ctx)
	}

	esc, err := types.ParseEscrow(data)
	if err != nil {
		return abci.NewError(abci.CodeType_BaseEncodingError, "Cannot parse data at location"), paybackCtx(ctx)
	}

	// only the Arbiter can resolve
	if !bytes.Equal(ctx.CallerAddress, esc.Arbiter) {
		return abci.ErrUnauthorized, paybackCtx(ctx)
	}

	// Okay, now let's resolve this transaction!
	pay := Payback{Amount: esc.Amount}
	if tx.Payout {
		pay.Addr = esc.Recipient
	} else {
		pay.Addr = esc.Sender
	}

	// wipe out the escrow and return the payment
	store.Set(tx.Escrow, nil)
	return abci.OK.AppendLog("Escrow settled"), pay
}

func (p Plugin) runExpireEscrow(store bc.KVStore, ctx bc.CallContext, tx types.ExpireEscrowTx) (abci.Result, Payback) {
	// first load the data
	data := store.Get(tx.Escrow)
	if len(data) == 0 { // nil and []byte{}
		return abci.ErrBaseUnknownAddress, paybackCtx(ctx)
	}

	esc, err := types.ParseEscrow(data)
	if err != nil {
		return abci.NewError(abci.CodeType_BaseEncodingError, "Cannot parse data at location"), paybackCtx(ctx)
	}

	// only resolve if expired
	if !esc.IsExpired(p.height) {
		return abci.ErrUnauthorized, paybackCtx(ctx)
	}

	// wipe out the escrow and return the payment to sender
	pay := Payback{Amount: esc.Amount, Addr: esc.Sender}
	store.Set(tx.Escrow, nil)
	return abci.OK.AppendLog("Escrow settled"), pay
}
