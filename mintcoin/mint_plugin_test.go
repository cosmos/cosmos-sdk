package mintcoin

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tendermint/basecoin/state"
	"github.com/tendermint/basecoin/types"
	wire "github.com/tendermint/go-wire"
)

func TestSaveLoad(t *testing.T) {
	assert := assert.New(t)
	store := types.NewMemKVStore()
	plugin := New("cash")
	addr1, addr2 := []byte("bigmoney"), []byte("litlefish")

	s := plugin.loadState(store)
	assert.NotNil(s)
	assert.False(s.IsBanker(addr1))
	s.AddBanker(addr1)
	plugin.saveState(store, s)

	s2 := plugin.loadState(store)
	assert.NotNil(s2)
	assert.True(s2.IsBanker(addr1))
	assert.False(s2.IsBanker(addr2))
}

func TestSetOptions(t *testing.T) {
	assert := assert.New(t)
	store := types.NewMemKVStore()
	plugin := New("cash")

	addr1, addr2 := []byte("bigmoney"), []byte("litlefish")
	hex1 := hex.EncodeToString(addr1)
	hex2 := hex.EncodeToString(addr2)
	assert.Equal("cash", plugin.Name())

	plugin.SetOption(store, AddBanker, hex1)
	st := plugin.loadState(store)
	assert.True(st.IsBanker(addr1))
	assert.False(st.IsBanker(addr2))

	plugin.SetOption(store, RemoveBanker, hex2)
	st = plugin.loadState(store)
	assert.True(st.IsBanker(addr1))
	assert.False(st.IsBanker(addr2))

	plugin.SetOption(store, AddBanker, hex2)
	plugin.SetOption(store, RemoveBanker, hex1)
	st = plugin.loadState(store)
	assert.False(st.IsBanker(addr1))
	assert.True(st.IsBanker(addr2))
}

func TestTransactions(t *testing.T) {
	assert := assert.New(t)
	store := types.NewMemKVStore()
	plugin := New("cash")

	addr1, addr2 := []byte("bigmoney"), []byte("litlefish")
	assert.Nil(state.GetAccount(store, addr1))

	tx := MintTx{
		Winners: []Winner{
			{
				Addr: addr1,
				Amount: types.Coins{
					{Denom: "BTC", Amount: 5},
					{Denom: "EUR", Amount: 100},
				},
			},
			{
				Addr: addr2,
				Amount: types.Coins{
					{Denom: "USD", Amount: 75},
				},
			},
		},
	}
	txBytes := wire.BinaryBytes(tx)
	ctx := types.CallContext{CallerAddress: addr1}
	res := plugin.RunTx(store, ctx, txBytes)

	// this won't work, cuz bigmoney isn't a banker yet
	assert.True(res.IsErr())
	assert.Nil(state.GetAccount(store, addr1))

	// let's set the options and watch the cash flow!
	hex1 := hex.EncodeToString(addr1)
	plugin.SetOption(store, AddBanker, hex1)
	res = plugin.RunTx(store, ctx, txBytes)
	assert.True(res.IsOK())
	acct1 := state.GetAccount(store, addr1)
	assert.NotNil(acct1)
	assert.True(acct1.Balance.IsPositive())
	assert.Equal(2, len(acct1.Balance))
	btc := acct1.Balance[0]
	assert.Equal("BTC", btc.Denom)
	assert.Equal(int64(5), btc.Amount)

	acct2 := state.GetAccount(store, addr2)
	assert.NotNil(acct2)
	assert.True(acct2.Balance.IsPositive())
	assert.Equal(1, len(acct2.Balance))
	usd := acct2.Balance[0]
	assert.Equal("USD", usd.Denom)
	assert.Equal(int64(75), usd.Amount)
}
