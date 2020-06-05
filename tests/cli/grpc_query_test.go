// +build cli_test

package cli

import (
	"context"
	"testing"

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
