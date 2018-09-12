package cli

import (
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"testing"
)

func TestParseSubmitProposalFlags(t *testing.T) {
	okJSON, err := ioutil.TempFile("", "proposal")
	require.Nil(t, err, "unexpected error")
	okJSON.WriteString(`
{
  "title": "Test Proposal",
  "description": "My awesome proposal",
  "type": "Text",
  "deposit": "1000test"
}
`)

	badJSON, err := ioutil.TempFile("", "proposal")
	require.Nil(t, err, "unexpected error")
	badJSON.WriteString("bad json")

	// nonexistent json
	viper.Set(flagProposal, "fileDoesNotExist")
	_, err = parseSubmitProposalFlags()
	require.Error(t, err)

	// invalid json
	viper.Set(flagProposal, badJSON.Name())
	_, err = parseSubmitProposalFlags()
	require.Error(t, err)

	// ok json
	viper.Set(flagProposal, okJSON.Name())
	proposal1, err := parseSubmitProposalFlags()
	require.Nil(t, err, "unexpected error")
	require.Equal(t, "Test Proposal", proposal1.Title)
	require.Equal(t, "My awesome proposal", proposal1.Description)
	require.Equal(t, "Text", proposal1.Type)
	require.Equal(t, "1000test", proposal1.Deposit)

	// flags that can't be used with --proposal
	for _, incompatibleFlag := range proposalFlags {
		viper.Set(incompatibleFlag, "some value")
		_, err := parseSubmitProposalFlags()
		require.Error(t, err)
		viper.Set(incompatibleFlag, "")
	}

	// no --proposal, only flags
	viper.Set(flagProposal, "")
	viper.Set(flagTitle, proposal1.Title)
	viper.Set(flagDescription, proposal1.Description)
	viper.Set(flagProposalType, proposal1.Type)
	viper.Set(flagDeposit, proposal1.Deposit)
	proposal2, err := parseSubmitProposalFlags()
	require.Nil(t, err, "unexpected error")
	require.Equal(t, proposal1.Title, proposal2.Title)
	require.Equal(t, proposal1.Description, proposal2.Description)
	require.Equal(t, proposal1.Type, proposal2.Type)
	require.Equal(t, proposal1.Deposit, proposal2.Deposit)

	err = okJSON.Close()
	require.Nil(t, err, "unexpected error")
	err = badJSON.Close()
	require.Nil(t, err, "unexpected error")
}
