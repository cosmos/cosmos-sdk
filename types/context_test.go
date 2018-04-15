package types_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/abci/types"
)

func TestContextGetOpShouldNeverPanic(t *testing.T) {
	var ms types.MultiStore
	ctx := types.NewContext(ms, abci.Header{}, false, nil)
	indices := []int64{
		-10, 1, 0, 10, 20,
	}

	for _, index := range indices {
		_, _ = ctx.GetOp(index)
	}
}
