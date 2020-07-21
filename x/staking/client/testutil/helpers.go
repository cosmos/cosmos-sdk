package testutil

import (
	"fmt"

	"github.com/stretchr/testify/require"

	clientkeys "github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/tests"
	"github.com/cosmos/cosmos-sdk/tests/cli"
	sdk "github.com/cosmos/cosmos-sdk/types"
	staking "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// TxStakingCreateValidator is simcli tx staking create-validator
func TxStakingCreateValidator(f *cli.Fixtures, from, consPubKey string, amount sdk.Coin, flags ...string) (bool, string, string) {
	cmd := fmt.Sprintf("%s tx staking create-validator %v --keyring-backend=test --from=%s"+
		" --pubkey=%s", f.SimdBinary, f.Flags(), from, consPubKey)
	cmd += fmt.Sprintf(" --amount=%v --moniker=%v --commission-rate=%v", amount, from, "0.05")
	cmd += fmt.Sprintf(" --commission-max-rate=%v --commission-max-change-rate=%v", "0.20", "0.10")
	cmd += fmt.Sprintf(" --min-self-delegation=%v", "1")

	return cli.ExecuteWriteRetStdStreams(f.T, cli.AddFlags(cmd, flags), clientkeys.DefaultKeyPass)
}

// TxStakingEditValidator is simcli tx staking update validator info
func TxStakingEditValidator(f *cli.Fixtures, from, moniker, website, identity, details string, flags ...string) (bool, string, string) {
	cmd := fmt.Sprintf("%s tx staking edit-validator %v --keyring-backend=test --from=%s", f.SimdBinary, f.Flags(), from)
	cmd += fmt.Sprintf(" --moniker=%v --website=%s", moniker, website)
	cmd += fmt.Sprintf(" --identity=%s --details=%s", identity, details)

	return cli.ExecuteWriteRetStdStreams(f.T, cli.AddFlags(cmd, flags), clientkeys.DefaultKeyPass)
}

// TxStakingUnbond is simcli tx staking unbond
func TxStakingUnbond(f *cli.Fixtures, from, shares string, validator sdk.ValAddress, flags ...string) bool {
	cmd := fmt.Sprintf("%s tx staking unbond --keyring-backend=test %s %v --from=%s %v",
		f.SimdBinary, validator, shares, from, f.Flags())
	return cli.ExecuteWrite(f.T, cli.AddFlags(cmd, flags), clientkeys.DefaultKeyPass)
}

// TxStakingDelegate is simcli tx staking delegate
func TxStakingDelegate(f *cli.Fixtures, from, valOperAddr string, amount sdk.Coin, flags ...string) (bool, string, string) {
	cmd := fmt.Sprintf("%s tx staking delegate %s %v --keyring-backend=test --from=%s %v", f.SimdBinary, valOperAddr, amount, from, f.Flags())
	return cli.ExecuteWriteRetStdStreams(f.T, cli.AddFlags(cmd, flags), clientkeys.DefaultKeyPass)
}

// TxStakingRedelegate is simcli tx staking redelegate
func TxStakingRedelegate(f *cli.Fixtures, from, srcVal, dstVal string, amount sdk.Coin, flags ...string) (bool, string, string) {
	cmd := fmt.Sprintf("%s tx staking redelegate %s %s %v --keyring-backend=test --from=%s %v", f.SimdBinary, srcVal, dstVal, amount, from, f.Flags())
	return cli.ExecuteWriteRetStdStreams(f.T, cli.AddFlags(cmd, flags), clientkeys.DefaultKeyPass)
}

// QueryStakingValidator is simcli query staking validator
func QueryStakingValidator(f *cli.Fixtures, valAddr sdk.ValAddress, flags ...string) staking.Validator {
	cmd := fmt.Sprintf("%s query staking validator %s %v", f.SimdBinary, valAddr, f.Flags())
	out, errStr := tests.ExecuteT(f.T, cli.AddFlags(cmd, flags), "")
	require.Empty(f.T, errStr)

	var validator staking.Validator
	require.NoError(f.T, f.Cdc.UnmarshalJSON([]byte(out), &validator))

	return validator
}

// QueryStakingValidators is simcli query staking validators
func QueryStakingValidators(f *cli.Fixtures, flags ...string) []staking.Validator {
	cmd := fmt.Sprintf("%s query staking validators %v", f.SimdBinary, f.Flags())
	out, errStr := tests.ExecuteT(f.T, cli.AddFlags(cmd, flags), "")
	require.Empty(f.T, errStr)

	var validators []staking.Validator
	require.NoError(f.T, f.Cdc.UnmarshalJSON([]byte(out), &validators))

	return validators
}

// QueryStakingUnbondingDelegationsFrom is simcli query staking unbonding-delegations-from
func QueryStakingUnbondingDelegationsFrom(f *cli.Fixtures, valAddr sdk.ValAddress, flags ...string) []staking.UnbondingDelegation {
	cmd := fmt.Sprintf("%s query staking unbonding-delegations-from %s %v", f.SimdBinary, valAddr, f.Flags())
	out, errStr := tests.ExecuteT(f.T, cli.AddFlags(cmd, flags), "")
	require.Empty(f.T, errStr)

	var ubds []staking.UnbondingDelegation
	require.NoError(f.T, f.Cdc.UnmarshalJSON([]byte(out), &ubds))

	return ubds
}

// QueryStakingDelegationsTo is simcli query staking delegations-to
func QueryStakingDelegationsTo(f *cli.Fixtures, valAddr sdk.ValAddress, flags ...string) []staking.Delegation {
	cmd := fmt.Sprintf("%s query staking delegations-to %s %v", f.SimdBinary, valAddr, f.Flags())
	out, errStr := tests.ExecuteT(f.T, cli.AddFlags(cmd, flags), "")
	require.Empty(f.T, errStr)

	var delegations []staking.Delegation
	require.NoError(f.T, f.Cdc.UnmarshalJSON([]byte(out), &delegations))

	return delegations
}

// QueryStakingPool is simcli query staking pool
func QueryStakingPool(f *cli.Fixtures, flags ...string) staking.Pool {
	cmd := fmt.Sprintf("%s query staking pool %v", f.SimdBinary, f.Flags())
	out, errStr := tests.ExecuteT(f.T, cli.AddFlags(cmd, flags), "")
	require.Empty(f.T, errStr)

	var pool staking.Pool
	require.NoError(f.T, f.Cdc.UnmarshalJSON([]byte(out), &pool))

	return pool
}

// QueryStakingParameters is simcli query staking parameters
func QueryStakingParameters(f *cli.Fixtures, flags ...string) staking.Params {
	cmd := fmt.Sprintf("%s query staking params %v", f.SimdBinary, f.Flags())
	out, errStr := tests.ExecuteT(f.T, cli.AddFlags(cmd, flags), "")
	require.Empty(f.T, errStr)

	var params staking.Params
	require.NoError(f.T, f.Cdc.UnmarshalJSON([]byte(out), &params))

	return params
}

// QueryStakingDelegation is simcli query staking delegation
func QueryStakingDelegation(f *cli.Fixtures, from string, valAddr sdk.ValAddress, flags ...string) staking.Delegation {
	cmd := fmt.Sprintf("%s query staking delegation %s %s %v", f.SimdBinary, from, valAddr, f.Flags())
	out, errStr := tests.ExecuteT(f.T, cli.AddFlags(cmd, flags), "")
	require.Empty(f.T, errStr)

	var delegation staking.Delegation
	require.NoError(f.T, f.Cdc.UnmarshalJSON([]byte(out), &delegation))

	return delegation
}

// QueryStakingDelegations is simcli query staking delegations
func QueryStakingDelegations(f *cli.Fixtures, from string, flags ...string) []staking.Delegation {
	cmd := fmt.Sprintf("%s query staking delegations %s %v", f.SimdBinary, from, f.Flags())
	out, errStr := tests.ExecuteT(f.T, cli.AddFlags(cmd, flags), "")
	require.Empty(f.T, errStr)

	var delegations []staking.Delegation
	require.NoError(f.T, f.Cdc.UnmarshalJSON([]byte(out), &delegations))

	return delegations
}

// QueryStakingRedelegation is simcli query staking redelegation
func QueryStakingRedelegation(f *cli.Fixtures, delAdrr, srcVal, dstVal string, flags ...string) []staking.RedelegationResponse {
	cmd := fmt.Sprintf("%s query staking redelegation %v %v %v %v", f.SimdBinary, delAdrr, srcVal, dstVal, f.Flags())
	out, errStr := tests.ExecuteT(f.T, cli.AddFlags(cmd, flags), "")
	require.Empty(f.T, errStr)

	var redelegations []staking.RedelegationResponse
	require.NoError(f.T, f.Cdc.UnmarshalJSON([]byte(out), &redelegations))

	return redelegations
}

// QueryStakingRedelegations is simcli query staking redelegation
func QueryStakingRedelegations(f *cli.Fixtures, delAdrr string, flags ...string) []staking.RedelegationResponse {
	cmd := fmt.Sprintf("%s query staking redelegations %v %v", f.SimdBinary, delAdrr, f.Flags())
	out, errStr := tests.ExecuteT(f.T, cli.AddFlags(cmd, flags), "")
	require.Empty(f.T, errStr)

	var redelegations []staking.RedelegationResponse
	require.NoError(f.T, f.Cdc.UnmarshalJSON([]byte(out), &redelegations))

	return redelegations
}

// QueryStakingRedelegationsFrom is simcli query staking redelegations-from
func QueryStakingRedelegationsFrom(f *cli.Fixtures, valAddr string, flags ...string) []staking.RedelegationResponse {
	cmd := fmt.Sprintf("%s query staking redelegations-from %v %v", f.SimdBinary, valAddr, f.Flags())
	out, errStr := tests.ExecuteT(f.T, cli.AddFlags(cmd, flags), "")
	require.Empty(f.T, errStr)

	var redelegations []staking.RedelegationResponse
	require.NoError(f.T, f.Cdc.UnmarshalJSON([]byte(out), &redelegations))

	return redelegations
}

// QueryStakingUnbondingDelegation is simcli query staking unbonding-delegation
func QueryStakingUnbondingDelegation(f *cli.Fixtures, delAdrr, valAddr string, flags ...string) staking.UnbondingDelegation {
	cmd := fmt.Sprintf("%s query staking unbonding-delegation %v %v %v", f.SimdBinary, delAdrr, valAddr, f.Flags())
	out, errStr := tests.ExecuteT(f.T, cli.AddFlags(cmd, flags), "")
	require.Empty(f.T, errStr)

	var ubd staking.UnbondingDelegation
	require.NoError(f.T, f.Cdc.UnmarshalJSON([]byte(out), &ubd))

	return ubd
}

// QueryStakingUnbondingDelegations is simcli query staking unbonding-delegations
func QueryStakingUnbondingDelegations(f *cli.Fixtures, delAdrr string, flags ...string) []staking.UnbondingDelegation {
	cmd := fmt.Sprintf("%s query staking unbonding-delegations %v %v", f.SimdBinary, delAdrr, f.Flags())
	out, errStr := tests.ExecuteT(f.T, cli.AddFlags(cmd, flags), "")
	require.Empty(f.T, errStr)

	var ubds []staking.UnbondingDelegation
	require.NoError(f.T, f.Cdc.UnmarshalJSON([]byte(out), &ubds))

	return ubds
}

// QueryStakingHistoricalInfo is simcli query staking historical-info
func QueryStakingHistoricalInfo(f *cli.Fixtures, height uint, flags ...string) staking.HistoricalInfo {
	cmd := fmt.Sprintf("%s query staking historical-info %d %v", f.SimdBinary, height, f.Flags())
	out, errStr := tests.ExecuteT(f.T, cli.AddFlags(cmd, flags), "")
	require.Empty(f.T, errStr)

	var historicalInfo staking.HistoricalInfo
	require.NoError(f.T, f.Cdc.UnmarshalJSON([]byte(out), &historicalInfo))

	return historicalInfo
}
