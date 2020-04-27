package cli

import (
	"fmt"
	clientkeys "github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/tests"
	"github.com/cosmos/cosmos-sdk/tests/cli/helpers"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/stretchr/testify/require"
)

// TxStakingCreateValidator is simcli tx staking create-validator
func TxStakingCreateValidator(f *helpers.Fixtures, from, consPubKey string, amount sdk.Coin, flags ...string) (bool, string, string) {
	cmd := fmt.Sprintf("%s tx staking create-validator %v --keyring-backend=test --from=%s"+
		" --pubkey=%s", f.SimcliBinary, f.Flags(), from, consPubKey)
	cmd += fmt.Sprintf(" --amount=%v --moniker=%v --commission-rate=%v", amount, from, "0.05")
	cmd += fmt.Sprintf(" --commission-max-rate=%v --commission-max-change-rate=%v", "0.20", "0.10")
	cmd += fmt.Sprintf(" --min-self-delegation=%v", "1")
	return helpers.ExecuteWriteRetStdStreams(f.T, helpers.AddFlags(cmd, flags), clientkeys.DefaultKeyPass)
}

// TxStakingUnbond is simcli tx staking unbond
func TxStakingUnbond(f *helpers.Fixtures, from, shares string, validator sdk.ValAddress, flags ...string) bool {
	cmd := fmt.Sprintf("%s tx staking unbond --keyring-backend=test %s %v --from=%s %v",
		f.SimcliBinary, validator, shares, from, f.Flags())
	return helpers.ExecuteWrite(f.T, helpers.AddFlags(cmd, flags), clientkeys.DefaultKeyPass)
}

// QueryStakingValidator is simcli query staking validator
func QueryStakingValidator(f *helpers.Fixtures, valAddr sdk.ValAddress, flags ...string) staking.Validator {
	cmd := fmt.Sprintf("%s query staking validator %s %v", f.SimcliBinary, valAddr, f.Flags())
	out, _ := tests.ExecuteT(f.T, helpers.AddFlags(cmd, flags), "")
	var validator staking.Validator

	err := f.Cdc.UnmarshalJSON([]byte(out), &validator)
	require.NoError(f.T, err, "out %v\n, err %v", out, err)
	return validator
}

// QueryStakingUnbondingDelegationsFrom is simcli query staking unbonding-delegations-from
func QueryStakingUnbondingDelegationsFrom(f *helpers.Fixtures, valAddr sdk.ValAddress, flags ...string) []staking.UnbondingDelegation {
	cmd := fmt.Sprintf("%s query staking unbonding-delegations-from %s %v", f.SimcliBinary, valAddr, f.Flags())
	out, _ := tests.ExecuteT(f.T, helpers.AddFlags(cmd, flags), "")
	var ubds []staking.UnbondingDelegation

	err := f.Cdc.UnmarshalJSON([]byte(out), &ubds)
	require.NoError(f.T, err, "out %v\n, err %v", out, err)
	return ubds
}

// QueryStakingDelegationsTo is simcli query staking delegations-to
func QueryStakingDelegationsTo(f *helpers.Fixtures, valAddr sdk.ValAddress, flags ...string) []staking.Delegation {
	cmd := fmt.Sprintf("%s query staking delegations-to %s %v", f.SimcliBinary, valAddr, f.Flags())
	out, _ := tests.ExecuteT(f.T, helpers.AddFlags(cmd, flags), "")
	var delegations []staking.Delegation

	err := f.Cdc.UnmarshalJSON([]byte(out), &delegations)
	require.NoError(f.T, err, "out %v\n, err %v", out, err)
	return delegations
}

// QueryStakingPool is simcli query staking pool
func QueryStakingPool(f *helpers.Fixtures, flags ...string) staking.Pool {
	cmd := fmt.Sprintf("%s query staking pool %v", f.SimcliBinary, f.Flags())
	out, _ := tests.ExecuteT(f.T, helpers.AddFlags(cmd, flags), "")
	var pool staking.Pool

	err := f.Cdc.UnmarshalJSON([]byte(out), &pool)
	require.NoError(f.T, err, "out %v\n, err %v", out, err)
	return pool
}

// QueryStakingParameters is simcli query staking parameters
func QueryStakingParameters(f *helpers.Fixtures, flags ...string) staking.Params {
	cmd := fmt.Sprintf("%s query staking params %v", f.SimcliBinary, f.Flags())
	out, _ := tests.ExecuteT(f.T, helpers.AddFlags(cmd, flags), "")
	var params staking.Params

	err := f.Cdc.UnmarshalJSON([]byte(out), &params)
	require.NoError(f.T, err, "out %v\n, err %v", out, err)
	return params
}
