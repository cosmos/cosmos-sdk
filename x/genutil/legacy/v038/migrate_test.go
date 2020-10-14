package v038_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/client"
	v036auth "github.com/cosmos/cosmos-sdk/x/auth/legacy/v036"
	v036genaccounts "github.com/cosmos/cosmos-sdk/x/genaccounts/legacy/v036"
	v038 "github.com/cosmos/cosmos-sdk/x/genutil/legacy/v038"
	"github.com/cosmos/cosmos-sdk/x/genutil/types"
	v036staking "github.com/cosmos/cosmos-sdk/x/staking/legacy/v036"

	"github.com/stretchr/testify/require"
)

var genAccountsState = []byte(`[
		{
			"account_number": "0",
			"address": "cosmos1q7380u26f7ntke3facjmynajs4umlr329vr4ja",
			"coins": [
				{
					"amount": "1000000000",
					"denom": "node0token"
				},
				{
					"amount": "400000198",
					"denom": "stake"
				}
			],
			"delegated_free": [],
			"delegated_vesting": [],
			"end_time": "0",
			"module_name": "",
			"module_permissions": [],
			"original_vesting": [],
			"sequence_number": "1",
			"start_time": "0"
		},
		{
			"account_number": "0",
			"address": "cosmos1tygms3xhhs3yv487phx3dw4a95jn7t7lpm470r",
			"coins": [],
			"delegated_free": [],
			"delegated_vesting": [],
			"end_time": "0",
			"module_name": "not_bonded_tokens_pool",
			"module_permissions": [
				"burner",
				"staking"
			],
			"original_vesting": [],
			"sequence_number": "0",
			"start_time": "0"
		},
		{
			"account_number": "0",
			"address": "cosmos1m3h30wlvsf8llruxtpukdvsy0km2kum8g38c8q",
			"coins": [],
			"delegated_free": [],
			"delegated_vesting": [],
			"end_time": "0",
			"module_name": "mint",
			"module_permissions": [
				"minter"
			],
			"original_vesting": [],
			"sequence_number": "0",
			"start_time": "0"
		}
	]`)

var genAuthState = []byte(`{
  "params": {
    "max_memo_characters": "256",
    "sig_verify_cost_ed25519": "590",
    "sig_verify_cost_secp256k1": "1000",
    "tx_sig_limit": "7",
    "tx_size_cost_per_byte": "10"
  }
}`)

var genStakingState = []byte(`{
  "delegations": [
    {
      "delegator_address": "cosmos1q7380u26f7ntke3facjmynajs4umlr329vr4ja",
      "shares": "100000000.000000000000000000",
      "validator_address": "cosmosvaloper1q7380u26f7ntke3facjmynajs4umlr32qchq7w"
    }
  ],
  "exported": true,
  "last_total_power": "400",
  "last_validator_powers": [
    {
      "Address": "cosmosvaloper1q7380u26f7ntke3facjmynajs4umlr32qchq7w",
      "Power": "100"
    }
  ],
  "params": {
    "bond_denom": "stake",
    "max_entries": 7,
    "max_validators": 100,
    "unbonding_time": "259200000000000"
  },
  "redelegations": null,
  "unbonding_delegations": null,
  "validators": [
    {
      "commission": {
        "commission_rates": {
          "max_change_rate": "0.000000000000000000",
          "max_rate": "0.000000000000000000",
          "rate": "0.000000000000000000"
        },
        "update_time": "2019-09-24T23:11:22.9692177Z"
      },
      "consensus_pubkey": "cosmosvalconspub1zcjduepqygqrt0saxf76lhsmp56rx52j0acdxyjvcdkq3tqvwrsmmm0ke28q36kh9h",
      "delegator_shares": "100000000.000000000000000000",
      "description": {
        "details": "",
        "identity": "",
        "moniker": "node0",
        "website": ""
      },
      "jailed": false,
      "min_self_delegation": "1",
      "operator_address": "cosmosvaloper1q7380u26f7ntke3facjmynajs4umlr32qchq7w",
      "status": 2,
      "tokens": "100000000",
      "unbonding_height": "0",
      "unbonding_time": "1970-01-01T00:00:00Z"
    }
  ]
}`)

func TestMigrate(t *testing.T) {
	genesis := types.AppMap{
		v036auth.ModuleName:        genAuthState,
		v036genaccounts.ModuleName: genAccountsState,
		v036staking.ModuleName:     genStakingState,
	}

	require.NotPanics(t, func() { v038.Migrate(genesis, client.Context{}) })
}
