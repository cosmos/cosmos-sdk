package cli

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/testutil"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
)

func TestParseSubmitLegacyProposalFlags(t *testing.T) {
	okJSON := testutil.WriteToNewTempFile(t, `
{
  "title": "Test Proposal",
  "description": "My awesome proposal",
  "type": "Text",
  "deposit": "1000test"
}
`)

	badJSON := testutil.WriteToNewTempFile(t, "bad json")
	fs := NewCmdSubmitLegacyProposal().Flags()

	// nonexistent json
	fs.Set(FlagProposal, "fileDoesNotExist")
	_, err := parseSubmitLegacyProposalFlags(fs)
	require.Error(t, err)

	// invalid json
	fs.Set(FlagProposal, badJSON.Name())
	_, err = parseSubmitLegacyProposalFlags(fs)
	require.Error(t, err)

	// ok json
	fs.Set(FlagProposal, okJSON.Name())
	proposal1, err := parseSubmitLegacyProposalFlags(fs)
	require.Nil(t, err, "unexpected error")
	require.Equal(t, "Test Proposal", proposal1.Title)
	require.Equal(t, "My awesome proposal", proposal1.Description)
	require.Equal(t, "Text", proposal1.Type)
	require.Equal(t, "1000test", proposal1.Deposit)

	// flags that can't be used with --proposal
	for _, incompatibleFlag := range ProposalFlags {
		fs.Set(incompatibleFlag, "some value")
		_, err := parseSubmitLegacyProposalFlags(fs)
		require.Error(t, err)
		fs.Set(incompatibleFlag, "")
	}

	// no --proposal, only flags
	fs.Set(FlagProposal, "")
	fs.Set(FlagTitle, proposal1.Title)
	fs.Set(FlagDescription, proposal1.Description)
	fs.Set(FlagProposalType, proposal1.Type)
	fs.Set(FlagDeposit, proposal1.Deposit)
	proposal2, err := parseSubmitLegacyProposalFlags(fs)

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

func TestparseSubmitProposal(t *testing.T) {
	_, _, addr := testdata.KeyTestPubAddr()
	encCfg := simapp.MakeTestEncodingConfig()

	okJSON := testutil.WriteToNewTempFile(t, fmt.Sprintf(`
{
	"messages": [
		{
			"@type":"/cosmos.bank.v1beta1.MsgSend",
			"from_address":"%s",
			"to_address":"%s",
			"amount":[{"denom":"stake","amount":"10"}]
		}
  	],
	"deposit": "1000test"
}
`, addr, addr))

	badJSON := testutil.WriteToNewTempFile(t, "bad json")
	fs := NewCmdSubmitProposal().Flags()

	// nonexistent json
	_, _, _, err := parseSubmitProposal(encCfg.Codec, "fileDoesNotExist", fs)
	require.Error(t, err)

	// invalid json
	_, _, _, err = parseSubmitProposal(encCfg.Codec, badJSON.Name(), fs)
	require.Error(t, err)

	// ok json
	fs.Set(FlagProposal, okJSON.Name())
	_, _, proposal1, err := parseSubmitProposal(fs)
	require.Nil(t, err, "unexpected error")
	require.Equal(t, "Test Proposal", proposal1.Title)
	require.Equal(t, "My awesome proposal", proposal1.Description)
	require.Equal(t, "Text", proposal1.Type)
	require.Equal(t, "1000test", proposal1.Deposit)

	// flags that can't be used with --proposal
	for _, incompatibleFlag := range ProposalFlags {
		fs.Set(incompatibleFlag, "some value")
		_, _, _, err := parseSubmitProposal(fs)
		require.Error(t, err)
		fs.Set(incompatibleFlag, "")
	}

	// no --proposal, only flags
	fs.Set(FlagProposal, "")
	fs.Set(FlagTitle, proposal1.Title)
	fs.Set(FlagDescription, proposal1.Description)
	fs.Set(FlagProposalType, proposal1.Type)
	fs.Set(FlagDeposit, proposal1.Deposit)
	_, _, proposal2, err := parseSubmitProposal(fs)

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
