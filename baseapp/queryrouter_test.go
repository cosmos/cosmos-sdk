package baseapp

import (
	"context"
	"fmt"
	"testing"

	"github.com/cosmos/cosmos-sdk/codec/testdata"

	"github.com/stretchr/testify/require"

	abci "github.com/tendermint/tendermint/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var testQuerier = func(_ sdk.Context, _ []string, _ abci.RequestQuery) ([]byte, error) {
	return nil, nil
}

func TestQueryRouter(t *testing.T) {
	qr := NewQueryRouter()

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

type echoer struct{}

func (e echoer) Echo(_ context.Context, req *testdata.EchoRequest) (*testdata.EchoResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("empty request")
	}
	return &testdata.EchoResponse{Message: req.Message}, nil
}

var _ testdata.EchoServiceServer = echoer{}

func TestRegisterQueryService(t *testing.T) {
	qr := NewQueryRouter()
	testdata.RegisterEchoServiceServer(qr, echoer{})
	helper := &QueryServiceTestHelper{
		QueryRouter: qr,
		ctx:         sdk.Context{},
	}
	client := testdata.NewEchoServiceClient(helper)

	res, err := client.Echo(context.Background(), &testdata.EchoRequest{Message: "hello"})
	require.Nil(t, err)
	require.NotNil(t, res)
	require.Equal(t, "hello", res.Message)

	_, err = client.Echo(context.Background(), nil)
	require.NotNil(t, err)

}
