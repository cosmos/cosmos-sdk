package types

import (
	"github.com/KiraCore/cosmos-sdk/codec"
	cdctypes "github.com/KiraCore/cosmos-sdk/codec/types"
	client "github.com/KiraCore/cosmos-sdk/x/ibc/02-client"
	connection "github.com/KiraCore/cosmos-sdk/x/ibc/03-connection"
	channel "github.com/KiraCore/cosmos-sdk/x/ibc/04-channel"
	ibctmtypes "github.com/KiraCore/cosmos-sdk/x/ibc/07-tendermint/types"
	localhosttypes "github.com/KiraCore/cosmos-sdk/x/ibc/09-localhost/types"
	commitmenttypes "github.com/KiraCore/cosmos-sdk/x/ibc/23-commitment/types"
)

// RegisterCodec registers the necessary x/ibc interfaces and concrete types
// on the provided Amino codec. These types are used for Amino JSON serialization.
func RegisterCodec(cdc *codec.Codec) {
	client.RegisterCodec(cdc)
	connection.RegisterCodec(cdc)
	channel.RegisterCodec(cdc)
	ibctmtypes.RegisterCodec(cdc)
	localhosttypes.RegisterCodec(cdc)
	commitmenttypes.RegisterCodec(cdc)
}

// RegisterInterfaces registers x/ibc interfaces into protobuf Any.
func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	connection.RegisterInterfaces(registry)
	channel.RegisterInterfaces(registry)
}
