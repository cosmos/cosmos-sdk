package escrow

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tendermint/basecoin-examples/trader"
	"github.com/tendermint/basecoin-examples/trader/types"
	bc "github.com/tendermint/basecoin/types"
	cmn "github.com/tendermint/go-common"
)

func TestTransactions(t *testing.T) {
	assert := assert.New(t)
	store := bc.NewMemKVStore()
	sender, recv, arb := cmn.RandBytes(20), cmn.RandBytes(20), cmn.RandBytes(20)

	money := bc.Coins{
		{
			Amount: 1000,
			Denom:  "ATOM",
		},
		{
			Amount: 65,
			Denom:  "BTC",
		},
	}
	fees := bc.Coins{{
		Amount: 3,
		Denom:  "ATOM",
	}}

	tx := types.CreateEscrowTx{
		Recipient:  recv,
		Arbiter:    arb,
		Expiration: 100,
	}
	ctx := bc.CallContext{
		CallerAddress: sender,
		Coins:         money,
	}
	plugin := Plugin{
		height: 123,
		name:   "escrow",
	}
	pstore := plugin.prefix(store)
	accts := trader.NewAccountant(store)

	// error if already expired
	as := accts.GetOrCreateAccount(sender).Balance
	res := plugin.Exec(store, ctx, tx)
	assert.True(res.IsErr())
	assert.Equal(as.Plus(ctx.Coins), accts.GetAccount(sender).Balance)

	// we create the tx
	tx.Expiration = 500
	as = accts.GetOrCreateAccount(sender).Balance
	res = plugin.Exec(store, ctx, tx)
	assert.False(res.IsErr())
	assert.Equal(as, accts.GetAccount(sender).Balance)
	addr := res.Data
	assert.NotEmpty(addr)

	// load the escrow data and make sure it is happy
	esc, err := types.LoadEscrow(pstore, addr)
	if assert.Nil(err) {
		assert.Equal(addr, esc.Address())
		assert.Equal(sender, esc.Sender)
		assert.Equal(recv, esc.Recipient)
		assert.Equal(arb, esc.Arbiter)
		assert.Equal(uint64(500), esc.Expiration)
		assert.Equal(money, esc.Amount)
	}

	// try to complete it as the wrong person
	ctx = bc.CallContext{
		CallerAddress: sender,
		Coins:         fees,
	}
	rtx := types.ResolveEscrowTx{
		Escrow: addr,
		Payout: false,
	}
	as = accts.GetAccount(sender).Balance
	res = plugin.Exec(store, ctx, rtx)
	assert.True(res.IsErr())
	assert.Equal(as.Plus(ctx.Coins), accts.GetAccount(sender).Balance)

	// and the wrong locations
	ab := accts.GetOrCreateAccount(arb).Balance
	ctx = bc.CallContext{
		CallerAddress: arb,
		Coins:         fees,
	}
	rtx = types.ResolveEscrowTx{
		Escrow: cmn.RandBytes(20),
	}
	res = plugin.Exec(store, ctx, rtx)
	assert.True(res.IsErr())
	assert.Equal(ab.Plus(ctx.Coins), accts.GetAccount(arb).Balance)

	// try to expire, fails
	etx := types.ExpireEscrowTx{
		Escrow: addr,
	}
	res = plugin.Exec(store, ctx, etx)
	assert.True(res.IsErr())
	// assert.Equal(ctx.CallerAddress, pay.Addr)

	// complete as arbiter - yes!
	ar := accts.GetOrCreateAccount(recv).Balance
	rtx = types.ResolveEscrowTx{
		Escrow: addr,
		Payout: true,
	}
	res = plugin.Exec(store, ctx, rtx)
	assert.False(res.IsErr())
	assert.Equal(ar.Plus(money), accts.GetAccount(recv).Balance)

	// complete 2nd time -> error
	res = plugin.Exec(store, ctx, rtx)
	assert.True(res.IsErr())

	// no data to be seen
	esc, err = types.LoadEscrow(store, addr)
	assert.NotNil(err)
}
