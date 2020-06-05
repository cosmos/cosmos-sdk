package testutil

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/tests"
	"github.com/cosmos/cosmos-sdk/tests/cli"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/stretchr/testify/require"
)

// QueryMintingParams returns the current minting parameters
func QueryMintingParams(f *cli.Fixtures, flags ...string) types.Params {
	cmd := fmt.Sprintf("%s query mint params %v", f.SimcliBinary, f.Flags())
	out, errStr := tests.ExecuteT(f.T, cli.AddFlags(cmd, flags), "")
	require.Empty(f.T, errStr)

	var params types.Params
	err := f.Cdc.UnmarshalJSON([]byte(out), &params)
	require.NoError(f.T, err, "out %v\n, err %v", out, err)
	return params
}

// QueryInflation returns the current minting inflation value
func QueryInflation(f *cli.Fixtures, flags ...string) sdk.Dec {
	cmd := fmt.Sprintf("%s query mint inflation %v", f.SimcliBinary, f.Flags())
	out, errStr := tests.ExecuteT(f.T, cli.AddFlags(cmd, flags), "")
	require.Empty(f.T, errStr)

	var inflation sdk.Dec
	err := f.Cdc.UnmarshalJSON([]byte(out), &inflation)
	require.NoError(f.T, err, "out %v\n, err %v", out, err)
	return inflation
}

// QueryAnnualProvisions returns the current minting annual provisions value
func QueryAnnualProvisions(f *cli.Fixtures, flags ...string) sdk.Dec {
	cmd := fmt.Sprintf("%s query mint annual-provisions %v", f.SimcliBinary, f.Flags())
	out, errStr := tests.ExecuteT(f.T, cli.AddFlags(cmd, flags), "")
	require.Empty(f.T, errStr)

	var annualProvisions sdk.Dec
	err := f.Cdc.UnmarshalJSON([]byte(out), &annualProvisions)
	require.NoError(f.T, err, "out1 %v\n, err1 %v", out, err)
	return annualProvisions
}
