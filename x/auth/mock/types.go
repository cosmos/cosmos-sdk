package mock

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	crypto "github.com/tendermint/go-crypto"
	"math/rand"
	"testing"
)

// Produce a random transaction, and check the state transition. Return a descriptive message "action" about
// what this random tx actually did, for ease of debugging.
type TestAndRunTx func(t *testing.T, r *rand.Rand, app *App, ctx sdk.Context, privKeys []crypto.PrivKey, log string) (action string, err sdk.Error)

// Perform the random setup the module needs.
type RandSetup func(r *rand.Rand, privKeys []crypto.PrivKey)

// Assert invariants for the module. Print out the log when failing
type AssertInvariants func(t *testing.T, app *App, log string)
