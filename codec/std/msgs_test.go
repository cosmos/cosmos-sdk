package std_test

import (
	"testing"
	"time"

	gov "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/codec/std"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/evidence"
)

func TestNewMsgSubmitEvidence(t *testing.T) {
	s := sdk.AccAddress("foo")
	e := evidence.Equivocation{
		Height:           100,
		Time:             time.Now().UTC(),
		Power:            40000000000,
		ConsensusAddress: sdk.ConsAddress("test"),
	}

	msg, err := std.NewMsgSubmitEvidence(e, s)
	require.NoError(t, err)
	require.Equal(t, msg.GetEvidence(), &e)
	require.Equal(t, msg.GetSubmitter(), s)
	require.NoError(t, msg.ValidateBasic())
}

func TestNewMsgSubmitProposalI(t *testing.T) {
	p := sdk.AccAddress("foo")
	d := sdk.NewCoins(sdk.NewInt64Coin("stake", 1000))
	c := gov.TextProposal{Title: "title", Description: "description"}

	cdc := &std.Codec{}
	msg, err := gov.NewMsgSubmitProposalI(cdc, c, d, p)
	require.NoError(t, err)
	require.Equal(t, msg.GetContent(), &c)
	require.Equal(t, msg.GetProposer(), p)
	require.Equal(t, msg.GetInitialDeposit(), d)
	require.NoError(t, msg.ValidateBasic())
}
