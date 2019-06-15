package delegate

import cosmos "github.com/cosmos/cosmos-sdk/types"

type MsgDelegatedAction struct {
	Actor  cosmos.AccAddress
	Action Action
}

func (msg MsgDelegatedAction) Route() string {
	panic("implement me")
}

func (msg MsgDelegatedAction) Type() string {
	panic("implement me")
}

func (msg MsgDelegatedAction) ValidateBasic() cosmos.Error {
	panic("implement me")
}

func (msg MsgDelegatedAction) GetSignBytes() []byte {
	panic("implement me")
}

func (msg MsgDelegatedAction) GetSigners() []cosmos.AccAddress {
	return []cosmos.AccAddress{msg.Actor}
}
