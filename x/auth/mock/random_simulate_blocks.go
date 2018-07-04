package mock

import (
	"fmt"
	"math/big"
	"math/rand"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
)

// RandomizedTesting tests application by sending random messages.
func (app *App) RandomizedTesting(t *testing.T, ops []TestAndRunTx, setups []RandSetup,
	invariants []AssertInvariants, numKeys int, numBlocks int, blockSize int) {

	time := time.Now().UnixNano()
	app.RandomizedTestingFromSeed(t, time, ops, setups, invariants, numKeys, numBlocks, blockSize)
}

// RandomizedTestingFromSeed tests an application by running the provided operations, testing the provided
// invariants, but using the provided seed.
func (app *App) RandomizedTestingFromSeed(t *testing.T, seed int64, ops []TestAndRunTx, setups []RandSetup,
	invariants []AssertInvariants, numKeys int, numBlocks int, blockSize int) {
	log := fmt.Sprintf("Starting SingleModuleTest with randomness created with seed %d", int(seed))
	keys, addrs := GenerateNPrivKeyAddressPairs(numKeys)
	r := rand.New(rand.NewSource(seed))

	for i := 0; i < len(setups); i++ {
		setups[i](r, keys)
	}
	RandomSetGenesis(r, app, addrs, []string{"foocoin"})
	header := abci.Header{Height: 0}
	for i := 0; i < numBlocks; i++ {
		app.BeginBlock(abci.RequestBeginBlock{})
		// Make sure invariants hold at begnning of block / when nothing was done.
		app.assertAllInvariants(t, invariants, log)
		ctx := app.NewContext(false, header)
		// TODO: Add modes to simulate "no load", "medium load", and "high load" blocks
		for j := 0; j < blockSize; j++ {
			logUpdate, err := ops[r.Intn(len(ops))](t, r, app, ctx, keys, log)
			// Add this to log
			log += "\n" + logUpdate
			require.Nil(t, err, log)
			app.assertAllInvariants(t, invariants, log)
		}
		app.EndBlock(abci.RequestEndBlock{})
		header.Height++
	}
}

func (app *App) assertAllInvariants(t *testing.T, tests []AssertInvariants, log string) {
	for i := 0; i < len(tests); i++ {
		tests[i](t, app, log)
	}
}

// nolint:
type Interval struct {
	lo int
	hi int
}

// Chooses an interval uniformly from the list of random
// intervals, and then chooses an element from an interval
// uniformly at random.
func RandFromInterval(r *rand.Rand, intervals []Interval) int {
	if len(intervals) == 0 {
		return 0
	}
	interval := intervals[r.Intn(len(intervals))]
	lo := interval.lo
	hi := interval.hi
	return r.Intn(hi-lo) + lo
}

// nolint:
type BigInterval struct {
	lo sdk.Int
	hi sdk.Int
}

// Chooses an interval uniformly from the list of random
// intervals, and then chooses an element from an interval
// uniformly at random.
func RandFromBigInterval(r *rand.Rand, intervals []BigInterval) sdk.Int {
	if len(intervals) == 0 {
		return sdk.ZeroInt()
	}
	interval := intervals[r.Intn(len(intervals))]
	lo := interval.lo
	hi := interval.hi
	diff := hi.Sub(lo)
	result := sdk.NewIntFromBigInt(new(big.Int).Rand(r, diff.BigInt()))
	result = result.Add(lo)
	return result
}
