package cli_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/testutil"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	"github.com/cosmos/cosmos-sdk/types/module"
	testutilmod "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	"github.com/cosmos/cosmos-sdk/x/genutil/client/cli"
	"github.com/cosmos/cosmos-sdk/x/staking"
)

// An example exported genesis file from a 0.37 chain. Note that evidence
// parameters only contains `max_age`.
var v037Exported = `{
	"app_hash": "",
	"app_state": {},
	"chain_id": "test",
	"consensus_params": {
		"block": {
		"max_bytes": "22020096",
		"max_gas": "-1",
		"time_iota_ms": "1000"
		},
		"evidence": { "max_age": "100000" },
		"validator": { "pub_key_types": ["ed25519"] }
	},
	"genesis_time": "2020-09-29T20:16:29.172362037Z",
	"validators": []
}`

func TestValidateGenesis(t *testing.T) {
	cdc := testutilmod.MakeTestEncodingConfig(genutil.AppModule{}).Codec
	testCases := []struct {
		name      string
		genesis   string
		expErrStr string
		genMM     module.BasicManager
	}{
		{
			"invalid json",
			`{"app_state": {x,}}`,
			"error at offset 16: invalid character",
			module.NewBasicManager(),
		},
		{
			"invalid: missing module config in app_state",
			func() string {
				bz, err := os.ReadFile("../../types/testdata/app_genesis.json")
				require.NoError(t, err)

				return string(bz)
			}(),
			"section is missing in the app_state",
			module.NewBasicManager(staking.NewAppModule(cdc, nil, nil, nil, nil)),
		},
		{
			"exported 0.37 genesis file",
			v037Exported,
			"make sure that you have correctly migrated all CometBFT consensus params",
			module.NewBasicManager(),
		},
		{
			"valid 0.50 genesis file",
			func() string {
				bz, err := os.ReadFile("../../types/testdata/app_genesis.json")
				require.NoError(t, err)

				return string(bz)
			}(),
			"",
			module.NewBasicManager(),
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			genesisFile := testutil.WriteToNewTempFile(t, tc.genesis)
			_, err := clitestutil.ExecTestCLICmd(client.Context{}, cli.ValidateGenesisCmd(tc.genMM), []string{genesisFile.Name()})
			if tc.expErrStr != "" {
				require.Contains(t, err.Error(), tc.expErrStr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
