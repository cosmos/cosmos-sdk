package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

// RegisterLegacyAminoCodec registers concrete types and interfaces on the given codec.
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterInterface((*FeeAllowanceI)(nil), nil)
	cdc.RegisterInterface((*sdk.MsgRequest)(nil), nil)
	cdc.RegisterConcrete(&MsgGrantFeeAllowance{}, "cosmos-sdk/MsgGrantFeeAllowance", nil)
	cdc.RegisterConcrete(&MsgRevokeFeeAllowance{}, "cosmos-sdk/MsgRevokeFeeAllowance", nil)
	cdc.RegisterConcrete(&BasicFeeAllowance{}, "cosmos-sdk/BasicFeeAllowance", nil)
	cdc.RegisterConcrete(&PeriodicFeeAllowance{}, "cosmos-sdk/PeriodicFeeAllowance", nil)
	// cdc.RegisterConcrete(FeeGrantTx{}, "cosmos-sdk/FeeGrantTx", nil)
}

// RegisterInterfaces registers the interfaces types with the interface registry
func RegisterInterfaces(registry types.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgGrantFeeAllowance{},
		&MsgRevokeFeeAllowance{},
	)

	registry.RegisterInterface(
		"cosmos.authz.v1beta1.FeeAllowance",
		(*FeeAllowanceI)(nil),
		&BasicFeeAllowance{},
		&PeriodicFeeAllowance{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}

var (
	amino = codec.NewLegacyAmino()

	// ModuleCdc references the global x/feegrant module codec. Note, the codec should
	// ONLY be used in certain instances of tests and for JSON encoding as Amino is
	// still used for that purpose.
	//
	// The actual codec used for serialization should be provided to x/feegrant and
	// defined at the application level.
	ModuleCdc = codec.NewAminoCodec(amino)
)

func init() {
	RegisterLegacyAminoCodec(amino)
	cryptocodec.RegisterCrypto(amino)
	amino.Seal()
}
