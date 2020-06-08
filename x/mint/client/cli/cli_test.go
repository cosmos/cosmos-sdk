// +build cli_test

package cli_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/tests/cli"
	"github.com/cosmos/cosmos-sdk/x/mint/client/testutil"
	"github.com/stretchr/testify/require"
)

func TestCLIMintQueries(t *testing.T) {
	t.Parallel()
	f := cli.InitFixtures(t)

	proc := f.SDStart()
	t.Cleanup(func() { proc.Stop(false) })

	params := testutil.QueryMintingParams(f)
	require.NotEmpty(t, params)

	inflation := testutil.QueryInflation(f)
	require.False(t, inflation.IsZero())

	annualProvisions := testutil.QueryAnnualProvisions(f)
	require.False(t, annualProvisions.IsZero())
}
