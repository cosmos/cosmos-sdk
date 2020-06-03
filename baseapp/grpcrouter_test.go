package baseapp

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/codec/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type testServer struct{}

func (e testServer) Echo(_ context.Context, req *testdata.EchoRequest) (*testdata.EchoResponse, error) {
	return &testdata.EchoResponse{Message: req.Message}, nil
}

func (e testServer) SayHello(_ context.Context, request *testdata.SayHelloRequest) (*testdata.SayHelloResponse, error) {
	greeting := fmt.Sprintf("Hello %s!", request.Name)
	return &testdata.SayHelloResponse{Greeting: greeting}, nil
}

var _ testdata.TestServiceServer = testServer{}

func TestGRPCRouter(t *testing.T) {
	qr := NewGRPCRouter()
	testdata.RegisterTestServiceServer(qr, testServer{})
	helper := &QueryServiceTestHelper{
		GRPCRouter: qr,
		ctx:        sdk.Context{},
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
