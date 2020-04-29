package cli_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/tests/cli/helpers"
	sdk "github.com/cosmos/cosmos-sdk/types"
	mintcli "github.com/cosmos/cosmos-sdk/x/mint/client/cli_test"
)

func TestGaiaCLIQuerySupply(t *testing.T) {
	t.Parallel()
	f := helpers.InitFixtures(t)

	// start simd server
	proc := f.SDStart()
	defer proc.Stop(false)

	totalSupply := mintcli.QueryTotalSupply(f)
	totalSupplyOf := mintcli.QueryTotalSupplyOf(f, helpers.FooDenom)

	require.Equal(t, helpers.TotalCoins, totalSupply)
	require.True(sdk.IntEq(t, helpers.TotalCoins.AmountOf(helpers.FooDenom), totalSupplyOf))

	f.Cleanup()
}
