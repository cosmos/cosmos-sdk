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

func (s *IntegrationTestSuite) TestValidateGenesis_FromV037() {
	val0 := s.network.Validators[0]

	genesisFile := testutil.WriteToNewTempFile(s.T(), v037Exported)
	// We expect an error decoding an older `consensus_params` with the latest
	// TM validation.
	_, err := clitestutil.ExecTestCLICmd(val0.ClientCtx, cli.ValidateGenesisCmd(nil), []string{genesisFile.Name()})
	s.Require().Contains(err.Error(), "Make sure that you have correctly migrated all Tendermint consensus params")
}
