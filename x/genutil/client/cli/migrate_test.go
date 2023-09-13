package cli_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/testutil"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/cosmos-sdk/x/genutil/client/cli"
)

func TestMigrateGenesis(t *testing.T) {
	testCases := []struct {
		name      string
		genesis   string
		target    string
		expErr    bool
		expErrMsg string
		check     func(jsonOut string)
	}{
		{
			"migrate 0.37 to 0.43",
			v037Exported,
			"v0.43",
			true, "make sure that you have correctly migrated all CometBFT consensus params", func(_ string) {},
		},
		{
			"invalid target version",
			func() string {
				bz, err := os.ReadFile("../../types/testdata/app_genesis.json")
				require.NoError(t, err)

				return string(bz)
			}(),
			"v0.10",
			true, "unknown migration function for version: v0.10 (supported versions v0.43, v0.46, v0.47)", func(_ string) {},
		},
		{
			"invalid target version",
			func() string {
				bz, err := os.ReadFile("../../types/testdata/cmt_genesis.json")
				require.NoError(t, err)

				return string(bz)
			}(),
			"v0.10",
			true, "unknown migration function for version: v0.10 (supported versions v0.43, v0.46, v0.47)", func(_ string) {},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			genesisFile := testutil.WriteToNewTempFile(t, tc.genesis)
			jsonOutput, err := clitestutil.ExecTestCLICmd(
				// the codec does not contain any modules so that genutil does not bring unnecessary dependencies in the test
				client.Context{Codec: moduletestutil.MakeTestEncodingConfig().Codec},
				cli.MigrateGenesisCmd(cli.MigrationMap),
				[]string{tc.target, genesisFile.Name()},
			)
			if tc.expErr {
				require.Contains(t, err.Error(), tc.expErrMsg)
			} else {
				require.NoError(t, err)
				tc.check(jsonOutput.String())
			}
		})
	}
}
