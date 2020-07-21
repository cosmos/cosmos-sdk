package testutil

import (
	"fmt"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/tests"
	"github.com/cosmos/cosmos-sdk/tests/cli"
	"github.com/cosmos/cosmos-sdk/x/slashing/types"
)

// QuerySigningInfo returns the signing info for a validator
func QuerySigningInfo(f *cli.Fixtures, val string) types.ValidatorSigningInfo {
	cmd := fmt.Sprintf("%s query slashing signing-info %s %s", f.SimdBinary, val, f.Flags())
	res, errStr := tests.ExecuteT(f.T, cmd, "")
	require.Empty(f.T, errStr)

	var sinfo types.ValidatorSigningInfo
	err := f.Cdc.UnmarshalJSON([]byte(res), &sinfo)
	require.NoError(f.T, err)
	return sinfo
}

// QuerySlashingParams returns query slashing params
func QuerySlashingParams(f *cli.Fixtures) types.Params {
	cmd := fmt.Sprintf("%s query slashing params %s", f.SimdBinary, f.Flags())
	res, errStr := tests.ExecuteT(f.T, cmd, "")
	require.Empty(f.T, errStr)

	var params types.Params
	err := f.Cdc.UnmarshalJSON([]byte(res), &params)
	require.NoError(f.T, err)
	return params
}
