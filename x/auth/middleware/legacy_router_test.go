package middleware_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/middleware"
)

var testHandler = func(_ sdk.Context, _ sdk.Msg) (*sdk.Result, error) {
	return &sdk.Result{}, nil
}

func TestLegacyRouter(t *testing.T) {
	rtr := middleware.NewLegacyRouter()

	// require panic on invalid route
	require.Panics(t, func() {
		rtr.AddRoute(sdk.NewRoute("*", testHandler))
	})

	rtr.AddRoute(sdk.NewRoute("testRoute", testHandler))
	h := rtr.Route(sdk.Context{}, "testRoute")
	require.NotNil(t, h)

	// require panic on duplicate route
	require.Panics(t, func() {
		rtr.AddRoute(sdk.NewRoute("testRoute", testHandler))
	})
}
