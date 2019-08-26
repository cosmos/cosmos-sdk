package simulation

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"testing"

	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// assertAll asserts the all invariants against application state
func assertAllInvariants(t *testing.T, app *baseapp.BaseApp, invs sdk.Invariants,
	event string, logWriter LogWriter, allInvariants bool) {

	ctx := app.NewContext(false, abci.Header{Height: app.LastBlockHeight() + 1})

	var invariantResults []string
	for i := 0; i < len(invs); i++ {
		res, stop := invs[i](ctx)
		if stop {
			invariantResults = append(invariantResults, res)
		}
	}

	if len(invariantResults) > 0 {
		fmt.Printf("Invariants broken after %s\n\n", event)
		for _, res := range invariantResults {
			fmt.Printf("%s\n", res)
		}
		logWriter.PrintLogs()
		t.Fatal()
	}
}

func getTestingMode(tb testing.TB) (testingMode bool, t *testing.T, b *testing.B) {
	testingMode = false
	if _t, ok := tb.(*testing.T); ok {
		t = _t
		testingMode = true
	} else {
		b = tb.(*testing.B)
	}
	return testingMode, t, b
}

// getBlockSize returns a block size as determined from the transition matrix.
// It targets making average block size the provided parameter. The three
// states it moves between are:
//  - "over stuffed" blocks with average size of 2 * avgblocksize,
//  - normal sized blocks, hitting avgBlocksize on average,
//  - and empty blocks, with no txs / only txs scheduled from the past.
func getBlockSize(r *rand.Rand, params Params, lastBlockSizeState, avgBlockSize int) (state, blockSize int) {
	// TODO: Make default blocksize transition matrix actually make the average
	// blocksize equal to avgBlockSize.
	state = params.BlockSizeTransitionMatrix.NextState(r, lastBlockSizeState)

	switch state {
	case 0:
		blockSize = r.Intn(avgBlockSize * 4)

	case 1:
		blockSize = r.Intn(avgBlockSize * 2)

	default:
		blockSize = 0
	}

	return state, blockSize
}

// PeriodicInvariants  returns an array of wrapped Invariants. Where each
// invariant function is only executed periodically defined by period and offset.
func PeriodicInvariants(invariants []sdk.Invariant, period, offset int) []sdk.Invariant {
	var outInvariants []sdk.Invariant
	for _, invariant := range invariants {
		outInvariant := func(ctx sdk.Context) (string, bool) {
			if int(ctx.BlockHeight())%period == offset {
				return invariant(ctx)
			}
			return "", false
		}
		outInvariants = append(outInvariants, outInvariant)
	}
	return outInvariants
}

func mustMarshalJSONIndent(o interface{}) []byte {
	bz, err := json.MarshalIndent(o, "", "  ")
	if err != nil {
		panic(fmt.Sprintf("failed to JSON encode: %s", err))
	}

	return bz
}
