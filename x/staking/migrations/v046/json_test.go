package v046_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/depinject"
	v046 "github.com/cosmos/cosmos-sdk/x/staking/migrations/v046"
	"github.com/cosmos/cosmos-sdk/x/staking/testutil"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

func TestMigrateJSON(t *testing.T) {
	var (
		cdc               codec.Codec
		txConfig          client.TxConfig
		interfaceRegistry codectypes.InterfaceRegistry
	)

	err := depinject.Inject(testutil.AppConfig,
		&cdc,
		&txConfig,
	)
	require.NoError(t, err)

	clientCtx := client.Context{}.
		WithInterfaceRegistry(interfaceRegistry).
		WithTxConfig(txConfig).
		WithCodec(cdc)

	oldState := types.DefaultGenesisState()

	newState, err := v046.MigrateJSON(*oldState)
	require.NoError(t, err)

	bz, err := clientCtx.Codec.MarshalJSON(&newState)
	require.NoError(t, err)

	// Indent the JSON bz correctly.
	var jsonObj map[string]interface{}
	err = json.Unmarshal(bz, &jsonObj)
	require.NoError(t, err)
	indentedBz, err := json.MarshalIndent(jsonObj, "", "\t")
	require.NoError(t, err)

	// Make sure about new param MinCommissionRate.
	expected := `{
	"delegations": [],
	"exported": false,
	"last_total_power": "0",
	"last_validator_powers": [],
	"params": {
		"bond_denom": "stake",
		"historical_entries": 10000,
		"max_entries": 7,
		"max_validators": 100,
		"min_commission_rate": "0.000000000000000000",
		"unbonding_time": "1814400s"
	},
	"redelegations": [],
	"unbonding_delegations": [],
	"validators": []
}`

	require.Equal(t, expected, string(indentedBz))
}
