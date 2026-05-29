package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	testapp "github.com/cosmos/cosmos-sdk/testutil/testapp"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
)

func TestAccountRetriever(t *testing.T) {
	cfg := network.DefaultConfig(testapp.SDKAppFixture)
	cfg.NumValidators = 1

	net, err := network.New(t, t.TempDir(), cfg)
	require.NoError(t, err)
	defer net.Cleanup()

	_, err = net.WaitForHeight(3)
	require.NoError(t, err)

	val := net.Validators[0]
	clientCtx := val.ClientCtx
	ar := types.AccountRetriever{}

	clientCtx = clientCtx.WithHeight(2)

	acc, err := ar.GetAccount(clientCtx, val.Address)
	require.NoError(t, err)
	require.NotNil(t, acc)

	acc, height, err := ar.GetAccountWithHeight(clientCtx, val.Address)
	require.NoError(t, err)
	require.NotNil(t, acc)
	require.Equal(t, height, int64(2))

	require.NoError(t, ar.EnsureExists(clientCtx, val.Address))

	accNum, accSeq, err := ar.GetAccountNumberSequence(clientCtx, val.Address)
	require.NoError(t, err)
	require.Equal(t, accNum, uint64(0))
	require.Equal(t, accSeq, uint64(1))
}
