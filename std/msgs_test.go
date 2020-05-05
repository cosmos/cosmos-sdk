package std_test

import (
	"testing"

	gov "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/std"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

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

	msg, err := std.NewMsgSubmitProposal(c, d, p)
	require.NoError(t, err)
	require.Equal(t, msg.GetContent(), c)
	require.Equal(t, msg.GetProposer(), p)
	require.Equal(t, msg.GetInitialDeposit(), d)
	require.NoError(t, msg.ValidateBasic())

	_, err = std.NewMsgSubmitProposal(invalidProposal{}, d, p)
	require.Error(t, err)

	//
	// test setter methods
	//

	msg = &std.MsgSubmitProposal{}
	msg.SetProposer(p)
	msg.SetInitialDeposit(d)
	err = msg.SetContent(c)
	require.NoError(t, err)
	require.Equal(t, msg.GetContent(), c)
	require.Equal(t, msg.GetProposer(), p)
	require.Equal(t, msg.GetInitialDeposit(), d)
	require.NoError(t, msg.ValidateBasic())

	msg = &std.MsgSubmitProposal{}
	err = msg.SetContent(invalidProposal{})
	require.Error(t, err)

}
