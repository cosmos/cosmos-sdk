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
)

func TestMigrate(t *testing.T) {
	encodingConfig := simapp.MakeEncodingConfig()
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
					Title:       "foo_test",
					Description: "bar_test",
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
		},
	}

	migrated := v040gov.Migrate(govGenState)

	bz, err := clientCtx.JSONMarshaler.MarshalJSON(migrated)
	require.NoError(t, err)

	// Indent the JSON bz correctly.
	var jsonObj map[string]interface{}
	err = json.Unmarshal(bz, &jsonObj)
	require.NoError(t, err)
	indentedBz, err := json.MarshalIndent(jsonObj, "", "  ")
	require.NoError(t, err)

	// Make sure about:
	// - TextProposal and CommunityPoolSpendProposal have correct any JSON.
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
        "description": "bar_test",
        "title": "foo_test"
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
