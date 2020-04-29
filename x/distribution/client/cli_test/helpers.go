package cli

import (
	"fmt"
	clientkeys "github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/tests"
	"github.com/cosmos/cosmos-sdk/tests/cli/helpers"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/distribution"
	"github.com/stretchr/testify/require"
)

func TxWithdrawRewards(f *helpers.Fixtures, valAddr sdk.ValAddress, from string, flags ...string) bool {
	cmd := fmt.Sprintf("%s tx distribution withdraw-rewards %s %v --keyring-backend=test --from=%s", f.SimcliBinary, valAddr, f.Flags(), from)
	return helpers.ExecuteWrite(f.T, helpers.AddFlags(cmd, flags), clientkeys.DefaultKeyPass)
}

// QueryRewards returns the rewards of a delegator
func QueryRewards(f *helpers.Fixtures, delAddr sdk.AccAddress, flags ...string) distribution.QueryDelegatorTotalRewardsResponse {
	cmd := fmt.Sprintf("%s query distribution rewards %s %s", f.SimcliBinary, delAddr, f.Flags())
	res, errStr := tests.ExecuteT(f.T, cmd, "")
	require.Empty(f.T, errStr)

	var rewards distribution.QueryDelegatorTotalRewardsResponse
	err := f.Cdc.UnmarshalJSON([]byte(res), &rewards)
	require.NoError(f.T, err)
	return rewards
}
