package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
	connectiontypes "github.com/cosmos/cosmos-sdk/x/ibc/03-connection/types"
	channel "github.com/cosmos/cosmos-sdk/x/ibc/04-channel"
	ibctmtypes "github.com/cosmos/cosmos-sdk/x/ibc/07-tendermint/types"
	localhosttypes "github.com/cosmos/cosmos-sdk/x/ibc/09-localhost/types"
	commitmenttypes "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/types"
)

// RegisterCodec registers the necessary x/ibc interfaces and concrete types
// on the provided Amino codec. These types are used for Amino JSON serialization.
func RegisterCodec(cdc *codec.Codec) {
	clienttypes.RegisterCodec(cdc)
	connectiontypes.RegisterCodec(cdc)
	channel.RegisterCodec(cdc)
	ibctmtypes.RegisterCodec(cdc)
	localhosttypes.RegisterCodec(cdc)
	commitmenttypes.RegisterCodec(cdc)
}

// RegisterInterfaces registers x/ibc interfaces into protobuf Any.
func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	connectiontypes.RegisterInterfaces(registry)
	channel.RegisterInterfaces(registry)
}
