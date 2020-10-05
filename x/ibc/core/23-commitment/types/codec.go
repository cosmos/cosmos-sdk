package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/core/exported"
)

// RegisterInterfaces registers the commitment interfaces to protobuf Any.
func RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	registry.RegisterInterface(
		"ibc.core.commitment.v1.Root",
		(*exported.Root)(nil),
	)
	registry.RegisterInterface(
		"ibc.core.commitment.v1.Prefix",
		(*exported.Prefix)(nil),
	)
	registry.RegisterInterface(
		"ibc.core.commitment.v1.Path",
		(*exported.Path)(nil),
	)
	registry.RegisterInterface(
		"ibc.core.commitment.v1.Proof",
		(*exported.Proof)(nil),
	)

	registry.RegisterImplementations(
		(*exported.Root)(nil),
		&MerkleRoot{},
	)
	registry.RegisterImplementations(
		(*exported.Prefix)(nil),
		&MerklePrefix{},
	)
	registry.RegisterImplementations(
		(*exported.Path)(nil),
		&MerklePath{},
	)
	registry.RegisterImplementations(
		(*exported.Proof)(nil),
		&MerkleProof{},
	)
}

var (
	// SubModuleCdc references the global x/ibc/core/23-commitmentl module codec. Note, the codec should
	// ONLY be used in certain instances of tests and for JSON encoding.
	//
	// The actual codec used for serialization should be provided to x/ibc/core/23-commitmentl and
	// defined at the application level.
	SubModuleCdc = codec.NewProtoCodec(codectypes.NewInterfaceRegistry())
)
