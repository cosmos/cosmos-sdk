package v045_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/simapp"
	v040gov "github.com/cosmos/cosmos-sdk/x/gov/migrations/v040"
	v045gov "github.com/cosmos/cosmos-sdk/x/gov/migrations/v045"
)

func TestMigrate(t *testing.T) {
	encodingConfig := simapp.MakeTestEncodingConfig()
	clientCtx := client.Context{}.
		WithInterfaceRegistry(encodingConfig.InterfaceRegistry).
		WithTxConfig(encodingConfig.TxConfig).
		WithLegacyAmino(encodingConfig.Amino).
		WithCodec(encodingConfig.Codec)

	govGenState := v040gov.GenesisState{
		Proposals: []v040gov.Proposal{
			{
				ProposalId: 5,
				Content:    v040gov.NewTextProposal("foo_title", "bar_description"),
			},
			// TODO: add the other proposal types and assert that they migrate correctly
			// {
			// 	Content: v036distr.CommunityPoolSpendProposal{
			// 		Title:       "foo_community",
			// 		Description: "bar_community",
			// 		Recipient:   recipient,
			// 		Amount:      sdk.NewCoins(sdk.NewCoin("footoken", sdk.NewInt(2))),
			// 	},
			// },
			// {
			// 	Content: v038upgrade.CancelSoftwareUpgradeProposal{
			// 		Title:       "foo_cancel_upgrade",
			// 		Description: "bar_cancel_upgrade",
			// 	},
			// },
			// {
			// 	Content: v038upgrade.SoftwareUpgradeProposal{
			// 		Title:       "foo_software_upgrade",
			// 		Description: "bar_software_upgrade",
			// 		Plan: v038upgrade.Plan{
			// 			Name:   "foo_upgrade_name",
			// 			Height: 123,
			// 			Info:   "foo_upgrade_info",
			// 		},
			// 	},
			// },
			// {
			// 	Content: v036params.ParameterChangeProposal{
			// 		Title:       "foo_param_change",
			// 		Description: "bar_param_change",
			// 		Changes: []v036params.ParamChange{
			// 			{
			// 				Subspace: "foo_param_change_subspace",
			// 				Key:      "foo_param_change_key",
			// 				Subkey:   "foo_param_change_subkey",
			// 				Value:    "foo_param_change_value",
			// 			},
			// 		},
			// 	},
			// },
		},
	}

	migrated := v045gov.Migrate(govGenState)

	bz, err := clientCtx.Codec.MarshalJSON(migrated)
	require.NoError(t, err)

	// Indent the JSON bz correctly.
	var jsonObj map[string]interface{}
	err = json.Unmarshal(bz, &jsonObj)
	require.NoError(t, err)
	indentedBz, err := json.MarshalIndent(jsonObj, "", "\t")
	require.NoError(t, err)

	// Make sure about:
	// - SignalProposal has correct JSON.
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
