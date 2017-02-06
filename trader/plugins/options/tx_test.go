package options

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tendermint/basecoin-examples/trader"
	"github.com/tendermint/basecoin-examples/trader/types"
	bc "github.com/tendermint/basecoin/types"
	cmn "github.com/tendermint/go-common"
)

func TestBasicFlow(t *testing.T) {
	assert := assert.New(t)

	store := bc.NewMemKVStore()
	plugin := Plugin{
		height: 200,
		name:   "options",
	}
	pstore := plugin.prefix(store)
	accts := trader.NewAccountant(store)

	a, b, c := cmn.RandBytes(20), cmn.RandBytes(20), cmn.RandBytes(20)
	bond := bc.Coins{{Amount: 1000, Denom: "ATOM"}}
	trade := bc.Coins{{Amount: 5, Denom: "BTC"}}
	price := bc.Coins{{Amount: 10, Denom: "ETH"}}
	low := bc.Coins{{Amount: 8, Denom: "ETH"}}

	tx := types.CreateOptionTx{
		Expiration: 100,
		Trade:      trade,
	}
	ctx := bc.CallContext{
		CallerAddress: a,
		Coins:         bond,
		CallerAccount: &bc.Account{
			Sequence: 20,
		},
	}
	// rejected as already expired
	res := plugin.Exec(store, ctx, tx)
	assert.True(res.IsErr())
	// acccept proper
	plugin.height = 50
	res = plugin.Exec(store, ctx, tx)
	assert.True(res.IsOK())
	addr := res.Data
	assert.NotEmpty(addr)

	// let's see the bond is set properly
	data, err := types.LoadData(pstore, addr)
	assert.Nil(err)
	assert.Equal(addr, data.Address())
	assert.Equal(bond, data.Bond)
	assert.Equal(20, data.Serial)
	assert.Equal(a, data.Issuer)
	assert.Equal(a, data.Holder)

	// no one can buy it right now
	tx2 := types.BuyOptionTx{
		Addr: addr,
	}
	ctxb := bc.CallContext{
		CallerAddress: b,
		Coins:         price,
	}
	// make sure money returned from failed purchase
	ab := accts.GetOrCreateAccount(b)
	assert.True(ab.Balance.IsZero())
	res = plugin.Exec(store, ctxb, tx2)
	assert.True(res.IsErr())
	ab = accts.GetOrCreateAccount(b)
	assert.False(ab.Balance.IsZero())
	assert.Equal(price, ab.Balance)

	// let us place it for sale....
	tx3 := types.SellOptionTx{
		Addr:      addr,
		Price:     price,
		NewHolder: b,
	}
	ctx.Coins = low

	// make sure someone else cannot sell
	res = plugin.Exec(store, ctxb, tx3)
	assert.False(res.IsOK())

	// make sure the sell offer succeeds and money is refunded
	foo := accts.GetOrCreateAccount(a).Balance
	res = plugin.Exec(store, ctx, tx3)
	assert.True(res.IsOK())
	aa := accts.GetOrCreateAccount(a)
	assert.False(aa.Balance.IsZero())
	// make sure increased by low during the call
	assert.Equal(low, aa.Balance.Minus(foo))

	// now, we can buy it, but only b
	ctxc := bc.CallContext{
		CallerAddress: c,
		Coins:         price,
	}
	// c is not authorized
	res = plugin.Exec(store, ctxc, tx2)
	assert.False(res.IsOK())
	// but b can go shopping!
	res = plugin.Exec(store, ctxb, tx2)
	assert.True(res.IsOK())

	// finally, our happy b (not c) can make the final trade
	tx4 := types.ExerciseOptionTx{
		Addr: addr,
	}
	// c is not authorized
	res = plugin.Exec(store, ctxc, tx4)
	assert.False(res.IsOK(), res.Log)
	ctxbl := bc.CallContext{
		CallerAddress: b,
		Coins:         low,
	}
	// b doesn't pay enought is not authorized
	res = plugin.Exec(store, ctxbl, tx4)
	assert.False(res.IsOK())
	// now we pay enough
	ctxb = bc.CallContext{
		CallerAddress: b,
		Coins:         trade,
	}
	res = plugin.Exec(store, ctxb, tx4)
	assert.True(res.IsOK(), res.Log)

	// now, let's make sure the option is gone
	data, err = types.LoadData(pstore, addr)
	assert.NotNil(err)

	// and the money is in everyone's account
	aa = accts.GetAccount(a)
	assert.False(aa.Balance.IsZero())
	assert.True(aa.Balance.IsGTE(trade)) // a got the trade
	ab = accts.GetAccount(b)
	assert.False(ab.Balance.IsZero())
	assert.True(ab.Balance.IsGTE(bond)) // b got the bond
}
