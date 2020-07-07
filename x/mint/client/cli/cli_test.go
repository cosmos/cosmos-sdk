// +build cli_test

package cli_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/tests/cli"
	"github.com/cosmos/cosmos-sdk/x/mint/client/testutil"
)

func TestCLIMintQueries(t *testing.T) {
	t.SkipNow() // TODO: Bring back once viper is refactored.
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
