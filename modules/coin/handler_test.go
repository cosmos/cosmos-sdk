package coin

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tendermint/basecoin"
	"github.com/tendermint/basecoin/stack"
	"github.com/tendermint/basecoin/types"
)

func TestHandlerPermissions(t *testing.T) {
	assert := assert.New(t)
	// TODO: need to update this when we actually have token store
	h := NewHandler()

	// these are all valid, except for minusCoins
	addr1 := basecoin.Actor{App: "coin", Address: []byte{1, 2}}
	addr2 := basecoin.Actor{App: "role", Address: []byte{7, 8}}
	someCoins := types.Coins{{"atom", 123}}
	minusCoins := types.Coins{{"eth", -34}}

	cases := []struct {
		valid bool
		tx    SendTx
		perms []basecoin.Actor
	}{
		// auth works with different apps
		{true,
			SendTx{
				Inputs:  []TxInput{NewTxInput(addr1, someCoins, 2)},
				Outputs: []TxOutput{NewTxOutput(addr2, someCoins)}},
			[]basecoin.Actor{addr1}},
		{true,
			SendTx{
				Inputs:  []TxInput{NewTxInput(addr2, someCoins, 2)},
				Outputs: []TxOutput{NewTxOutput(addr1, someCoins)}},
			[]basecoin.Actor{addr1, addr2}},
		// wrong permissions fail
		{false,
			SendTx{
				Inputs:  []TxInput{NewTxInput(addr1, someCoins, 2)},
				Outputs: []TxOutput{NewTxOutput(addr2, someCoins)}},
			[]basecoin.Actor{}},
		{false,
			SendTx{
				Inputs:  []TxInput{NewTxInput(addr1, someCoins, 2)},
				Outputs: []TxOutput{NewTxOutput(addr2, someCoins)}},
			[]basecoin.Actor{addr2}},
		// invalid input fails
		{false,
			SendTx{
				Inputs:  []TxInput{NewTxInput(addr1, minusCoins, 2)},
				Outputs: []TxOutput{NewTxOutput(addr2, minusCoins)}},
			[]basecoin.Actor{addr2}},
	}

	for i, tc := range cases {
		ctx := stack.MockContext().WithPermissions(tc.perms...)
		_, err := h.CheckTx(ctx, nil, tc.tx.Wrap())
		if tc.valid {
			assert.Nil(err, "%d: %+v", i, err)
		} else {
			assert.NotNil(err, "%d", i)
		}

		_, err = h.DeliverTx(ctx, nil, tc.tx.Wrap())
		if tc.valid {
			assert.Nil(err, "%d: %+v", i, err)
		} else {
			assert.NotNil(err, "%d", i)
		}

	}
}
