package cli

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/tests"
	"github.com/cosmos/cosmos-sdk/tests/cli/helpers"
	"github.com/cosmos/cosmos-sdk/x/slashing"
	"github.com/stretchr/testify/require"
)

// QuerySigningInfo returns the signing info for a validator
func QuerySigningInfo(f *helpers.Fixtures, val string) slashing.ValidatorSigningInfo {
	cmd := fmt.Sprintf("%s query slashing signing-info %s %s", f.SimcliBinary, val, f.Flags())
	res, errStr := tests.ExecuteT(f.T, cmd, "")
	require.Empty(f.T, errStr)

	var sinfo slashing.ValidatorSigningInfo
	err := f.Cdc.UnmarshalJSON([]byte(res), &sinfo)
	require.NoError(f.T, err)
	return sinfo
}

// QuerySlashingParams is gaiacli query slashing params
func QuerySlashingParams(f *helpers.Fixtures) slashing.Params {
	cmd := fmt.Sprintf("%s query slashing params %s", f.SimcliBinary, f.Flags())
	res, errStr := tests.ExecuteT(f.T, cmd, "")
	require.Empty(f.T, errStr)

	var params slashing.Params
	err := f.Cdc.UnmarshalJSON([]byte(res), &params)
	require.NoError(f.T, err)
	return params
}
