package cli

import (
	"encoding/base64"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/testutil"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
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
	flagTestCases := map[string]struct {
		pTitle       string
		pDescription string
		pType        string
		expErr       bool
		errMsg       string
	}{
		"valid flags": {
			pTitle:       proposal1.Title,
			pDescription: proposal1.Description,
			pType:        proposal1.Type,
		},
		"empty type": {
			pTitle:       proposal1.Title,
			pDescription: proposal1.Description,
			expErr:       true,
			errMsg:       "proposal type is required",
		},
		"empty title": {
			pDescription: proposal1.Description,
			pType:        proposal1.Type,
			expErr:       true,
			errMsg:       "proposal title is required",
		},
		"empty description": {
			pTitle: proposal1.Title,
			pType:  proposal1.Type,
			expErr: true,
			errMsg: "proposal description is required",
		},
	}
	for name, tc := range flagTestCases {
		t.Run(name, func(t *testing.T) {
			fs.Set(FlagTitle, tc.pTitle)
			fs.Set(FlagDescription, tc.pDescription)
			fs.Set(FlagProposalType, tc.pType)
			fs.Set(FlagDeposit, proposal1.Deposit)
			proposal2, err := parseSubmitLegacyProposalFlags(fs)

			if tc.expErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.errMsg)
			} else {
				require.NoError(t, err)
				require.Equal(t, proposal1.Title, proposal2.Title)
				require.Equal(t, proposal1.Description, proposal2.Description)
				require.Equal(t, proposal1.Type, proposal2.Type)
				require.Equal(t, proposal1.Deposit, proposal2.Deposit)
			}
		})
	}

	err = okJSON.Close()
	require.Nil(t, err, "unexpected error")
	err = badJSON.Close()
	require.Nil(t, err, "unexpected error")
}

func TestParseSubmitProposal(t *testing.T) {
	_, _, addr := testdata.KeyTestPubAddr()
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(interfaceRegistry)
	banktypes.RegisterInterfaces(interfaceRegistry)
	stakingtypes.RegisterInterfaces(interfaceRegistry)
	v1beta1.RegisterInterfaces(interfaceRegistry)
	v1.RegisterInterfaces(interfaceRegistry)
	expectedMetadata := []byte{42}

	okJSON := testutil.WriteToNewTempFile(t, fmt.Sprintf(`
{
	"messages": [
		{
			"@type": "/cosmos.bank.v1beta1.MsgSend",
			"from_address": "%s",
			"to_address": "%s",
			"amount":[{"denom": "stake","amount": "10"}]
		},
		{
			"@type": "/cosmos.staking.v1beta1.MsgDelegate",
			"delegator_address": "%s",
			"validator_address": "%s",
			"amount":{"denom": "stake","amount": "10"}
		},
		{
			"@type": "/cosmos.gov.v1.MsgExecLegacyContent",
			"authority": "%s",
			"content": {
				"@type": "/cosmos.gov.v1beta1.TextProposal",
				"title": "My awesome title",
				"description": "My awesome description"
			}
		}
  	],
	"metadata": "%s",
	"deposit": "1000test"
}
`, addr, addr, addr, addr, addr, base64.StdEncoding.EncodeToString(expectedMetadata)))

	badJSON := testutil.WriteToNewTempFile(t, "bad json")

	// nonexistent json
	_, _, _, err := parseSubmitProposal(cdc, "fileDoesNotExist")
	require.Error(t, err)

	// invalid json
	_, _, _, err = parseSubmitProposal(cdc, badJSON.Name())
	require.Error(t, err)

	// ok json
	msgs, metadata, deposit, err := parseSubmitProposal(cdc, okJSON.Name())
	require.NoError(t, err, "unexpected error")
	require.Equal(t, sdk.NewCoins(sdk.NewCoin("test", sdk.NewInt(1000))), deposit)
	require.Equal(t, base64.StdEncoding.EncodeToString(expectedMetadata), metadata)
	require.Len(t, msgs, 3)
	msg1, ok := msgs[0].(*banktypes.MsgSend)
	require.True(t, ok)
	require.Equal(t, addr.String(), msg1.FromAddress)
	require.Equal(t, addr.String(), msg1.ToAddress)
	require.Equal(t, sdk.NewCoins(sdk.NewCoin("stake", sdk.NewInt(10))), msg1.Amount)
	msg2, ok := msgs[1].(*stakingtypes.MsgDelegate)
	require.True(t, ok)
	require.Equal(t, addr.String(), msg2.DelegatorAddress)
	require.Equal(t, addr.String(), msg2.ValidatorAddress)
	require.Equal(t, sdk.NewCoin("stake", sdk.NewInt(10)), msg2.Amount)
	msg3, ok := msgs[2].(*v1.MsgExecLegacyContent)
	require.True(t, ok)
	require.Equal(t, addr.String(), msg3.Authority)
	textProp, ok := msg3.Content.GetCachedValue().(*v1beta1.TextProposal)
	require.True(t, ok)
	require.Equal(t, "My awesome title", textProp.Title)
	require.Equal(t, "My awesome description", textProp.Description)

	err = okJSON.Close()
	require.Nil(t, err, "unexpected error")
	err = badJSON.Close()
	require.Nil(t, err, "unexpected error")
}
