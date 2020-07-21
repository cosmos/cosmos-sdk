package testutil

import (
	"fmt"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/tests"
	"github.com/cosmos/cosmos-sdk/tests/cli"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/mint/types"
)

// QueryMintingParams returns the current minting parameters
func QueryMintingParams(f *cli.Fixtures, flags ...string) types.Params {
	cmd := fmt.Sprintf("%s query mint params %v", f.SimdBinary, f.Flags())
	out, errStr := tests.ExecuteT(f.T, cli.AddFlags(cmd, flags), "")
	require.Empty(f.T, errStr)

	var params types.Params
	require.NoError(f.T, f.Cdc.UnmarshalJSON([]byte(out), &params))
	return params
}

// QueryInflation returns the current minting inflation value
func QueryInflation(f *cli.Fixtures, flags ...string) sdk.Dec {
	cmd := fmt.Sprintf("%s query mint inflation %v", f.SimdBinary, f.Flags())
	out, errStr := tests.ExecuteT(f.T, cli.AddFlags(cmd, flags), "")
	require.Empty(f.T, errStr)

	var inflation sdk.Dec
	require.NoError(f.T, f.Cdc.UnmarshalJSON([]byte(out), &inflation))
	return inflation
}

// QueryAnnualProvisions returns the current minting annual provisions value
func QueryAnnualProvisions(f *cli.Fixtures, flags ...string) sdk.Dec {
	cmd := fmt.Sprintf("%s query mint annual-provisions %v", f.SimdBinary, f.Flags())
	out, errStr := tests.ExecuteT(f.T, cli.AddFlags(cmd, flags), "")
	require.Empty(f.T, errStr)

	var annualProvisions sdk.Dec
	require.NoError(f.T, f.Cdc.UnmarshalJSON([]byte(out), &annualProvisions))
	return annualProvisions
}
