package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	clientexported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
)

// RegisterCodec registers the necessary x/ibc/07-tendermint interfaces and conrete types
// on the provided Amino codec. These types are used for Amino JSON serialization.
func RegisterCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(ClientState{}, "ibc/client/tendermint/ClientState", nil)
	cdc.RegisterConcrete(ConsensusState{}, "ibc/client/tendermint/ConsensusState", nil)
	cdc.RegisterConcrete(Header{}, "ibc/client/tendermint/Header", nil)
	cdc.RegisterConcrete(Evidence{}, "ibc/client/tendermint/Evidence", nil)
	cdc.RegisterConcrete(&MsgCreateClient{}, "ibc/client/tendermint/MsgCreateClient", nil)
	cdc.RegisterConcrete(&MsgUpdateClient{}, "ibc/client/tendermint/MsgUpdateClient", nil)
	cdc.RegisterConcrete(&MsgSubmitClientMisbehaviour{}, "ibc/client/tendermint/MsgSubmitClientMisbehaviour", nil)
}

// RegisterInterfaces registers the tendermint concrete evidence and client-related
// implementations and interfaces.
func RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	registry.RegisterImplementations(
		(*clientexported.ClientState)(nil),
		&ClientState{},
	)
	registry.RegisterImplementations(
		(*clientexported.ConsensusState)(nil),
		&ConsensusState{},
	)
}

var (
	amino = codec.New()

	// SubModuleCdc references the global x/ibc/07-tendermint module codec. Note, the codec should
	// ONLY be used in certain instances of tests and for JSON encoding as Amino is
	// still used for that purpose.
	//
	// The actual codec used for serialization should be provided to x/ibc/07-tendermint and
	// defined at the application level.
	SubModuleCdc = codec.NewProtoCodec(codectypes.NewInterfaceRegistry())
)

func init() {
	cryptocodec.RegisterCrypto(amino)
	RegisterCodec(amino)
	amino.Seal()
}
