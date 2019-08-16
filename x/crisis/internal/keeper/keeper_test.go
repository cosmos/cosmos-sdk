package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
)

func TestLogger(t *testing.T) {
	app := createTestApp()

	ctx := app.NewContext(true, abci.Header{})
	require.Equal(t, ctx.Logger(), app.CrisisKeeper.Logger(ctx))
}

func TestInvariants(t *testing.T) {
	app := createTestApp()
	require.Equal(t, app.CrisisKeeper.InvCheckPeriod(), uint(5))

	// SimApp has 11 registered invariants
	orgInvRoutes := app.CrisisKeeper.Routes()
	app.CrisisKeeper.RegisterRoute("testModule", "testRoute", testPassingInvariant)
	require.Equal(t, len(app.CrisisKeeper.Routes()), len(orgInvRoutes)+1)
}

func TestAssertInvariants(t *testing.T) {
	app := createTestApp()
	ctx := app.NewContext(true, abci.Header{})

	app.CrisisKeeper.RegisterRoute("testModule", "testRoute1", testPassingInvariant)
	require.NotPanics(t, func() { app.CrisisKeeper.AssertInvariants(ctx) })

	app.CrisisKeeper.RegisterRoute("testModule", "testRoute2", testFailingInvariant)
	require.Panics(t, func() { app.CrisisKeeper.AssertInvariants(ctx) })
}
