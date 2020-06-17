// +build cli_test

package cli

import (
	"context"
	"testing"

	"google.golang.org/grpc"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec/testdata"
)

func TestCliQueryConn(t *testing.T) {
	t.Parallel()
	f := NewFixtures(t)

	// start simd server
	proc := f.SDStart()
	t.Cleanup(func() { proc.Stop(false) })

	ctx := client.NewContext()
	testClient := testdata.NewTestServiceClient(ctx)
	res, err := testClient.Echo(context.Background(), &testdata.EchoRequest{Message: "hello"})
	require.NoError(t, err)
	require.Equal(t, "hello", res.Message)
}

func TestGRPCProxy(t *testing.T) {
	t.Parallel()
	f := NewFixtures(t)

	// start simd server
	proc := f.SDStart()
	t.Cleanup(func() { proc.Stop(false) })

	conn, err := grpc.Dial("tcp://0.0.0.0:9090")
	require.NoError(t, err)
	testClient := testdata.NewTestServiceClient(conn)
	res, err := testClient.Echo(context.Background(), &testdata.EchoRequest{Message: "hello"})
	require.NoError(t, err)
	require.Equal(t, "hello", res.Message)
}
