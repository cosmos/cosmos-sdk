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
	invariants []Invariant, numKeys int, numBlocks int, blockSize int, minTimePerBlock int64, maxTimePerBlock int64, signingFraction float64, evidenceFraction float64,
) {
	time := time.Now().UnixNano()
	SimulateFromSeed(t, app, appStateFn, time, ops, setups, invariants, numKeys, numBlocks, blockSize, minTimePerBlock, maxTimePerBlock, signingFraction, evidenceFraction)
}

// SimulateFromSeed tests an application by running the provided
// operations, testing the provided invariants, but using the provided seed.
func SimulateFromSeed(
	t *testing.T, app *baseapp.BaseApp, appStateFn func(r *rand.Rand, accs []sdk.AccAddress) json.RawMessage, seed int64, ops []TestAndRunTx, setups []RandSetup,
	invariants []Invariant, numKeys int, numBlocks int, blockSize int, minTimePerBlock int64, maxTimePerBlock int64, signingFraction float64, evidenceFraction float64,
) {
	log := fmt.Sprintf("Starting SimulateFromSeed with randomness created with seed %d", int(seed))
	keys, addrs := mock.GeneratePrivKeyAddressPairs(numKeys)
	r := rand.New(rand.NewSource(seed))

	// Setup event stats
	events := make(map[string]uint)
	event := func(what string) {
		events[what]++
	}

	time := int64(0)
	timeDiff := maxTimePerBlock - minTimePerBlock

	res := app.InitChain(abci.RequestInitChain{AppStateBytes: appStateFn(r, addrs)})
	validators := make(map[string]abci.Validator)
	for _, validator := range res.Validators {
		validators[string(validator.Address)] = validator
	}

	for i := 0; i < len(setups); i++ {
		setups[i](r, keys)
	}
	app.Commit()

	header := abci.Header{Height: 0, Time: time}

	for i := 0; i < numBlocks; i++ {

		// Generate a random RequestBeginBlock with the current validator set
		request := RandomRequestBeginBlock(t, r, validators, signingFraction, evidenceFraction, header.Height, header.Time)
		app.BeginBlock(request)

		// Make sure invariants hold at beginning of block
		AssertAllInvariants(t, app, invariants, log)

		ctx := app.NewContext(false, header)

		var thisBlockSize int
		load := r.Float64()
		switch {
		case load < 0.33:
			thisBlockSize = 0
		case load < 0.66:
			thisBlockSize = blockSize
		default:
			thisBlockSize = blockSize * 2
		}
		for j := 0; j < thisBlockSize; j++ {
			logUpdate, err := ops[r.Intn(len(ops))](t, r, app, ctx, keys, log, event)
			log += "\n" + logUpdate

			require.Nil(t, err, log)
			AssertAllInvariants(t, app, invariants, log)
		}

		res := app.EndBlock(abci.RequestEndBlock{})
		UpdateValidators(t, validators, res.ValidatorUpdates)
		header.Height++
		header.Time += minTimePerBlock + int64(r.Intn(int(timeDiff)))
	}

	fmt.Printf("Simulation complete. Final height (blocks): %d, final time (seconds): %d\n", header.Height, header.Time)
	DisplayEvents(events)
}

// RandomRequestBeginBlock generates a list of signing validators according to the provided list of validators, signing fraction, and evidence fraction
func RandomRequestBeginBlock(t *testing.T, r *rand.Rand, validators map[string]abci.Validator, signingFraction float64, evidenceFraction float64,
	currentHeight int64, currentTime int64) abci.RequestBeginBlock {
	require.True(t, len(validators) > 0, "Zero validators can't sign a block!")
	signingValidators := make([]abci.SigningValidator, len(validators))
	i := 0
	for _, val := range validators {
		signed := r.Float64() < signingFraction
		signingValidators[i] = abci.SigningValidator{
			Validator:       val,
			SignedLastBlock: signed,
		}
		i++
	}
	evidence := make([]abci.Evidence, 0)
	if r.Float64() < evidenceFraction {
		// TODO Also include past evidence
		validator := signingValidators[r.Intn(len(signingValidators))].Validator
		var currentTotalVotingPower int64
		for _, val := range validators {
			currentTotalVotingPower += val.Power
		}
		evidence = append(evidence, abci.Evidence{
			Type:             "DOUBLE_SIGN",
			Validator:        validator,
			Height:           currentHeight,
			Time:             currentTime,
			TotalVotingPower: currentTotalVotingPower,
		})
	}
	return abci.RequestBeginBlock{
		Validators:          signingValidators,
		ByzantineValidators: evidence,
	}
}

// AssertAllInvariants asserts a list of provided invariants against application state
func AssertAllInvariants(t *testing.T, app *baseapp.BaseApp, tests []Invariant, log string) {
	for i := 0; i < len(tests); i++ {
		tests[i](t, app, log)
	}
}

// UpdateValidators mimicks Tendermint's update logic
func UpdateValidators(t *testing.T, current map[string]abci.Validator, updates []abci.Validator) {
	for _, update := range updates {
		switch {
		case update.Power == 0:
			require.NotNil(t, current[string(update.Address)], "tried to delete a nonexistent validator")
			delete(current, string(update.Address))
		default:
			current[string(update.Address)] = update
		}
	}
}
