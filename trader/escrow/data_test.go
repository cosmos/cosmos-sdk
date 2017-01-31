package escrow

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tendermint/basecoin/types"
)

func TestData(t *testing.T) {
	assert := assert.New(t)
	data := EscrowData{
		Sender:    []byte("1234567890qwertyuiop"),
		Recipient: []byte("AS1234567890qwertyui"),
		Arbiter:   []byte("ASDF1234567890qwerty"),
		Amount: types.Coins{
			{
				Amount: 1000,
				Denom:  "ATOM",
			},
		},
	}

	// make sure expiration only has meaning if non-zero
	assert.False(data.IsExpired(100))
	data.Expiration = 200
	assert.False(data.IsExpired(100))
	assert.True(data.IsExpired(201))

	// make sure we get a valid address
	addr := data.Address()
	assert.NotEmpty(addr)
	assert.Equal(20, len(addr))

	// make sure serialization/deserialization works
	b := data.Bytes()
	assert.NotEmpty(b)
	d2, err := ParseData(b)
	assert.Nil(err)
	assert.Equal(data, d2)

	// and make sure they have the same address
	assert.Equal(addr, d2.Address())
}

func TestTxParse(t *testing.T) {
	assert := assert.New(t)
	ctx := CreateEscrowTx{
		Recipient:  []byte("AS1234567890qwertyui"),
		Arbiter:    []byte("ASDF1234567890qwerty"),
		Expiration: 12345,
	}
	rtx := ResolveEscrowTx{
		Escrow: []byte("1234567890qwertyuiop"),
		Payout: true,
	}
	etx := ExpireEscrowTx{
		Escrow: []byte("1234567890qwertyuiop"),
	}

	// make sure all of them serialize and deserialize fine
	txs := []Tx{ctx, rtx, etx}
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
