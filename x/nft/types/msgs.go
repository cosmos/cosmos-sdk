package types

import sdk "github.com/cosmos/cosmos-sdk/types"


const (
	// TypeMsgSend nft message types
	TypeMsgSend = "send"
)

var (
	_ sdk.Msg = &MsgIssue{}
	_ sdk.Msg = &MsgMint{}
	_ sdk.Msg = &MsgEdit{}
	_ sdk.Msg = &MsgSend{}
	_ sdk.Msg = &MsgBurn{}
)

func (m *MsgIssue) ValidateBasic() error {
	panic("implement me")
}

func (m *MsgIssue) GetSigners() []sdk.AccAddress {
	panic("implement me")
}

func (m *MsgMint) ValidateBasic() error {
	panic("implement me")
}

func (m *MsgMint) GetSigners() []sdk.AccAddress {
	panic("implement me")
}


func (m *MsgEdit) ValidateBasic() error {
	panic("implement me")
}

func (m *MsgEdit) GetSigners() []sdk.AccAddress {
	panic("implement me")
}

func (m *MsgSend) ValidateBasic() error {
	panic("implement me")
}

func (m *MsgSend) GetSigners() []sdk.AccAddress {
	panic("implement me")
}

func (m *MsgBurn) ValidateBasic() error {
	panic("implement me")
}

func (m *MsgBurn) GetSigners() []sdk.AccAddress {
	panic("implement me")
}
