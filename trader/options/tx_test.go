package options

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tendermint/basecoin-examples/trader"
	"github.com/tendermint/basecoin/types"
	cmn "github.com/tendermint/go-common"
)

func TestBasicFlow(t *testing.T) {
	assert := assert.New(t)

	store := types.NewMemKVStore()
	pstore := trader.PrefixStore(store, []byte("options/"))
	accts := Accountant{store}

	a, b, c := cmn.RandBytes(20), cmn.RandBytes(20), cmn.RandBytes(20)
	bond := types.Coins{
		{Amount: 1000, Denom: "ATOM"},
	}
	trade := types.Coins{
		{Amount: 5, Denom: "BTC"},
	}
	price := types.Coins{
		{Amount: 10, Denom: "ETH"},
	}
	low := types.Coins{
		{Amount: 8, Denom: "ETH"},
	}

	tx := CreateOptionTx{
		Expiration: 100,
		Trade:      trade,
	}
	ctx := types.CallContext{
		CallerAddress: a,
		Coins:         bond,
		CallerAccount: &types.Account{
			Sequence: 20,
		},
	}
	// rejected as already expired
	res := tx.Apply(pstore, accts, ctx, 200)
	assert.True(res.IsErr())
	// acccept proper
	res = tx.Apply(pstore, accts, ctx, 50)
	assert.True(res.IsOK())
	addr := res.Data
	assert.NotEmpty(addr)

	// let's see the bond is set properly
	data, err := LoadData(pstore, addr)
	assert.Nil(err)
	assert.Equal(addr, data.Address())
	assert.Equal(bond, data.Bond)
	assert.Equal(20, data.Serial)
	assert.Equal(a, data.Issuer)
	assert.Equal(a, data.Holder)

	// no one can buy it right now
	tx2 := BuyOptionTx{
		Addr: addr,
	}
	ctxb := types.CallContext{
		CallerAddress: b,
		Coins:         price,
	}
	// make sure money returned from failed purchase
	ab := accts.GetOrCreateAccount(b)
	assert.True(ab.Balance.IsZero())
	res = tx2.Apply(pstore, accts, ctxb, 50)
	assert.True(res.IsErr())
	ab = accts.GetOrCreateAccount(b)
	assert.False(ab.Balance.IsZero())
	assert.Equal(price, ab.Balance)

	// let us place it for sale....
	tx3 := SellOptionTx{
		Addr:      addr,
		Price:     price,
		NewHolder: b,
	}
	ctx.Coins = low

	// make sure someone else cannot sell
	res = tx3.Apply(pstore, accts, ctxb, 50)
	assert.False(res.IsOK())

	// make sure the sell offer succeeds and money is refunded
	foo := accts.GetOrCreateAccount(a).Balance
	res = tx3.Apply(pstore, accts, ctx, 50)
	assert.True(res.IsOK())
	aa := accts.GetOrCreateAccount(a)
	assert.False(aa.Balance.IsZero())
	// make sure increased by low during the call
	assert.Equal(low, aa.Balance.Minus(foo))

	// now, we can buy it, but only b
	ctxc := types.CallContext{
		CallerAddress: c,
		Coins:         price,
	}
	// c is not authorized
	res = tx2.Apply(pstore, accts, ctxc, 50)
	assert.False(res.IsOK())
	// but b can go shopping!
	res = tx2.Apply(pstore, accts, ctxb, 50)
	assert.True(res.IsOK())

	// finally, our happy b (not c) can make the final trade
	tx4 := ExerciseOptionTx{
		Addr: addr,
	}
	// c is not authorized
	res = tx4.Apply(pstore, accts, ctxc, 50)
	assert.False(res.IsOK(), res.Log)
	ctxbl := types.CallContext{
		CallerAddress: b,
		Coins:         low,
	}
	// b doesn't pay enought is not authorized
	res = tx4.Apply(pstore, accts, ctxbl, 50)
	assert.False(res.IsOK())
	// now we pay enough
	ctxb = types.CallContext{
		CallerAddress: b,
		Coins:         trade,
	}
	res = tx4.Apply(pstore, accts, ctxb, 50)
	assert.True(res.IsOK(), res.Log)

	// now, let's make sure the option is gone
	data, err = LoadData(pstore, addr)
	assert.NotNil(err)

	// and the money is in everyone's account
	aa = accts.GetAccount(a)
	assert.False(aa.Balance.IsZero())
	assert.True(aa.Balance.IsGTE(trade)) // a got the trade
	ab = accts.GetAccount(b)
	assert.False(ab.Balance.IsZero())
	assert.True(ab.Balance.IsGTE(bond)) // b got the bond
}
