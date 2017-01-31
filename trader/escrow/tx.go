package escrow

import (
	"bytes"

	abci "github.com/tendermint/abci/types"

	"github.com/tendermint/basecoin/types"
)

// CreateEscrowTx is used to create an escrow in the first place
type CreateEscrowTx struct {
	Recipient  []byte
	Arbiter    []byte
	Expiration uint64 // height when the offer expires
	// Sender and Amount come from the basecoin context
}

func (tx CreateEscrowTx) Apply(store types.KVStore, ctx types.CallContext, height uint64) (abci.Result, Payback) {
	// TODO: require fees? limit the size of the escrow?

	data := EscrowData{
		Sender:     ctx.CallerAddress,
		Recipient:  tx.Recipient,
		Arbiter:    tx.Arbiter,
		Expiration: tx.Expiration,
		Amount:     ctx.Coins,
	}
	// make sure all settings are valid, if not abort and return money
	if data.IsExpired(height) {
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

// ResolveEscrowTx must be signed by the Arbiter and resolves the escrow
// by sending the money to Sender or Recipient as specified
type ResolveEscrowTx struct {
	Escrow []byte
	Payout bool // if true, to Recipient, else back to Sender
}

func (tx ResolveEscrowTx) Apply(store types.KVStore, ctx types.CallContext, height uint64) (abci.Result, Payback) {
	// first load the data
	data := store.Get(tx.Escrow)
	if len(data) == 0 { // nil and []byte{}
		return abci.ErrBaseUnknownAddress, paybackCtx(ctx)
	}

	esc, err := ParseData(data)
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

// ExpireEscrowTx can be signed by anyone, and only succeeds if the
// Expiration height has passed.  All coins go back to the Sender
// (Intended to be used by the sender to recover old payments)
type ExpireEscrowTx struct {
	Escrow []byte
}

func (tx ExpireEscrowTx) Apply(store types.KVStore, ctx types.CallContext, height uint64) (abci.Result, Payback) {
	// first load the data
	data := store.Get(tx.Escrow)
	if len(data) == 0 { // nil and []byte{}
		return abci.ErrBaseUnknownAddress, paybackCtx(ctx)
	}

	esc, err := ParseData(data)
	if err != nil {
		return abci.NewError(abci.CodeType_BaseEncodingError, "Cannot parse data at location"), paybackCtx(ctx)
	}

	// only resolve if expired
	if !esc.IsExpired(height) {
		return abci.ErrUnauthorized, paybackCtx(ctx)
	}

	// wipe out the escrow and return the payment to sender
	pay := Payback{Amount: esc.Amount, Addr: esc.Sender}
	store.Set(tx.Escrow, nil)
	return abci.OK.AppendLog("Escrow settled"), pay
}
