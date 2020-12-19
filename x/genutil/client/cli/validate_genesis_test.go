package cli_test

import (
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

// An example exported genesis file that's 0.40 compatible.
var v040Valid = `{
	"app_hash": "",
	"app_state": {},
	"chain_id": "test",
	"consensus_params": {
		"block": {
		"max_bytes": "22020096",
		"max_gas": "-1",
		"time_iota_ms": "1000"
		},
		"evidence": {
			"max_age_num_blocks": "100000",
			"max_age_duration": "172800000000000",
			"max_bytes": "0"
		},
		"validator": { "pub_key_types": ["ed25519"] }
	},
	"genesis_time": "2020-09-29T20:16:29.172362037Z",
	"validators": []
}`

func (s *IntegrationTestSuite) TestValidateGenesis() {
	val0 := s.network.Validators[0]

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
			"valid 0.40 genesis file",
			v040Valid,
			false,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			genesisFile := testutil.WriteToNewTempFile(s.T(), tc.genesis)
			_, err := clitestutil.ExecTestCLICmd(val0.ClientCtx, cli.ValidateGenesisCmd(nil), []string{genesisFile.Name()})
			if tc.expErr {
				s.Require().Contains(err.Error(), "Make sure that you have correctly migrated all Tendermint consensus params")

			} else {
				s.Require().NoError(err)
			}
		})
	}
}
