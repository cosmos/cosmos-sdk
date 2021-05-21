package v043_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	v040staking "github.com/cosmos/cosmos-sdk/x/staking/legacy/v040"
	v043staking "github.com/cosmos/cosmos-sdk/x/staking/legacy/v043"
)

func TestMigrateJSON(t *testing.T) {
	encodingConfig := simapp.MakeTestEncodingConfig()
	clientCtx := client.Context{}.
		WithInterfaceRegistry(encodingConfig.InterfaceRegistry).
		WithTxConfig(encodingConfig.TxConfig).
		WithJSONCodec(encodingConfig.Marshaler)

	// voter, err := sdk.AccAddressFromBech32("cosmos1fl48vsnmsdzcv85q5d2q4z5ajdha8yu34mf0eh")
	// require.NoError(t, err)
	stakingGenState := &v040staking.GenesisState{
		Params: v040staking.DefaultParams(),
	}

	migrated := v043staking.MigrateJSON(stakingGenState)

	require.True(t, migrated.Params.PowerReduction.Equal(sdk.DefaultPowerReduction))

	bz, err := clientCtx.JSONCodec.MarshalJSON(migrated)
	require.NoError(t, err)

	// Indent the JSON bz correctly.
	var jsonObj map[string]interface{}
	err = json.Unmarshal(bz, &jsonObj)
	require.NoError(t, err)
	indentedBz, err := json.MarshalIndent(jsonObj, "", "\t")
	require.NoError(t, err)

	// Make sure about:
	// - Votes are all ADR-037 weighted votes with weight 1.
	expected := `{
	"buffered_msgs": [],
	"delegations": [],
	"epoch_number": "0",
	"exported": false,
	"last_total_power": "0",
	"last_validator_powers": [],
	"params": {
		"bond_denom": "stake",
		"epoch_interval": "10",
		"historical_entries": 10000,
		"max_entries": 7,
		"max_validators": 100,
		"power_reduction": "1000000",
		"unbonding_time": "1814400s"
	},
	"redelegations": [],
	"unbonding_delegations": [],
	"validators": []
}`

	require.Equal(t, expected, string(indentedBz))
}
