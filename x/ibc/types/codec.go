package types

import (
	"github.com/KiraCore/cosmos-sdk/codec"
	cdctypes "github.com/KiraCore/cosmos-sdk/codec/types"
	clienttypes "github.com/KiraCore/cosmos-sdk/x/ibc/02-client/types"
	connectiontypes "github.com/KiraCore/cosmos-sdk/x/ibc/03-connection/types"
	channeltypes "github.com/KiraCore/cosmos-sdk/x/ibc/04-channel/types"
	ibctmtypes "github.com/KiraCore/cosmos-sdk/x/ibc/07-tendermint/types"
	localhosttypes "github.com/KiraCore/cosmos-sdk/x/ibc/09-localhost/types"
	commitmenttypes "github.com/KiraCore/cosmos-sdk/x/ibc/23-commitment/types"
)

// RegisterCodec registers the necessary x/ibc interfaces and concrete types
// on the provided Amino codec. These types are used for Amino JSON serialization.
func RegisterCodec(cdc *codec.Codec) {
	clienttypes.RegisterCodec(cdc)
	connectiontypes.RegisterCodec(cdc)
	channeltypes.RegisterCodec(cdc)
	ibctmtypes.RegisterCodec(cdc)
	localhosttypes.RegisterCodec(cdc)
	commitmenttypes.RegisterCodec(cdc)
}

// RegisterInterfaces registers x/ibc interfaces into protobuf Any.
func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	connectiontypes.RegisterInterfaces(registry)
	channeltypes.RegisterInterfaces(registry)
}
