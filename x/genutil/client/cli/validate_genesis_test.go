package cli_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/testutil"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	"github.com/cosmos/cosmos-sdk/x/genutil/client/cli"
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
	testCases := []struct {
		name    string
		genesis string
		expErr  bool
	}{
		{
			"exported 0.37 genesis file",
			v037Exported,
			true,
		},
		{
			"valid 0.50 genesis file",
			func() string {
				bz, err := os.ReadFile("../../types/testdata/app_genesis.json")
				require.NoError(t, err)

				return string(bz)
			}(),
			false,
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			genesisFile := testutil.WriteToNewTempFile(t, tc.genesis)
			_, err := clitestutil.ExecTestCLICmd(client.Context{}, cli.ValidateGenesisCmd(nil), []string{genesisFile.Name()})
			if tc.expErr {
				require.Contains(t, err.Error(), "make sure that you have correctly migrated all CometBFT consensus params")
			} else {
				require.NoError(t, err)
			}
		})
	}
}
