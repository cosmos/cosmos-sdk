package simulation

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/mock"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
)

// Simulate tests application by sending random messages.
func Simulate(
	t *testing.T, app *baseapp.BaseApp, appStateFn func(r *rand.Rand, accs []sdk.AccAddress) json.RawMessage, ops []TestAndRunTx, setups []RandSetup,
	invariants []Invariant, numKeys int, numBlocks int, blockSize int,
) {
	time := time.Now().UnixNano()
	SimulateFromSeed(t, app, appStateFn, time, ops, setups, invariants, numKeys, numBlocks, blockSize)
}

// SimulateFromSeed tests an application by running the provided
// operations, testing the provided invariants, but using the provided seed.
func SimulateFromSeed(
	t *testing.T, app *baseapp.BaseApp, appStateFn func(r *rand.Rand, accs []sdk.AccAddress) json.RawMessage, seed int64, ops []TestAndRunTx, setups []RandSetup,
	invariants []Invariant, numKeys int, numBlocks int, blockSize int,
) {
	log := fmt.Sprintf("Starting SimulateFromSeed with randomness created with seed %d", int(seed))
	keys, addrs := mock.GeneratePrivKeyAddressPairs(numKeys)
	r := rand.New(rand.NewSource(seed))

	// Setup event stats
	events := make(map[string]uint)
	event := func(what string) {
		events[what]++
	}

	app.InitChain(abci.RequestInitChain{AppStateBytes: appStateFn(r, addrs)})
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
			logUpdate, err := ops[r.Intn(len(ops))](t, r, app, ctx, keys, log, event)
			log += "\n" + logUpdate

			require.Nil(t, err, log)
			AssertAllInvariants(t, app, invariants, log)
		}

		app.EndBlock(abci.RequestEndBlock{})
		header.Height++
	}

	DisplayEvents(events)
}

// AssertAllInvariants asserts a list of provided invariants against application state
func AssertAllInvariants(t *testing.T, app *baseapp.BaseApp, tests []Invariant, log string) {
	for i := 0; i < len(tests); i++ {
		tests[i](t, app, log)
	}
}
