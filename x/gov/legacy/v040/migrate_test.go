package v040_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	v036distr "github.com/cosmos/cosmos-sdk/x/distribution/legacy/v036"
	v036gov "github.com/cosmos/cosmos-sdk/x/gov/legacy/v036"
	v040gov "github.com/cosmos/cosmos-sdk/x/gov/legacy/v040"
	v036params "github.com/cosmos/cosmos-sdk/x/params/legacy/v036"
	v038upgrade "github.com/cosmos/cosmos-sdk/x/upgrade/legacy/v038"
)

func TestMigrate(t *testing.T) {
	encodingConfig := simapp.MakeTestEncodingConfig()
	clientCtx := client.Context{}.
		WithInterfaceRegistry(encodingConfig.InterfaceRegistry).
		WithTxConfig(encodingConfig.TxConfig).
		WithLegacyAmino(encodingConfig.Amino).
		WithJSONMarshaler(encodingConfig.Marshaler)

	recipient, err := sdk.AccAddressFromBech32("cosmos1fl48vsnmsdzcv85q5d2q4z5ajdha8yu34mf0eh")
	require.NoError(t, err)
	govGenState := v036gov.GenesisState{
		Proposals: []v036gov.Proposal{
			{
				Content: v036gov.TextProposal{
					Title:       "foo_text",
					Description: "bar_text",
				},
			},
			{
				Content: v036distr.CommunityPoolSpendProposal{
					Title:       "foo_community",
					Description: "bar_community",
					Recipient:   recipient,
					Amount:      sdk.NewCoins(sdk.NewCoin("footoken", sdk.NewInt(2))),
				},
			},
			{
				Content: v038upgrade.CancelSoftwareUpgradeProposal{
					Title:       "foo_cancel_upgrade",
					Description: "bar_cancel_upgrade",
				},
			},
			{
				Content: v038upgrade.SoftwareUpgradeProposal{
					Title:       "foo_software_upgrade",
					Description: "bar_software_upgrade",
					Plan: v038upgrade.Plan{
						Name:   "foo_upgrade_name",
						Height: 123,
						Info:   "foo_upgrade_info",
					},
				},
			},
			{
				Content: v036params.ParameterChangeProposal{
					Title:       "foo_param_change",
					Description: "bar_param_change",
					Changes: []v036params.ParamChange{
						{
							Subspace: "foo_param_change_subspace",
							Key:      "foo_param_change_key",
							Subkey:   "foo_param_change_subkey",
							Value:    "foo_param_change_value",
						},
					},
				},
			},
		},
	}

	migrated := v040gov.Migrate(govGenState)

	bz, err := clientCtx.JSONMarshaler.MarshalJSON(migrated)
	require.NoError(t, err)

	// Indent the JSON bz correctly.
	var jsonObj map[string]interface{}
	err = json.Unmarshal(bz, &jsonObj)
	require.NoError(t, err)
	indentedBz, err := json.MarshalIndent(jsonObj, "", "\t")
	require.NoError(t, err)

	// Make sure about:
	// - TextProposal has correct JSON.
	// - CommunityPoolSpendProposal has correct JSON.
	// - CancelSoftwareUpgradeProposal has correct JSON.
	// - SoftwareUpgradeProposal has correct JSON.
	// - ParameterChangeProposal has correct JSON.
	expected := `{
	"deposit_params": {
		"max_deposit_period": "0s",
		"min_deposit": []
	},
	"deposits": [],
	"proposals": [
		{
			"content": {
				"@type": "/cosmos.gov.v1beta1.TextProposal",
				"description": "bar_text",
				"title": "foo_text"
			},
			"deposit_end_time": "0001-01-01T00:00:00Z",
			"final_tally_result": {
				"abstain": "0",
				"no": "0",
				"no_with_veto": "0",
				"yes": "0"
			},
			"proposal_id": "0",
			"status": "PROPOSAL_STATUS_UNSPECIFIED",
			"submit_time": "0001-01-01T00:00:00Z",
			"total_deposit": [],
			"voting_end_time": "0001-01-01T00:00:00Z",
			"voting_start_time": "0001-01-01T00:00:00Z"
		},
		{
			"content": {
				"@type": "/cosmos.distribution.v1beta1.CommunityPoolSpendProposal",
				"amount": [
					{
						"amount": "2",
						"denom": "footoken"
					}
				],
				"description": "bar_community",
				"recipient": "cosmos1fl48vsnmsdzcv85q5d2q4z5ajdha8yu34mf0eh",
				"title": "foo_community"
			},
			"deposit_end_time": "0001-01-01T00:00:00Z",
			"final_tally_result": {
				"abstain": "0",
				"no": "0",
				"no_with_veto": "0",
				"yes": "0"
			},
			"proposal_id": "0",
			"status": "PROPOSAL_STATUS_UNSPECIFIED",
			"submit_time": "0001-01-01T00:00:00Z",
			"total_deposit": [],
			"voting_end_time": "0001-01-01T00:00:00Z",
			"voting_start_time": "0001-01-01T00:00:00Z"
		},
		{
			"content": {
				"@type": "/cosmos.upgrade.v1beta1.CancelSoftwareUpgradeProposal",
				"description": "bar_cancel_upgrade",
				"title": "foo_cancel_upgrade"
			},
			"deposit_end_time": "0001-01-01T00:00:00Z",
			"final_tally_result": {
				"abstain": "0",
				"no": "0",
				"no_with_veto": "0",
				"yes": "0"
			},
			"proposal_id": "0",
			"status": "PROPOSAL_STATUS_UNSPECIFIED",
			"submit_time": "0001-01-01T00:00:00Z",
			"total_deposit": [],
			"voting_end_time": "0001-01-01T00:00:00Z",
			"voting_start_time": "0001-01-01T00:00:00Z"
		},
		{
			"content": {
				"@type": "/cosmos.upgrade.v1beta1.SoftwareUpgradeProposal",
				"description": "bar_software_upgrade",
				"plan": {
					"height": "123",
					"info": "foo_upgrade_info",
					"name": "foo_upgrade_name",
					"time": "0001-01-01T00:00:00Z",
					"upgraded_client_state": null
				},
				"title": "foo_software_upgrade"
			},
			"deposit_end_time": "0001-01-01T00:00:00Z",
			"final_tally_result": {
				"abstain": "0",
				"no": "0",
				"no_with_veto": "0",
				"yes": "0"
			},
			"proposal_id": "0",
			"status": "PROPOSAL_STATUS_UNSPECIFIED",
			"submit_time": "0001-01-01T00:00:00Z",
			"total_deposit": [],
			"voting_end_time": "0001-01-01T00:00:00Z",
			"voting_start_time": "0001-01-01T00:00:00Z"
		},
		{
			"content": {
				"@type": "/cosmos.params.v1beta1.ParameterChangeProposal",
				"changes": [
					{
						"key": "foo_param_change_key",
						"subspace": "foo_param_change_subspace",
						"value": "foo_param_change_value"
					}
				],
				"description": "bar_param_change",
				"title": "foo_param_change"
			},
			"deposit_end_time": "0001-01-01T00:00:00Z",
			"final_tally_result": {
				"abstain": "0",
				"no": "0",
				"no_with_veto": "0",
				"yes": "0"
			},
			"proposal_id": "0",
			"status": "PROPOSAL_STATUS_UNSPECIFIED",
			"submit_time": "0001-01-01T00:00:00Z",
			"total_deposit": [],
			"voting_end_time": "0001-01-01T00:00:00Z",
			"voting_start_time": "0001-01-01T00:00:00Z"
		}
	],
	"starting_proposal_id": "0",
	"tally_params": {
		"quorum": "0",
		"threshold": "0",
		"veto_threshold": "0"
	},
	"votes": [],
	"voting_params": {
		"voting_period": "0s"
	}
}`

	require.Equal(t, expected, string(indentedBz))
}
