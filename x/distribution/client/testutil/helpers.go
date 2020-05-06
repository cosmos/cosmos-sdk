package testutil

import (
	"fmt"

	"github.com/stretchr/testify/require"

	clientkeys "github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/tests"
	"github.com/cosmos/cosmos-sdk/tests/cli"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/distribution"
)

func TxWithdrawRewards(f *cli.Fixtures, valAddr sdk.ValAddress, from string, flags ...string) bool {
	cmd := fmt.Sprintf("%s tx distribution withdraw-rewards %s %v --keyring-backend=test --from=%s", f.SimcliBinary, valAddr, f.Flags(), from)
	return cli.ExecuteWrite(f.T, cli.AddFlags(cmd, flags), clientkeys.DefaultKeyPass)
}

// QueryRewards returns the rewards of a delegator
func QueryRewards(f *cli.Fixtures, delAddr sdk.AccAddress, flags ...string) distribution.QueryDelegatorTotalRewardsResponse {
	cmd := fmt.Sprintf("%s query distribution rewards %s %s", f.SimcliBinary, delAddr, f.Flags())
	res, errStr := tests.ExecuteT(f.T, cmd, "")
	require.Empty(f.T, errStr)

	var rewards distribution.QueryDelegatorTotalRewardsResponse
	err := f.Cdc.UnmarshalJSON([]byte(res), &rewards)
	require.NoError(f.T, err)
	return rewards
}
