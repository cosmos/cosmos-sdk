package mock

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	crypto "github.com/tendermint/go-crypto"
	"math/rand"
	"testing"
)

// Produce a random transaction
type TestAndRunTx func(t *testing.T, r *rand.Rand, app *App, ctx sdk.Context, privKeys []crypto.PrivKey, log string) (action string, err sdk.Error)

type RandSetup func(r *rand.Rand, privKeys []crypto.PrivKey)

type AssertInvariants func(t *testing.T, app *App, log string)
