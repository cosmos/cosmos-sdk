package types

import (
	corelegacy "cosmossdk.io/core/legacy"
	"cosmossdk.io/core/registry"
	coretransaction "cosmossdk.io/core/transaction"

	"github.com/cosmos/cosmos-sdk/codec/legacy"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

// RegisterLegacyAminoCodec registers the necessary x/bank interfaces and concrete types
// on the provided LegacyAmino codec. These types are used for Amino JSON serialization.
func RegisterLegacyAminoCodec(cdc corelegacy.Amino) {
	legacy.RegisterAminoMsg(cdc, &MsgSend{}, "cosmos-sdk/MsgSend")
	legacy.RegisterAminoMsg(cdc, &MsgMultiSend{}, "cosmos-sdk/MsgMultiSend")
	legacy.RegisterAminoMsg(cdc, &MsgUpdateParams{}, "cosmos-sdk/x/bank/MsgUpdateParams")
	legacy.RegisterAminoMsg(cdc, &MsgSetSendEnabled{}, "cosmos-sdk/MsgSetSendEnabled")

	cdc.RegisterConcrete(&SendAuthorization{}, "cosmos-sdk/SendAuthorization")
	cdc.RegisterConcrete(&Params{}, "cosmos-sdk/x/bank/Params")
}

func RegisterInterfaces(registrar registry.InterfaceRegistrar) {
	registrar.RegisterImplementations((*coretransaction.Msg)(nil),
		&MsgSend{},
		&MsgMultiSend{},
		&MsgUpdateParams{},
		&MsgBurn{},
		&MsgSetSendEnabled{},
	)

	msgservice.RegisterMsgServiceDesc(registrar, &_Msg_serviceDesc)
}
