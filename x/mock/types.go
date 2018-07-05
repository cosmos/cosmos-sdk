package mock

import (
	"math/rand"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/crypto"
)

type (
	// TestAndRunTx produces a fuzzed transaction, and ensures the state
	// transition was as expected. It returns a descriptive message "action"
	// about what this fuzzed tx actually did, for ease of debugging.
	TestAndRunTx func(
		t *testing.T, r *rand.Rand, app *App, ctx sdk.Context,
		privKeys []crypto.PrivKey, log string,
	) (action string, err sdk.Error)

	// RandSetup performs the random setup the mock module needs.
	RandSetup func(r *rand.Rand, privKeys []crypto.PrivKey)

	// An Invariant is a function which tests a particular invariant.
	// If the invariant has been broken, the function should halt the
	// test and output the log.
	Invariant func(t *testing.T, app *App, log string)
)

// PeriodicInvariant returns an Invariant function closure that asserts
// a given invariant if the mock application's last block modulo the given
// period is congruent to the given offset.
func PeriodicInvariant(invariant Invariant, period int, offset int) Invariant {
	return func(t *testing.T, app *App, log string) {
		if int(app.LastBlockHeight())%period == offset {
			invariant(t, app, log)
		}
	}
}
