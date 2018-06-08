package mock

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	crypto "github.com/tendermint/go-crypto"
	"math/rand"
	"testing"
)

// Produce a random transaction
type TestAndRunTx func(r *rand.Rand, app *App, ctx sdk.Context) (action string, err sdk.Error)

type RandSetup func(r *rand.Rand, privKey []crypto.PrivKey)

type AssertInvariants func(t *testing.T, app *App, log string) bool
