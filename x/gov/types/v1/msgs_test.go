package v1_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
)

var (
	coinsPos   = sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 1000))
	coinsMulti = sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 1000), sdk.NewInt64Coin("foo", 10000))
	addrs      = []sdk.AccAddress{
		sdk.AccAddress("test1"),
		sdk.AccAddress("test2"),
	}
)

func init() {
	coinsMulti.Sort()
}

func TestMsgDepositGetSignBytes(t *testing.T) {
	addr := sdk.AccAddress("addr1")
	msg := v1.NewMsgDeposit(addr, 0, coinsPos)
	pc := codec.NewProtoCodec(types.NewInterfaceRegistry())
	res, err := pc.MarshalAminoJSON(msg)
	require.NoError(t, err)
	expected := `{"type":"cosmos-sdk/v1/MsgDeposit","value":{"amount":[{"amount":"1000","denom":"stake"}],"depositor":"cosmos1v9jxgu33kfsgr5","proposal_id":"0"}}`
	require.Equal(t, expected, string(res))
}

// this tests that Amino JSON MsgSubmitProposal.GetSignBytes() still works with Content as Any using the ModuleCdc
func TestMsgSubmitProposal_GetSignBytes(t *testing.T) {
	pc := codec.NewProtoCodec(types.NewInterfaceRegistry())
	testcases := []struct {
		name      string
		proposal  []sdk.Msg
		title     string
		summary   string
		expedited bool
		expSignBz string
	}{
		{
			"MsgVote",
			[]sdk.Msg{v1.NewMsgVote(addrs[0], 1, v1.OptionYes, "")},
			"gov/MsgVote",
			"Proposal for a governance vote msg",
			false,
			`{"type":"cosmos-sdk/v1/MsgSubmitProposal","value":{"initial_deposit":[],"messages":[{"type":"cosmos-sdk/v1/MsgVote","value":{"option":1,"proposal_id":"1","voter":"cosmos1w3jhxap3gempvr"}}],"summary":"Proposal for a governance vote msg","title":"gov/MsgVote"}}`,
		},
		{
			"MsgSend",
			[]sdk.Msg{banktypes.NewMsgSend(addrs[0], addrs[0], sdk.NewCoins())},
			"bank/MsgSend",
			"Proposal for a bank msg send",
			false,
			fmt.Sprintf(`{"type":"cosmos-sdk/v1/MsgSubmitProposal","value":{"initial_deposit":[],"messages":[{"type":"cosmos-sdk/MsgSend","value":{"amount":[],"from_address":"%s","to_address":"%s"}}],"summary":"Proposal for a bank msg send","title":"bank/MsgSend"}}`, addrs[0], addrs[0]),
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			msg, err := v1.NewMsgSubmitProposal(tc.proposal, sdk.NewCoins(), sdk.AccAddress{}.String(), "", tc.title, tc.summary, tc.expedited)
			require.NoError(t, err)
			bz, err := pc.MarshalAminoJSON(msg)
			require.NoError(t, err)
			require.Equal(t, tc.expSignBz, string(bz))
		})
	}
}
