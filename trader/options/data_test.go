package options

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tendermint/basecoin/types"
	cmn "github.com/tendermint/go-common"
)

func TestData(t *testing.T) {
	assert := assert.New(t)
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

	data := OptionData{
		OptionIssue: OptionIssue{
			Issuer:     a,
			Serial:     5,
			Expiration: uint64(20),
			Bond:       bond,
			Trade:      trade,
		},
		OptionHolder: OptionHolder{
			Holder: a,
		},
	}

	addr := data.Address()
	assert.NotEmpty(addr)
	assert.True(data.IsExpired(30))
	assert.False(data.IsExpired(10))

	// make sure the initial state is only issuer dissolve or sell
	assert.True(data.CanSell(a))
	assert.False(data.CanSell(b))
	assert.False(data.CanBuy(a))
	assert.False(data.CanBuy(b))
	assert.True(data.CanDissolve(a, 10))
	assert.False(data.CanDissolve(b, 10))
	assert.True(data.CanDissolve(b, 50)) // b can dissolve if it ends up expired

	// set a price and make sure anyone can buy
	data.Price = price
	assert.True(data.CanBuy(b))
	assert.True(data.CanBuy(c))
	// or a closed sale
	data.NewHolder = b
	assert.True(data.CanBuy(b))
	assert.False(data.CanBuy(c))

	// we complete the sale and make sure the address didn't change
	data.Price = nil
	data.NewHolder = nil
	data.Holder = b
	newAddr := data.Address()
	assert.Equal(addr, newAddr)

	// now make sure the new buyer can sell and exercise
	assert.False(data.CanBuy(c))
	assert.False(data.CanSell(a))
	assert.True(data.CanSell(b))
	assert.False(data.CanSell(c))
	assert.False(data.CanExercise(a, 10))
	assert.True(data.CanExercise(b, 10))
	assert.False(data.CanExercise(c, 10))
	// and neither the holder nor the issuer cannot dissolve
	assert.False(data.CanDissolve(a, 10))
	assert.False(data.CanDissolve(b, 10))
}

func TestTxParse(t *testing.T) {
	assert := assert.New(t)

	trade := types.Coins{
		{Amount: 5, Denom: "BTC"},
		{Amount: 1000, Denom: "ATOM"},
	}
	price := types.Coins{
		{Amount: 3, Denom: "ETH"},
	}

	txs := []Tx{
		CreateOptionTx{
			Expiration: 12345,
			Trade:      trade,
		},
		SellOptionTx{
			Addr:      cmn.RandBytes(20),
			Price:     price,
			NewHolder: []byte{}, // note: nil is serialized/parsed as empty
		},
		SellOptionTx{
			Addr:      cmn.RandBytes(20),
			Price:     price,
			NewHolder: cmn.RandBytes(20),
		},
		BuyOptionTx{
			Addr: cmn.RandBytes(20),
		},
		ExerciseOptionTx{
			Addr: cmn.RandBytes(20),
		},
		DisolveOptionTx{
			Addr: cmn.RandBytes(20),
		},
	}

	// make sure all of them serialize and deserialize fine
	for i, tx := range txs {
		idx := strconv.Itoa(i)
		b := TxBytes(tx)
		if assert.NotEmpty(b, idx) {
			p, err := ParseTx(b)
			assert.Nil(err, idx)
			assert.NotNil(p, idx)
			assert.Equal(tx, p, idx)
		}
	}
}
