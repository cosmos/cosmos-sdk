package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/x/auth/testutil"
	"cosmossdk.io/x/auth/types"

	"github.com/cosmos/cosmos-sdk/testutil/network"
)

func TestAccountRetriever(t *testing.T) {
	cfg, err := network.DefaultConfigWithAppConfig(testutil.AppConfig)
	require.NoError(t, err)
	cfg.NumValidators = 1

	network, err := network.New(t, t.TempDir(), cfg)
	require.NoError(t, err)
	defer network.Cleanup()

	_, err = network.WaitForHeight(3)
	require.NoError(t, err)

	val := network.GetValidators()[0]
	clientCtx := val.GetClientCtx()
	ar := types.AccountRetriever{}

	clientCtx = clientCtx.WithHeight(2)

	acc, err := ar.GetAccount(clientCtx, val.GetAddress())
	require.NoError(t, err)
	require.NotNil(t, acc)

	acc, height, err := ar.GetAccountWithHeight(clientCtx, val.GetAddress())
	require.NoError(t, err)
	require.NotNil(t, acc)
	require.Equal(t, height, int64(2))

	require.NoError(t, ar.EnsureExists(clientCtx, val.GetAddress()))

	accNum, accSeq, err := ar.GetAccountNumberSequence(clientCtx, val.GetAddress())
	require.NoError(t, err)
	require.Equal(t, accNum, uint64(0))
	require.Equal(t, accSeq, uint64(1))
}
