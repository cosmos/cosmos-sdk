package systemtests

import (
	"context"
	"testing"

	client "github.com/cometbft/cometbft/rpc/client/http"
	cmtypes "github.com/cometbft/cometbft/types"
	"github.com/stretchr/testify/require"
)

// RPCClient is a test helper to interact with a node via the RPC endpoint.
type RPCClient struct {
	client *client.HTTP
	t      *testing.T
}

// NewRPCClient constructor
func NewRPCClient(t *testing.T, addr string) RPCClient {
	t.Helper()
	httpClient, err := client.New(addr, "/websocket")
	require.NoError(t, err)
	require.NoError(t, httpClient.Start())
	t.Cleanup(func() { _ = httpClient.Stop() })
	return RPCClient{client: httpClient, t: t}
}

// Validators returns list of validators
func (r RPCClient) Validators() []*cmtypes.Validator {
	v, err := r.client.Validators(context.Background(), nil, nil, nil)
	require.NoError(r.t, err)
	return v.Validators
}
