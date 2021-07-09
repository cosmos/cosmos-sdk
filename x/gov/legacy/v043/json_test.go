package v043_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	v043gov "github.com/cosmos/cosmos-sdk/x/gov/legacy/v043"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
)

func TestMigrateJSON(t *testing.T) {
	encodingConfig := simapp.MakeTestEncodingConfig()
	clientCtx := client.Context{}.
		WithInterfaceRegistry(encodingConfig.InterfaceRegistry).
		WithTxConfig(encodingConfig.TxConfig).
		WithCodec(encodingConfig.Marshaler)

	voter, err := sdk.AccAddressFromBech32("cosmos1fl48vsnmsdzcv85q5d2q4z5ajdha8yu34mf0eh")
	require.NoError(t, err)
	govGenState := &types.GenesisState{
		Votes: types.Votes{
			types.Vote{ProposalId: 1, Voter: voter.String(), Option: types.OptionAbstain},
			types.Vote{ProposalId: 2, Voter: voter.String(), Option: types.OptionEmpty},
			types.Vote{ProposalId: 3, Voter: voter.String(), Option: types.OptionNo},
			types.Vote{ProposalId: 4, Voter: voter.String(), Option: types.OptionNoWithVeto},
			types.Vote{ProposalId: 5, Voter: voter.String(), Option: types.OptionYes},
		},
	}

	migrated := v043gov.MigrateJSON(govGenState)

	bz, err := clientCtx.Codec.MarshalJSON(migrated)
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
	"deposit_params": {
		"max_deposit_period": "0s",
		"min_deposit": []
	},
	"deposits": [],
	"proposals": [],
	"starting_proposal_id": "0",
	"tally_params": {
		"quorum": "0",
		"threshold": "0",
		"veto_threshold": "0"
	},
	"votes": [
		{
			"option": "VOTE_OPTION_UNSPECIFIED",
			"options": [
				{
					"option": "VOTE_OPTION_ABSTAIN",
					"weight": "1.000000000000000000"
				}
			],
			"proposal_id": "1",
			"voter": "cosmos1fl48vsnmsdzcv85q5d2q4z5ajdha8yu34mf0eh"
		},
		{
			"option": "VOTE_OPTION_UNSPECIFIED",
			"options": [
				{
					"option": "VOTE_OPTION_UNSPECIFIED",
					"weight": "1.000000000000000000"
				}
			],
			"proposal_id": "2",
			"voter": "cosmos1fl48vsnmsdzcv85q5d2q4z5ajdha8yu34mf0eh"
		},
		{
			"option": "VOTE_OPTION_UNSPECIFIED",
			"options": [
				{
					"option": "VOTE_OPTION_NO",
					"weight": "1.000000000000000000"
				}
			],
			"proposal_id": "3",
			"voter": "cosmos1fl48vsnmsdzcv85q5d2q4z5ajdha8yu34mf0eh"
		},
		{
			"option": "VOTE_OPTION_UNSPECIFIED",
			"options": [
				{
					"option": "VOTE_OPTION_NO_WITH_VETO",
					"weight": "1.000000000000000000"
				}
			],
			"proposal_id": "4",
			"voter": "cosmos1fl48vsnmsdzcv85q5d2q4z5ajdha8yu34mf0eh"
		},
		{
			"option": "VOTE_OPTION_UNSPECIFIED",
			"options": [
				{
					"option": "VOTE_OPTION_YES",
					"weight": "1.000000000000000000"
				}
			],
			"proposal_id": "5",
			"voter": "cosmos1fl48vsnmsdzcv85q5d2q4z5ajdha8yu34mf0eh"
		}
	],
	"voting_params": {
		"voting_period": "0s"
	}
}`

	fmt.Println(string(indentedBz))

	require.Equal(t, expected, string(indentedBz))
}
