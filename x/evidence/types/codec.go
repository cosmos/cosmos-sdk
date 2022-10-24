package types

import (
	"github.com/pointnetwork/cosmos-point-sdk/codec"
	"github.com/pointnetwork/cosmos-point-sdk/codec/legacy"
	"github.com/pointnetwork/cosmos-point-sdk/codec/types"
	cryptocodec "github.com/pointnetwork/cosmos-point-sdk/crypto/codec"
	sdk "github.com/pointnetwork/cosmos-point-sdk/types"
	"github.com/pointnetwork/cosmos-point-sdk/types/msgservice"
	authzcodec "github.com/pointnetwork/cosmos-point-sdk/x/authz/codec"
	"github.com/pointnetwork/cosmos-point-sdk/x/evidence/exported"
)

// RegisterLegacyAminoCodec registers all the necessary types and interfaces for the
// evidence module.
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterInterface((*exported.Evidence)(nil), nil)
	legacy.RegisterAminoMsg(cdc, &MsgSubmitEvidence{}, "cosmos-sdk/MsgSubmitEvidence")
	cdc.RegisterConcrete(&Equivocation{}, "cosmos-sdk/Equivocation", nil)
}

func RegisterInterfaces(registry types.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil), &MsgSubmitEvidence{})
	registry.RegisterInterface(
		"cosmos.evidence.v1beta1.Evidence",
		(*exported.Evidence)(nil),
		&Equivocation{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}

var (
	amino     = codec.NewLegacyAmino()
	ModuleCdc = codec.NewAminoCodec(amino)
)

func init() {
	RegisterLegacyAminoCodec(amino)
	cryptocodec.RegisterCrypto(amino)
	sdk.RegisterLegacyAminoCodec(amino)

	// Register all Amino interfaces and concrete types on the authz Amino codec so that this can later be
	// used to properly serialize MsgGrant and MsgExec instances
	RegisterLegacyAminoCodec(authzcodec.Amino)
}
