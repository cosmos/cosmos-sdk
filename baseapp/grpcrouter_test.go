package baseapp

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/codec/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestGRPCRouter(t *testing.T) {
	qr := NewGRPCQueryRouter()
	testdata.RegisterTestServiceServer(qr, testdata.TestServiceImpl{})
	helper := &QueryServiceTestHelper{
		GRPCQueryRouter: qr,
		ctx:             sdk.Context{},
	}
	client := testdata.NewTestServiceClient(helper)

	res, err := client.Echo(context.Background(), &testdata.EchoRequest{Message: "hello"})
	require.Nil(t, err)
	require.NotNil(t, res)
	require.Equal(t, "hello", res.Message)

	require.Panics(t, func() {
		_, _ = client.Echo(context.Background(), nil)
	})

	res2, err := client.SayHello(context.Background(), &testdata.SayHelloRequest{Name: "Foo"})
	require.Nil(t, err)
	require.NotNil(t, res)
	require.Equal(t, "Hello Foo!", res2.Greeting)
}
