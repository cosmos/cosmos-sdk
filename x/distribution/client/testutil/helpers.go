package testutil

import (
	"fmt"

	"github.com/stretchr/testify/require"

	clientkeys "github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/tests"
	"github.com/cosmos/cosmos-sdk/tests/cli"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
)

// TxWithdrawRewards raises a txn to withdraw rewards
func TxWithdrawRewards(f *cli.Fixtures, valAddr sdk.ValAddress, from string, flags ...string) bool {
	cmd := fmt.Sprintf("%s tx distribution withdraw-rewards %s %v --keyring-backend=test --from=%s", f.SimdBinary, valAddr, f.Flags(), from)
	return cli.ExecuteWrite(f.T, cli.AddFlags(cmd, flags), clientkeys.DefaultKeyPass)
}

// TxSetWithdrawAddress helps to set the withdraw address for rewards associated with a delegator address
func TxSetWithdrawAddress(f *cli.Fixtures, from, withDrawAddr string, flags ...string) (bool, string, string) {
	cmd := fmt.Sprintf("%s tx distribution set-withdraw-addr %s --from %s %v --keyring-backend=test", f.SimdBinary, withDrawAddr, from, f.Flags())
	return cli.ExecuteWriteRetStdStreams(f.T, cli.AddFlags(cmd, flags), clientkeys.DefaultKeyPass)
}

// TxWithdrawAllRewards raises a txn to withdraw all rewards of a delegator address
func TxWithdrawAllRewards(f *cli.Fixtures, from string, flags ...string) (bool, string, string) {
	cmd := fmt.Sprintf("%s tx distribution withdraw-all-rewards %v --keyring-backend=test --from=%s", f.SimdBinary, f.Flags(), from)
	return cli.ExecuteWriteRetStdStreams(f.T, cli.AddFlags(cmd, flags), clientkeys.DefaultKeyPass)
}

// TxFundCommunityPool Funds the community pool with the specified amount
func TxFundCommunityPool(f *cli.Fixtures, from string, amount sdk.Coin, flags ...string) (bool, string, string) {
	cmd := fmt.Sprintf("%s tx distribution fund-community-pool %v %v --keyring-backend=test --from=%s", f.SimdBinary, amount, f.Flags(), from)
	return cli.ExecuteWriteRetStdStreams(f.T, cli.AddFlags(cmd, flags), clientkeys.DefaultKeyPass)
}

// QueryRewards returns the rewards of a delegator
func QueryRewards(f *cli.Fixtures, delAddr sdk.AccAddress, flags ...string) types.QueryDelegatorTotalRewardsResponse {
	cmd := fmt.Sprintf("%s query distribution rewards %s %s", f.SimdBinary, delAddr, f.Flags())
	res, errStr := tests.ExecuteT(f.T, cmd, "")
	require.Empty(f.T, errStr)

	var rewards types.QueryDelegatorTotalRewardsResponse
	require.NoError(f.T, f.Cdc.UnmarshalJSON([]byte(res), &rewards))
	return rewards
}

// QueryValidatorOutstandingRewards distribution outstanding (un-withdrawn) rewards
func QueryValidatorOutstandingRewards(f *cli.Fixtures, valAddr string) types.ValidatorOutstandingRewards {
	cmd := fmt.Sprintf("%s query distribution validator-outstanding-rewards %s %v", f.SimdBinary, valAddr, f.Flags())
	res, errStr := tests.ExecuteT(f.T, cmd, "")
	require.Empty(f.T, errStr)

	var outstandingRewards types.ValidatorOutstandingRewards
	require.NoError(f.T, f.Cdc.UnmarshalJSON([]byte(res), &outstandingRewards))
	return outstandingRewards
}

// QueryParameters is simcli query distribution parameters
func QueryParameters(f *cli.Fixtures, flags ...string) types.Params {
	cmd := fmt.Sprintf("%s query distribution params %v", f.SimdBinary, f.Flags())
	out, errStr := tests.ExecuteT(f.T, cli.AddFlags(cmd, flags), "")
	require.Empty(f.T, errStr)

	var params types.Params
	require.NoError(f.T, f.Cdc.UnmarshalJSON([]byte(out), &params))
	return params
}

// QueryCommission returns validator commission rewards from delegators to that validator.
func QueryCommission(f *cli.Fixtures, valAddr string, flags ...string) types.ValidatorAccumulatedCommission {
	cmd := fmt.Sprintf("%s query distribution commission %s %v", f.SimdBinary, valAddr, f.Flags())
	out, errStr := tests.ExecuteT(f.T, cli.AddFlags(cmd, flags), "")
	require.Empty(f.T, errStr)

	var commission types.ValidatorAccumulatedCommission
	require.NoError(f.T, f.Cdc.UnmarshalJSON([]byte(out), &commission))
	return commission
}

// QuerySlashes returns all slashes of a validator for a given block range.
func QuerySlashes(f *cli.Fixtures, valAddr string, flags ...string) []types.ValidatorSlashEvent {
	cmd := fmt.Sprintf("%s query distribution slashes %s 0 5 %v ", f.SimdBinary, valAddr, f.Flags())
	out, errStr := tests.ExecuteT(f.T, cli.AddFlags(cmd, flags), "")
	require.Empty(f.T, errStr)

	var slashes []types.ValidatorSlashEvent
	require.NoError(f.T, f.Cdc.UnmarshalJSON([]byte(out), &slashes))
	return slashes
}

// QueryCommunityPool returns the amount of coins in the community pool
func QueryCommunityPool(f *cli.Fixtures, flags ...string) sdk.DecCoins {
	cmd := fmt.Sprintf("%s query distribution community-pool %v ", f.SimdBinary, f.Flags())
	out, errStr := tests.ExecuteT(f.T, cli.AddFlags(cmd, flags), "")
	require.Empty(f.T, errStr)

	var amount sdk.DecCoins
	require.NoError(f.T, f.Cdc.UnmarshalJSON([]byte(out), &amount))
	return amount
}
