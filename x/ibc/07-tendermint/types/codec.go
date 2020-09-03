package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/exported"
)

// RegisterInterfaces registers the tendermint concrete evidence and client-related
// implementations and interfaces.
func RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	registry.RegisterImplementations(
		(*exported.ClientState)(nil),
		&ClientState{},
	)
	registry.RegisterImplementations(
		(*exported.ConsensusState)(nil),
		&ConsensusState{},
	)
	registry.RegisterImplementations(
		(*exported.Header)(nil),
		&Header{},
	)
	registry.RegisterImplementations(
		(*exported.Misbehaviour)(nil),
		&Misbehaviour{},
	)
	registry.RegisterImplementations(
		(*exported.Header)(nil),
		&Header{},
	)
}

var (
	// SubModuleCdc references the global x/ibc/07-tendermint module codec. Note, the codec should
	// ONLY be used in certain instances of tests and for JSON encoding.
	//
	// The actual codec used for serialization should be provided to x/ibc/07-tendermint and
	// defined at the application level.
	SubModuleCdc = codec.NewProtoCodec(codectypes.NewInterfaceRegistry())
)
