package v038

import (
	"encoding/json"
	"testing"

	v034auth "github.com/cosmos/cosmos-sdk/x/auth/legacy/v0_34"
	v036auth "github.com/cosmos/cosmos-sdk/x/auth/legacy/v0_36"

	"github.com/stretchr/testify/require"
)

func TestMigrate(t *testing.T) {
	var genesisState GenesisState

	params := v034auth.Params{
		MaxMemoCharacters:      10,
		TxSigLimit:             10,
		TxSizeCostPerByte:      10,
		SigVerifyCostED25519:   10,
		SigVerifyCostSecp256k1: 10,
	}
	rawAccounts := `[
		{
			"address": "cosmos1dfp05pasnts7a4lupn889vptjtrxzkk5f7027f",
			"coins": [
				{
					"denom": "node0token",
					"amount": "1000000000"
				},
				{
					"denom": "stake",
					"amount": "500000000"
				}
			],
			"sequence_number": "0",
			"account_number": "0",
			"original_vesting": [],
			"delegated_free": [],
			"delegated_vesting": [],
			"start_time": "0",
			"end_time": "0",
			"module_name": "",
			"module_permissions": null
		},
		{
			"address": "cosmos1f6dangl9ggdhuvkcwhswserr8fzra6vfzfjvh2",
			"coins": [
				{
					"denom": "node1token",
					"amount": "1000000000"
				},
				{
					"denom": "stake",
					"amount": "500000000"
				}
			],
			"sequence_number": "0",
			"account_number": "0",
			"original_vesting": [],
			"delegated_free": [],
			"delegated_vesting": [],
			"start_time": "0",
			"end_time": "0",
			"module_name": "",
			"module_permissions": null
		},
		{
			"address": "cosmos1gudmxhn5anh5m6m2rr4rsfhgvps8fchtgmk7a6",
			"coins": [
				{
					"denom": "node2token",
					"amount": "1000000000"
				},
				{
					"denom": "stake",
					"amount": "500000000"
				}
			],
			"sequence_number": "0",
			"account_number": "0",
			"original_vesting": [],
			"delegated_free": [],
			"delegated_vesting": [],
			"start_time": "0",
			"end_time": "0",
			"module_name": "",
			"module_permissions": null
		},
		{
			"address": "cosmos1kluvs8ff2s3hxad4jpmhvca4crqpcwn9xyhchv",
			"coins": [
				{
					"denom": "node3token",
					"amount": "1000000000"
				},
				{
					"denom": "stake",
					"amount": "500000000"
				}
			],
			"sequence_number": "0",
			"account_number": "0",
			"original_vesting": [],
			"delegated_free": [],
			"delegated_vesting": [],
			"start_time": "0",
			"end_time": "0",
			"module_name": "",
			"module_permissions": null
		}
	]`

	require.NotPanics(t, func() {
		genesisState = Migrate(
			v036auth.GenesisState{
				Params: params,
			},
			json.RawMessage(rawAccounts),
		)
	})

	require.Equal(t, genesisState, GenesisState{Params: params, Accounts: json.RawMessage(rawAccounts)})
}
