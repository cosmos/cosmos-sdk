package types

import sdk "github.com/cosmos/cosmos-sdk/types"

var (
	_ sdk.Msg = &MsgFundCommunityPool{}
	_ sdk.Msg = &MsgCommunityPoolSpend{}
)
