package util

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"

	sdk "github.com/cosmos/cosmos-sdk"
	"github.com/cosmos/cosmos-sdk/state"
)

func NewChainTx(name string, height uint64, data []byte) ChainedTx {
	return chainTx{
		ChainData: ChainData{name, height},
		Data:      data,
	}
}

type chainTx struct {
	ChainData
	Data []byte
}

func (c chainTx) GetChain() ChainData {
	return c.ChainData
}

func (c chainTx) GetTx() interface{} {
	return RawTx{c.Data}
}

func TestChain(t *testing.T) {
	assert := assert.New(t)
	msg := "got it"
	chainID := "my-chain"
	height := uint64(100)

	raw := []byte{1, 2, 3, 4}
	cases := []struct {
		tx       interface{}
		valid    bool
		errorMsg string
	}{
		// check the chain ids are validated
		{NewChainTx(chainID, 0, raw), true, ""},
		// non-matching chainid, or impossible chain id
		{NewChainTx("someone-else", 0, raw), false, "someone-else: Wrong chain"},
		{NewChainTx("Inval$$d:CH%%n", 0, raw), false, "Wrong chain"},
		// Wrong tx type
		{raw, false, "No chain id provided"},
		// Check different heights - must be 0 or higher than current height
		{NewChainTx(chainID, height+1, raw), true, ""},
		{NewChainTx(chainID, height, raw), false, "Tx expired"},
		{NewChainTx(chainID, 1, raw), false, "expired"},
		{NewChainTx(chainID, 0, raw), true, ""},
	}

	// generic args here...
	ctx := MockContext(chainID, height)
	store := state.NewMemKVStore()

	// build the stack
	ok := OKHandler{Log: msg}
	app := sdk.ChainDecorators(Chain{}).WithHandler(ok)

	for idx, tc := range cases {
		i := strconv.Itoa(idx)

		// make sure check returns error, not a panic crash
		cres, err := app.CheckTx(ctx, store, tc.tx)
		if tc.valid {
			assert.Nil(err, "%d: %+v", idx, err)
			assert.Equal(msg, cres.Log, i)
		} else {
			if assert.NotNil(err, i) {
				assert.Contains(err.Error(), tc.errorMsg, i)
			}
		}

		// make sure deliver returns error, not a panic crash
		dres, err := app.DeliverTx(ctx, store, tc.tx)
		if tc.valid {
			assert.Nil(err, "%d: %+v", idx, err)
			assert.Equal(msg, dres.Log, i)
		} else {
			if assert.NotNil(err, i) {
				assert.Contains(err.Error(), tc.errorMsg, i)
			}
		}
	}
}
