package baseapp_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var testQuerier = func(_ sdk.Context, _ []string, _ abci.RequestQuery) ([]byte, error) {
	return nil, nil
}

func TestQueryRouter(t *testing.T) {
	qr := baseapp.NewQueryRouter()

	// require panic on invalid route
	require.Panics(t, func() {
		qr.AddRoute("*", testQuerier)
	})

	qr.AddRoute("testRoute", testQuerier)
	q := qr.Route("testRoute")
	require.NotNil(t, q)

	// require panic on duplicate route
	require.Panics(t, func() {
		qr.AddRoute("testRoute", testQuerier)
	})
}
