package testutil

import (
	"fmt"

	"github.com/stretchr/testify/require"

	clientkeys "github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/tests"
	"github.com/cosmos/cosmos-sdk/tests/cli"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
)

// TxStakingCreateValidator is simcli tx staking create-validator
func TxStakingCreateValidator(f *cli.Fixtures, from, consPubKey string, amount sdk.Coin, flags ...string) (bool, string, string) {
	cmd := fmt.Sprintf("%s tx staking create-validator %v --keyring-backend=test --from=%s"+
		" --pubkey=%s", f.SimcliBinary, f.Flags(), from, consPubKey)
	cmd += fmt.Sprintf(" --amount=%v --moniker=%v --commission-rate=%v", amount, from, "0.05")
	cmd += fmt.Sprintf(" --commission-max-rate=%v --commission-max-change-rate=%v", "0.20", "0.10")
	cmd += fmt.Sprintf(" --min-self-delegation=%v", "1")

	return cli.ExecuteWriteRetStdStreams(f.T, cli.AddFlags(cmd, flags), clientkeys.DefaultKeyPass)
}

// TxStakingEditValidator is simcli tx staking update validator info
func TxStakingEditValidator(f *cli.Fixtures, from, moniker, website, identity, details string, flags ...string) (bool, string, string) {
	cmd := fmt.Sprintf("%s tx staking edit-validator %v --keyring-backend=test --from=%s", f.SimcliBinary, f.Flags(), from)
	cmd += fmt.Sprintf(" --moniker=%v --website=%s", moniker, website)
	cmd += fmt.Sprintf(" --identity=%s --details=%s", identity, details)

	return cli.ExecuteWriteRetStdStreams(f.T, cli.AddFlags(cmd, flags), clientkeys.DefaultKeyPass)
}

// TxStakingUnbond is simcli tx staking unbond
func TxStakingUnbond(f *cli.Fixtures, from, shares string, validator sdk.ValAddress, flags ...string) bool {
	cmd := fmt.Sprintf("%s tx staking unbond --keyring-backend=test %s %v --from=%s %v",
		f.SimcliBinary, validator, shares, from, f.Flags())
	return cli.ExecuteWrite(f.T, cli.AddFlags(cmd, flags), clientkeys.DefaultKeyPass)
}

// TxStakingDelegate is simcli tx staking delegate
func TxStakingDelegate(f *cli.Fixtures, from, valOperAddr string, amount sdk.Coin, flags ...string) (bool, string, string) {
	cmd := fmt.Sprintf("%s tx staking delegate %s %v --keyring-backend=test --from=%s %v", f.SimcliBinary, valOperAddr, amount, from, f.Flags())
	return cli.ExecuteWriteRetStdStreams(f.T, cli.AddFlags(cmd, flags), clientkeys.DefaultKeyPass)
}

// TxStakingRedelegate is simcli tx staking redelegate
func TxStakingRedelegate(f *cli.Fixtures, from, srcVal, dstVal string, amount sdk.Coin, flags ...string) (bool, string, string) {
	cmd := fmt.Sprintf("%s tx staking redelegate %s %s %v --keyring-backend=test --from=%s %v", f.SimcliBinary, srcVal, dstVal, amount, from, f.Flags())
	return cli.ExecuteWriteRetStdStreams(f.T, cli.AddFlags(cmd, flags), clientkeys.DefaultKeyPass)
}

// QueryStakingValidator is simcli query staking validator
func QueryStakingValidator(f *cli.Fixtures, valAddr sdk.ValAddress, flags ...string) staking.Validator {
	cmd := fmt.Sprintf("%s query staking validator %s %v", f.SimcliBinary, valAddr, f.Flags())
	out, errStr := tests.ExecuteT(f.T, cli.AddFlags(cmd, flags), "")
	require.Empty(f.T, errStr)

	var validator staking.Validator

	err := f.Cdc.UnmarshalJSON([]byte(out), &validator)
	f.T.Log(fmt.Sprintf("\n out %v\n err %v", out, err))
	require.NoError(f.T, err)

	return validator
}

// QueryStakingValidators is simcli query staking validators
func QueryStakingValidators(f *cli.Fixtures, flags ...string) []staking.Validator {
	cmd := fmt.Sprintf("%s query staking validators %v", f.SimcliBinary, f.Flags())
	out, errStr := tests.ExecuteT(f.T, cli.AddFlags(cmd, flags), "")
	require.Empty(f.T, errStr)

	var validator []staking.Validator

	err := f.Cdc.UnmarshalJSON([]byte(out), &validator)
	f.T.Log(fmt.Sprintf("\n out %v\n err %v", out, err))
	require.NoError(f.T, err)

	return validator
}

// QueryStakingUnbondingDelegationsFrom is simcli query staking unbonding-delegations-from
func QueryStakingUnbondingDelegationsFrom(f *cli.Fixtures, valAddr sdk.ValAddress, flags ...string) []staking.UnbondingDelegation {
	cmd := fmt.Sprintf("%s query staking unbonding-delegations-from %s %v", f.SimcliBinary, valAddr, f.Flags())
	out, errStr := tests.ExecuteT(f.T, cli.AddFlags(cmd, flags), "")
	require.Empty(f.T, errStr)

	var ubds []staking.UnbondingDelegation

	err := f.Cdc.UnmarshalJSON([]byte(out), &ubds)
	f.T.Log(fmt.Sprintf("\n out %v\n err %v", out, err))
	require.NoError(f.T, err)

	return ubds
}

// QueryStakingDelegationsTo is simcli query staking delegations-to
func QueryStakingDelegationsTo(f *cli.Fixtures, valAddr sdk.ValAddress, flags ...string) []staking.Delegation {
	cmd := fmt.Sprintf("%s query staking delegations-to %s %v", f.SimcliBinary, valAddr, f.Flags())
	out, errStr := tests.ExecuteT(f.T, cli.AddFlags(cmd, flags), "")
	require.Empty(f.T, errStr)

	var delegations []staking.Delegation

	err := f.Cdc.UnmarshalJSON([]byte(out), &delegations)
	f.T.Log(fmt.Sprintf("\n out %v\n err %v", out, err))
	require.NoError(f.T, err)

	return delegations
}

// QueryStakingPool is simcli query staking pool
func QueryStakingPool(f *cli.Fixtures, flags ...string) staking.Pool {
	cmd := fmt.Sprintf("%s query staking pool %v", f.SimcliBinary, f.Flags())
	out, errStr := tests.ExecuteT(f.T, cli.AddFlags(cmd, flags), "")
	require.Empty(f.T, errStr)

	var pool staking.Pool

	err := f.Cdc.UnmarshalJSON([]byte(out), &pool)
	f.T.Log(fmt.Sprintf("\n out %v\n err %v", out, err))
	require.NoError(f.T, err)

	return pool
}

// QueryStakingParameters is simcli query staking parameters
func QueryStakingParameters(f *cli.Fixtures, flags ...string) staking.Params {
	cmd := fmt.Sprintf("%s query staking params %v", f.SimcliBinary, f.Flags())
	out, errStr := tests.ExecuteT(f.T, cli.AddFlags(cmd, flags), "")
	require.Empty(f.T, errStr)

	var params staking.Params

	err := f.Cdc.UnmarshalJSON([]byte(out), &params)
	f.T.Log(fmt.Sprintf("\n out %v\n err %v", out, err))
	require.NoError(f.T, err)

	return params
}

// QueryStakingDelegation is simcli query staking delegation
func QueryStakingDelegation(f *cli.Fixtures, from string, valAddr sdk.ValAddress, flags ...string) staking.Delegation {
	cmd := fmt.Sprintf("%s query staking delegation %s %s %v", f.SimcliBinary, from, valAddr, f.Flags())
	out, errStr := tests.ExecuteT(f.T, cli.AddFlags(cmd, flags), "")
	require.Empty(f.T, errStr)

	var delegation staking.Delegation

	err := f.Cdc.UnmarshalJSON([]byte(out), &delegation)
	f.T.Log(fmt.Sprintf("\n out %v\n err %v", out, err))
	require.NoError(f.T, err)

	return delegation
}

// QueryStakingDelegations is simcli query staking delegations
func QueryStakingDelegations(f *cli.Fixtures, from string, flags ...string) []staking.Delegation {
	cmd := fmt.Sprintf("%s query staking delegations %s %v", f.SimcliBinary, from, f.Flags())
	out, errStr := tests.ExecuteT(f.T, cli.AddFlags(cmd, flags), "")
	require.Empty(f.T, errStr)

	var delegations []staking.Delegation

	err := f.Cdc.UnmarshalJSON([]byte(out), &delegations)
	f.T.Log(fmt.Sprintf("\n out %v\n err %v", out, err))
	require.NoError(f.T, err)

	return delegations
}

// QueryStakingRedelegation is simcli query staking redelegation
func QueryStakingRedelegation(f *cli.Fixtures, delAdrr, srcVal, dstVal string, flags ...string) []staking.RedelegationResponse {
	cmd := fmt.Sprintf("%s query staking redelegation %v %v %v %v", f.SimcliBinary, delAdrr, srcVal, dstVal, f.Flags())
	out, errStr := tests.ExecuteT(f.T, cli.AddFlags(cmd, flags), "")
	require.Empty(f.T, errStr)

	var redelegations []staking.RedelegationResponse

	err := f.Cdc.UnmarshalJSON([]byte(out), &redelegations)
	f.T.Log(fmt.Sprintf("\n out %v\n err %v", out, err))
	require.NoError(f.T, err)

	return redelegations
}

// QueryStakingRedelegations is simcli query staking redelegation
func QueryStakingRedelegations(f *cli.Fixtures, delAdrr string, flags ...string) []staking.RedelegationResponse {
	cmd := fmt.Sprintf("%s query staking redelegations %v %v", f.SimcliBinary, delAdrr, f.Flags())
	out, errStr := tests.ExecuteT(f.T, cli.AddFlags(cmd, flags), "")
	require.Empty(f.T, errStr)

	var redelegations []staking.RedelegationResponse

	err := f.Cdc.UnmarshalJSON([]byte(out), &redelegations)
	f.T.Log(fmt.Sprintf("\n out %v\n err %v", out, err))
	require.NoError(f.T, err)

	return redelegations
}

// QueryStakingRedelegationsFrom is simcli query staking redelegation
func QueryStakingRedelegationsFrom(f *cli.Fixtures, valAddr string, flags ...string) []staking.RedelegationResponse {
	cmd := fmt.Sprintf("%s query staking redelegations-from %v %v", f.SimcliBinary, valAddr, f.Flags())
	out, errStr := tests.ExecuteT(f.T, cli.AddFlags(cmd, flags), "")
	require.Empty(f.T, errStr)

	var redelegations []staking.RedelegationResponse

	err := f.Cdc.UnmarshalJSON([]byte(out), &redelegations)
	f.T.Log(fmt.Sprintf("\n out %v\n err %v", out, err))
	require.NoError(f.T, err)

	return redelegations
}

// QueryStakingUnbondingDelegation is simcli query staking redelegation
func QueryStakingUnbondingDelegation(f *cli.Fixtures, delAdrr, valAddr string, flags ...string) staking.UnbondingDelegation {
	cmd := fmt.Sprintf("%s query staking unbonding-delegation %v %v %v", f.SimcliBinary, delAdrr, valAddr, f.Flags())
	out, err := tests.ExecuteT(f.T, cli.AddFlags(cmd, flags), "")

	fmt.Println("Error staking unbonding", err)
	var ubd staking.UnbondingDelegation

	err1 := f.Cdc.UnmarshalJSON([]byte(out), &ubd)
	require.NoError(f.T, err1, "out %v\n, err %v", out, err)

	return ubd
}

// QueryStakingUnbondingDelegations is simcli query staking redelegation
func QueryStakingUnbondingDelegations(f *cli.Fixtures, delAdrr string, flags ...string) []staking.UnbondingDelegation {
	cmd := fmt.Sprintf("%s query staking unbonding-delegations %v %v", f.SimcliBinary, delAdrr, f.Flags())
	out, errStr := tests.ExecuteT(f.T, cli.AddFlags(cmd, flags), "")
	require.Empty(f.T, errStr)

	var ubds []staking.UnbondingDelegation

	err := f.Cdc.UnmarshalJSON([]byte(out), &ubds)
	f.T.Log(fmt.Sprintf("\n out %v\n err %v", out, err))
	require.NoError(f.T, err)

	return ubds
}
