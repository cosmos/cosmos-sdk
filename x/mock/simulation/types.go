package simulation

import (
	"math/rand"
	"testing"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto"
)

type (
	// TestAndRunTx produces a fuzzed transaction, and ensures the state
	// transition was as expected. It returns a descriptive message "action"
	// about what this fuzzed tx actually did, for ease of debugging.
	TestAndRunTx func(
		t *testing.T, r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
		privKeys []crypto.PrivKey, log string, event func(string),
	) (action string, err sdk.Error)

	// RandSetup performs the random setup the mock module needs.
	RandSetup func(r *rand.Rand, privKeys []crypto.PrivKey)

	// An Invariant is a function which tests a particular invariant.
	// If the invariant has been broken, the function should halt the
	// test and output the log.
	Invariant func(t *testing.T, app *baseapp.BaseApp, log string)

	mockValidator struct {
		val           abci.Validator
		livenessState int
	}
)

// PeriodicInvariant returns an Invariant function closure that asserts
// a given invariant every 1 / period times it is called.
func PeriodicInvariant(invariant Invariant, period int) Invariant {
	counter := 0
	return func(t *testing.T, app *baseapp.BaseApp, log string) {
		if counter == 0 {
			invariant(t, app, log)
		}
		counter = (counter + 1) % period
	}
}
