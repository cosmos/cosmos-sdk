package baseapp

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var testHandler = func(_ sdk.Context, _ sdk.Msg) sdk.Result {
	return sdk.Result{}
}

func TestRouter(t *testing.T) {
	rtr := NewRouter()

	// require panic on invalid route
	require.Panics(t, func() {
		rtr.AddRoute("*", testHandler)
	})

	rtr.AddRoute("testRoute", testHandler)
	h := rtr.Route("testRoute")
	require.NotNil(t, h)

	// require panic on duplicate route
	require.Panics(t, func() {
		rtr.AddRoute("testRoute", testHandler)
	})
}
