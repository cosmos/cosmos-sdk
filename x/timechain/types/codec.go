// SPDX-License-Identifier: MPL-2.0
// Copyright © 2025 Timechain-Arweave-LunCoSim Contributors

package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

// RegisterLegacyAminoCodec registers the necessary x/timechain interfaces and concrete types
// on the provided LegacyAmino codec. These types are used for Amino JSON serialization.
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgProposeSlot{}, "timechain/ProposeSlot", nil)
	cdc.RegisterConcrete(&MsgConfirmSlot{}, "timechain/ConfirmSlot", nil)
	cdc.RegisterConcrete(&MsgRelayEvent{}, "timechain/RelayEvent", nil)
}

// RegisterInterfaces registers the x/timechain interfaces types with the interface registry
func RegisterInterfaces(registry types.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgProposeSlot{},
		&MsgConfirmSlot{},
		&MsgRelayEvent{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}

var (
	amino = codec.NewLegacyAmino()
	// ModuleCdc references the global x/timechain module codec. Note, the codec should
	// ONLY be used in certain instances of tests and export genesis file.
	//
	// The actual codec used for serialization in the application is assigned in the Go
	// client constructor.
	ModuleCdc = codec.NewProtoCodec(types.NewInterfaceRegistry())
)

func init() {
	RegisterLegacyAminoCodec(amino)
	// Register all Amino interfaces and concrete types on the authz Amino codec so that this can later be
	// used to properly serialize MsgGrant and MsgExec instances
	// authzcodec.RegisterLegacyAminoCodec(amino)
}
