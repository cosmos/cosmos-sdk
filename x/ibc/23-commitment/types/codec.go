package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/exported"
)

// RegisterInterfaces registers the commitment interfaces to protobuf Any.
func RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	registry.RegisterInterface(
		"cosmos.ibc.commitment.Root",
		(*exported.Root)(nil),
	)
	registry.RegisterInterface(
		"cosmos.ibc.commitment.Prefix",
		(*exported.Prefix)(nil),
	)
	registry.RegisterInterface(
		"cosmos.ibc.commitment.Path",
		(*exported.Path)(nil),
	)
	registry.RegisterInterface(
		"cosmos.ibc.commitment.Proof",
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

// RegisterCodec registers the necessary x/ibc/23-commitment interfaces and concrete types
// on the provided Amino codec. These types are used for Amino JSON serialization.
func RegisterCodec(cdc *codec.Codec) {
	cdc.RegisterInterface((*exported.Root)(nil), nil)
	cdc.RegisterInterface((*exported.Prefix)(nil), nil)
	cdc.RegisterInterface((*exported.Path)(nil), nil)
	cdc.RegisterInterface((*exported.Proof)(nil), nil)

	cdc.RegisterConcrete(MerkleRoot{}, "ibc/commitment/MerkleRoot", nil)
	cdc.RegisterConcrete(MerklePrefix{}, "ibc/commitment/MerklePrefix", nil)
	cdc.RegisterConcrete(MerklePath{}, "ibc/commitment/MerklePath", nil)
	cdc.RegisterConcrete(MerkleProof{}, "ibc/commitment/MerkleProof", nil)
}

var (
	amino = codec.New()

	// SubModuleCdc references the global x/ibc/23-commitmentl module codec. Note, the codec should
	// ONLY be used in certain instances of tests and for JSON encoding as Amino is
	// still used for that purpose.
	//
	// The actual codec used for serialization should be provided to x/ibc/23-commitmentl and
	// defined at the application level.
	SubModuleCdc = codec.NewHybridCodec(amino, codectypes.NewInterfaceRegistry())
)

func init() {
	RegisterCodec(amino)
	amino.Seal()
}
