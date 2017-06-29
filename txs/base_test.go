package txs

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tendermint/go-wire/data"

	"github.com/tendermint/basecoin"
	"github.com/tendermint/basecoin/types"
)

func TestEncoding(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	raw := NewRaw([]byte{0x34, 0xa7}).Wrap()
	raw2 := NewRaw([]byte{0x73, 0x86, 0x22}).Wrap()
	coin := types.Coin{Denom: "atom", Amount: 123}
	addr := basecoin.Actor{
		App:     "cash",
		Address: []byte{0x12, 0x34, 0x56, 0x78, 0x90, 0xab, 0xcd, 0xef},
	}

	cases := []struct {
		Tx basecoin.Tx
	}{
		{raw},
		{NewFee(raw, coin, addr).Wrap()},
		{NewMultiTx(raw, raw2).Wrap()},
	}

	for idx, tc := range cases {
		i := strconv.Itoa(idx)
		tx := tc.Tx

		// test json in and out
		js, err := data.ToJSON(tx)
		require.Nil(err, i)
		var jtx basecoin.Tx
		err = data.FromJSON(js, &jtx)
		require.Nil(err, i)
		assert.Equal(tx, jtx, i)

		// test wire in and out
		bin, err := data.ToWire(tx)
		require.Nil(err, i)
		var wtx basecoin.Tx
		err = data.FromWire(bin, &wtx)
		require.Nil(err, i)
		assert.Equal(tx, wtx, i)
	}
}
