package std_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/codec/std"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/evidence"
	"github.com/cosmos/cosmos-sdk/x/gov"
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

func TestNewNewMsgSubmitProposal(t *testing.T) {
	p := sdk.AccAddress("foo")
	d := sdk.NewCoins(sdk.NewInt64Coin("stake", 1000))
	c := gov.TextProposal{Title: "title", Description: "description"}

	msg, err := std.NewMsgSubmitProposal(c, d, p)
	require.NoError(t, err)
	require.Equal(t, msg.GetContent(), &c)
	require.Equal(t, msg.GetProposer(), p)
	require.Equal(t, msg.GetInitialDeposit(), d)
	require.NoError(t, msg.ValidateBasic())
}
