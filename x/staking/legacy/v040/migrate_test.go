package v040_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/simapp"
	v034staking "github.com/cosmos/cosmos-sdk/x/staking/legacy/v034"
	v038staking "github.com/cosmos/cosmos-sdk/x/staking/legacy/v038"
	v040staking "github.com/cosmos/cosmos-sdk/x/staking/legacy/v040"
)

func TestMigrate(t *testing.T) {
	encodingConfig := simapp.MakeTestEncodingConfig()
	clientCtx := client.Context{}.
		WithInterfaceRegistry(encodingConfig.InterfaceRegistry).
		WithTxConfig(encodingConfig.TxConfig).
		WithLegacyAmino(encodingConfig.Amino).
		WithJSONCodec(encodingConfig.Marshaler)

	consPubKey := ed25519.GenPrivKeyFromSecret([]byte("val0")).PubKey()
	stakingGenState := v038staking.GenesisState{
		Validators: v038staking.Validators{v038staking.Validator{
			ConsPubKey: consPubKey,
			Status:     v034staking.Unbonded,
		}},
	}

	migrated := v040staking.Migrate(stakingGenState)

	bz, err := clientCtx.JSONCodec.MarshalJSON(migrated)
	require.NoError(t, err)

	// Indent the JSON bz correctly.
	var jsonObj map[string]interface{}
	err = json.Unmarshal(bz, &jsonObj)
	require.NoError(t, err)
	indentedBz, err := json.MarshalIndent(jsonObj, "", "  ")
	require.NoError(t, err)

	// Make sure about:
	// - consensus_pubkey: should be an any
	// - validator's status should be 1 (new unbonded)
	expected := `{
  "delegations": [],
  "exported": false,
  "last_total_power": "0",
  "last_validator_powers": [],
  "params": {
    "bond_denom": "",
    "historical_entries": 0,
    "max_entries": 0,
    "max_validators": 0,
    "unbonding_time": "0s"
  },
  "redelegations": [],
  "unbonding_delegations": [],
  "validators": [
    {
      "commission": {
        "commission_rates": {
          "max_change_rate": "0",
          "max_rate": "0",
          "rate": "0"
        },
        "update_time": "0001-01-01T00:00:00Z"
      },
      "consensus_pubkey": {
        "@type": "/cosmos.crypto.ed25519.PubKey",
        "key": "KTeVrjP7NJIufvgMJsQRxZjfFyD+Exda6O7x+oxIvmA="
      },
      "delegator_shares": "0",
      "description": {
        "details": "",
        "identity": "",
        "moniker": "",
        "security_contact": "",
        "website": ""
      },
      "jailed": false,
      "min_self_delegation": "0",
      "operator_address": "",
      "status": "BOND_STATUS_UNBONDED",
      "tokens": "0",
      "unbonding_height": "0",
      "unbonding_time": "0001-01-01T00:00:00Z"
    }
  ]
}`

	require.Equal(t, expected, string(indentedBz))
}
