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

// TxWithdrawRewards raises a txn to withdraw rewards
func TxWithdrawRewards(f *cli.Fixtures, valAddr sdk.ValAddress, from string, flags ...string) bool {
	cmd := fmt.Sprintf("%s tx distribution withdraw-rewards %s %v --keyring-backend=test --from=%s", f.SimcliBinary, valAddr, f.Flags(), from)
	return cli.ExecuteWrite(f.T, cli.AddFlags(cmd, flags), clientkeys.DefaultKeyPass)
}

// TxSetWithdrawAddress helps to set the withdraw address for rewards associated with a delegator address
func TxSetWithdrawAddress(f *cli.Fixtures, from, withDrawAddr string, flags ...string) (bool, string, string) {
	cmd := fmt.Sprintf("%s tx distribution set-withdraw-addr %s --from %s %v --keyring-backend=test", f.SimcliBinary, withDrawAddr, from, f.Flags())
	return cli.ExecuteWriteRetStdStreams(f.T, cli.AddFlags(cmd, flags), clientkeys.DefaultKeyPass)
}

// TxWithdrawAllRewards raises a txn to withdraw rewards
func TxWithdrawAllRewards(f *cli.Fixtures, from string, flags ...string) (bool, string, string) {
	cmd := fmt.Sprintf("%s tx distribution withdraw-all-rewards %v --keyring-backend=test --from=%s", f.SimcliBinary, f.Flags(), from)
	return cli.ExecuteWriteRetStdStreams(f.T, cli.AddFlags(cmd, flags), clientkeys.DefaultKeyPass)
}

// QueryRewards returns the rewards of a delegator
func QueryRewards(f *cli.Fixtures, delAddr sdk.AccAddress, flags ...string) distribution.QueryDelegatorTotalRewardsResponse {
	cmd := fmt.Sprintf("%s query distribution rewards %s %s", f.SimcliBinary, delAddr, f.Flags())
	res, errStr := tests.ExecuteT(f.T, cmd, "")
	require.Empty(f.T, errStr)

	var rewards distribution.QueryDelegatorTotalRewardsResponse
	err := f.Cdc.UnmarshalJSON([]byte(res), &rewards)
	f.T.Log(fmt.Sprintf("\n out %v\n err %v", res, err))
	require.NoError(f.T, err)
	return rewards
}

// QueryValidatorOutstandingRewards distribution outstanding (un-withdrawn) rewards
func QueryValidatorOutstandingRewards(f *cli.Fixtures, valAddr string) distribution.ValidatorOutstandingRewards {
	cmd := fmt.Sprintf("%s query distribution validator-outstanding-rewards %s %v", f.SimcliBinary, valAddr, f.Flags())
	res, errStr := tests.ExecuteT(f.T, cmd, "")
	require.Empty(f.T, errStr)

	var outstandingRewards distribution.ValidatorOutstandingRewards
	err := f.Cdc.UnmarshalJSON([]byte(res), &outstandingRewards)
	f.T.Log(fmt.Sprintf("\n out %v\n err %v", res, err))
	require.NoError(f.T, err)
	return outstandingRewards
}

// QueryParameters is simcli query distribution parameters
func QueryParameters(f *cli.Fixtures, flags ...string) distribution.Params {
	cmd := fmt.Sprintf("%s query distribution params %v", f.SimcliBinary, f.Flags())
	out, errStr := tests.ExecuteT(f.T, cli.AddFlags(cmd, flags), "")
	require.Empty(f.T, errStr)

	var params distribution.Params
	err := f.Cdc.UnmarshalJSON([]byte(out), &params)
	f.T.Log(fmt.Sprintf("\n out %v\n err %v", out, err))
	require.NoError(f.T, err)
	return params
}

// QueryCommission returns validator commission rewards from delegators to that validator.
func QueryCommission(f *cli.Fixtures, valAddr string, flags ...string) distribution.ValidatorAccumulatedCommission {
	cmd := fmt.Sprintf("%s query distribution commission %s %v", f.SimcliBinary, valAddr, f.Flags())
	out, errStr := tests.ExecuteT(f.T, cli.AddFlags(cmd, flags), "")
	require.Empty(f.T, errStr)

	var commission distribution.ValidatorAccumulatedCommission
	err := f.Cdc.UnmarshalJSON([]byte(out), &commission)
	f.T.Log(fmt.Sprintf("\n out %v\n err %v", out, err))
	require.NoError(f.T, err)
	return commission
}

// QuerySlashes returns all slashes of a validator for a given block range.
func QuerySlashes(f *cli.Fixtures, valAddr string, flags ...string) []distribution.ValidatorSlashEvent {
	cmd := fmt.Sprintf("%s query distribution slashes %s 0 5 %v ", f.SimcliBinary, valAddr, f.Flags())
	out, errStr := tests.ExecuteT(f.T, cli.AddFlags(cmd, flags), "")
	fmt.Println("errStr---------------->", errStr)
	require.Empty(f.T, errStr)

	var slashes []distribution.ValidatorSlashEvent
	err := f.Cdc.UnmarshalJSON([]byte(out), &slashes)
	f.T.Log(fmt.Sprintf("\n out %v\n err %v", out, err))
	require.NoError(f.T, err)
	return slashes
}
