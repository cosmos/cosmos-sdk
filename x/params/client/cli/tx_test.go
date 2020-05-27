package cli

import (
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/params/client/utils"
)

func TestParseProposal(t *testing.T) {
	cdc := codec.New()
	okJSON, err := ioutil.TempFile("", "proposal")
	require.Nil(t, err, "unexpected error")
	_, err = okJSON.WriteString(`
{
  "title": "Staking Param Change",
  "description": "Update max validators",
  "changes": [
    {
      "subspace": "staking",
      "key": "MaxValidators",
      "value": 1
    }
  ],
  "deposit": "1000stake"
}
`)
	require.NoError(t, err)

	proposal, err := utils.ParseParamChangeProposalJSON(cdc, okJSON.Name())
	require.NoError(t, err)

	require.Equal(t, "Staking Param Change", proposal.Title)
	require.Equal(t, "Update max validators", proposal.Description)
	require.Equal(t, "1000stake", proposal.Deposit)
	require.Equal(t, utils.ParamChangesJSON{
		{
			Subspace: "staking",
			Key:      "MaxValidators",
			Value:    []byte{0x31},
		},
	}, proposal.Changes)
}
