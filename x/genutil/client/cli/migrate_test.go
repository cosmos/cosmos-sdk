package cli_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/testutil"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	"github.com/cosmos/cosmos-sdk/x/genutil/client/cli"
)

func TestGetMigrationCallback(t *testing.T) {
	for _, version := range cli.GetMigrationVersions() {
		require.NotNil(t, cli.GetMigrationCallback(version))
	}
}

func (s *IntegrationTestSuite) TestMigrateGenesis() {
	val0 := s.network.Validators[0]

	testCases := []struct {
		name      string
		genesis   string
		target    string
		expErr    bool
		expErrMsg string
	}{
		{
			"migrate to 0.36",
			`{"chain_id":"test","app_state":{}}`,
			"v0.36",
			false, "",
		},
		{
			"exported 0.37 genesis file",
			v037Exported,
			"v0.40",
			true, "Make sure that you have correctly migrated all Tendermint consensus params",
		},
		{
			"valid 0.40 genesis file",
			v040Valid,
			"v0.40",
			false, "",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			genesisFile := testutil.WriteToNewTempFile(s.T(), tc.genesis)
			_, err := clitestutil.ExecTestCLICmd(val0.ClientCtx, cli.MigrateGenesisCmd(), []string{tc.target, genesisFile.Name()})
			if tc.expErr {
				s.Require().Contains(err.Error(), tc.expErrMsg)
			} else {
				s.Require().NoError(err)
			}
		})
	}
}
