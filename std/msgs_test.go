package std_test

import (
	std2 "github.com/cosmos/cosmos-sdk/std"
	"testing"
	"time"

	gov "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/stretchr/testify/require"

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

	msg, err := std2.NewMsgSubmitEvidence(e, s)
	require.NoError(t, err)
	require.Equal(t, msg.GetEvidence(), &e)
	require.Equal(t, msg.GetSubmitter(), s)
	require.NoError(t, msg.ValidateBasic())
}

type invalidProposal struct {
	*gov.TextProposal
}

func TestMsgSubmitProposal(t *testing.T) {
	p := sdk.AccAddress("foo")
	d := sdk.NewCoins(sdk.NewInt64Coin("stake", 1000))
	c := gov.NewTextProposal("title", "description")

	//
	// test constructor
	//

	msg, err := std2.NewMsgSubmitProposal(c, d, p)
	require.NoError(t, err)
	require.Equal(t, msg.GetContent(), c)
	require.Equal(t, msg.GetProposer(), p)
	require.Equal(t, msg.GetInitialDeposit(), d)
	require.NoError(t, msg.ValidateBasic())

	_, err = std2.NewMsgSubmitProposal(invalidProposal{}, d, p)
	require.Error(t, err)

	//
	// test setter methods
	//

	msg = &std2.MsgSubmitProposal{}
	msg.SetProposer(p)
	msg.SetInitialDeposit(d)
	err = msg.SetContent(c)
	require.NoError(t, err)
	require.Equal(t, msg.GetContent(), c)
	require.Equal(t, msg.GetProposer(), p)
	require.Equal(t, msg.GetInitialDeposit(), d)
	require.NoError(t, msg.ValidateBasic())

	msg = &std2.MsgSubmitProposal{}
	err = msg.SetContent(invalidProposal{})
	require.Error(t, err)

}
