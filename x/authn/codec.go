package authn

import "github.com/cosmos/cosmos-sdk/core/codec"

func RegisterTypes(registry codec.TypeRegistry) {
	registry.RegisterMsgServiceDesc(_Msg_serviceDesc, NewMsgClient)
	registry.RegisterQueryServiceDesc(_Query_serviceDesc, NewQueryClient)
}
