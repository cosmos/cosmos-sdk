package cli

import (
	"testing"

	"github.com/spf13/pflag"

	"github.com/KiraCore/cosmos-sdk/testutil"
	"github.com/KiraCore/cosmos-sdk/testutil/testdata"

	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/crypto/secp256k1"

	"github.com/KiraCore/cosmos-sdk/client"
	"github.com/KiraCore/cosmos-sdk/codec"
	sdk "github.com/KiraCore/cosmos-sdk/types"
)

func Test_splitAndCall_NoMessages(t *testing.T) {
	clientCtx := client.Context{}

	err := newSplitAndApply(nil, clientCtx, nil, nil, 10)
	assert.NoError(t, err, "")
}

func Test_splitAndCall_Splitting(t *testing.T) {
	clientCtx := client.Context{}

	addr := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())

	// Add five messages
	msgs := []sdk.Msg{
		testdata.NewTestMsg(addr),
		testdata.NewTestMsg(addr),
		testdata.NewTestMsg(addr),
		testdata.NewTestMsg(addr),
		testdata.NewTestMsg(addr),
	}

	// Keep track of number of calls
	const chunkSize = 2

	callCount := 0
	err := newSplitAndApply(
		func(clientCtx client.Context, fs *pflag.FlagSet, msgs ...sdk.Msg) error {
			callCount++

			assert.NotNil(t, clientCtx)
			assert.NotNil(t, msgs)

			if callCount < 3 {
				assert.Equal(t, len(msgs), 2)
			} else {
				assert.Equal(t, len(msgs), 1)
			}

			return nil
		},
		clientCtx, nil, msgs, chunkSize)

	assert.NoError(t, err, "")
	assert.Equal(t, 3, callCount)
}

func TestParseProposal(t *testing.T) {
	cdc := codec.New()
	okJSON, cleanup := testutil.WriteToNewTempFile(t, `
{
  "title": "Community Pool Spend",
  "description": "Pay me some Atoms!",
  "recipient": "cosmos1s5afhd6gxevu37mkqcvvsj8qeylhn0rz46zdlq",
  "amount": "1000stake",
  "deposit": "1000stake"
}
`)
	t.Cleanup(cleanup)

	proposal, err := ParseCommunityPoolSpendProposalJSON(cdc, okJSON.Name())
	require.NoError(t, err)

	addr, err := sdk.AccAddressFromBech32("cosmos1s5afhd6gxevu37mkqcvvsj8qeylhn0rz46zdlq")
	require.NoError(t, err)

	require.Equal(t, "Community Pool Spend", proposal.Title)
	require.Equal(t, "Pay me some Atoms!", proposal.Description)
	require.Equal(t, addr, proposal.Recipient)
	require.Equal(t, "1000stake", proposal.Deposit)
	require.Equal(t, "1000stake", proposal.Amount)
}
