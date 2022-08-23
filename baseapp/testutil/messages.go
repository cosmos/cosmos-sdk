package testutil

import (
	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

func RegisterInterfaces(registry types.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil), &MsgCounter{})
	msgservice.RegisterMsgServiceDesc(registry, &_Counter_serviceDesc)
}

var _ sdk.Msg = &MsgCounter{}

func (msg *MsgCounter) GetSigners() []sdk.AccAddress { return []sdk.AccAddress{} }
func (msg *MsgCounter) ValidateBasic() error         { return nil }
