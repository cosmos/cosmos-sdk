package mock

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
)

// RandomizedTesting tests application by sending random messages.
func RandomizedTesting(
	t *testing.T, app *baseapp.BaseApp, ops []TestAndRunTx, setups []RandSetup,
	invariants []Invariant, numKeys int, numBlocks int, blockSize int,
) {
	time := time.Now().UnixNano()
	RandomizedTestingFromSeed(t, app, time, ops, setups, invariants, numKeys, numBlocks, blockSize)
}

// RandomizedTestingFromSeed tests an application by running the provided
// operations, testing the provided invariants, but using the provided seed.
func RandomizedTestingFromSeed(
	t *testing.T, app *baseapp.BaseApp, seed int64, ops []TestAndRunTx, setups []RandSetup,
	invariants []Invariant, numKeys int, numBlocks int, blockSize int,
) {
	log := fmt.Sprintf("Starting SingleModuleTest with randomness created with seed %d", int(seed))
	keys, _ := GeneratePrivKeyAddressPairs(numKeys)
	r := rand.New(rand.NewSource(seed))

	// XXX TODO
	// RandomSetGenesis(r, app, addrs, []string{"foocoin"})
	app.InitChain(abci.RequestInitChain{})
	for i := 0; i < len(setups); i++ {
		setups[i](r, keys)
	}
	app.Commit()

	header := abci.Header{Height: 0}

	for i := 0; i < numBlocks; i++ {
		app.BeginBlock(abci.RequestBeginBlock{})

		// Make sure invariants hold at beginning of block and when nothing was
		// done.
		AssertAllInvariants(t, app, invariants, log)

		ctx := app.NewContext(false, header)

		// TODO: Add modes to simulate "no load", "medium load", and
		// "high load" blocks.
		for j := 0; j < blockSize; j++ {
			logUpdate, err := ops[r.Intn(len(ops))](t, r, app, ctx, keys, log)
			log += "\n" + logUpdate

			require.Nil(t, err, log)
			AssertAllInvariants(t, app, invariants, log)
		}

		app.EndBlock(abci.RequestEndBlock{})
		header.Height++
	}
}

func AssertAllInvariants(t *testing.T, app *baseapp.BaseApp, tests []Invariant, log string) {
	for i := 0; i < len(tests); i++ {
		tests[i](t, app, log)
	}
}
