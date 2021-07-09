package v043_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	v043bank "github.com/cosmos/cosmos-sdk/x/bank/migrations/v043"
	"github.com/cosmos/cosmos-sdk/x/bank/types"
)

func TestMigrateJSON(t *testing.T) {
	encodingConfig := simapp.MakeTestEncodingConfig()
	clientCtx := client.Context{}.
		WithInterfaceRegistry(encodingConfig.InterfaceRegistry).
		WithTxConfig(encodingConfig.TxConfig).
		WithCodec(encodingConfig.Codec)

	voter, err := sdk.AccAddressFromBech32("cosmos1fl48vsnmsdzcv85q5d2q4z5ajdha8yu34mf0eh")
	require.NoError(t, err)
	bankGenState := &types.GenesisState{
		Balances: []types.Balance{
			{
				Address: voter.String(),
				Coins: sdk.Coins{
					sdk.NewCoin("foo", sdk.NewInt(10)),
					sdk.NewCoin("bar", sdk.NewInt(20)),
					sdk.NewCoin("foobar", sdk.NewInt(0)),
				},
			},
		},
		Supply: sdk.Coins{
			sdk.NewCoin("foo", sdk.NewInt(10)),
			sdk.NewCoin("bar", sdk.NewInt(20)),
			sdk.NewCoin("foobar", sdk.NewInt(0)),
			sdk.NewCoin("barfoo", sdk.NewInt(0)),
		},
	}

	migrated := v043bank.MigrateJSON(bankGenState)

	bz, err := clientCtx.Codec.MarshalJSON(migrated)
	require.NoError(t, err)

	// Indent the JSON bz correctly.
	var jsonObj map[string]interface{}
	err = json.Unmarshal(bz, &jsonObj)
	require.NoError(t, err)
	indentedBz, err := json.MarshalIndent(jsonObj, "", "\t")
	require.NoError(t, err)

	// Make sure about:
	// - zero coin balances pruned.
	// - zero supply denoms pruned.
	expected := `{
	"balances": [
		{
			"address": "cosmos1fl48vsnmsdzcv85q5d2q4z5ajdha8yu34mf0eh",
			"coins": [
				{
					"amount": "20",
					"denom": "bar"
				},
				{
					"amount": "10",
					"denom": "foo"
				}
			]
		}
	],
	"denom_metadata": [],
	"params": {
		"default_send_enabled": false,
		"send_enabled": []
	},
	"supply": [
		{
			"amount": "20",
			"denom": "bar"
		},
		{
			"amount": "10",
			"denom": "foo"
		}
	]
}`

	require.Equal(t, expected, string(indentedBz))
}
