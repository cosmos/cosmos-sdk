package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/exported"
)

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
	cdc.RegisterConcrete(SignaturePrefix{}, "ibc/commitment/SignaturePrefix", nil)
	cdc.RegisterConcrete(SignatureProof{}, "ibc/commitment/SignatureProof", nil)
}

// RegisterInterfaces associates a proto name with the Prefix and Proof
// interfaces and creates a registry of their concrete implementations.
func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterInterface(
		"cosmos_sdk.x.ibc.commitment.v1.Prefix",
		(*exported.Prefix)(nil),
		&MerklePrefix{},
		&SignaturePrefix{},
	)
	registry.RegisterInterface(
		"cosmos_sdk.x.ibc.commitment.v1.Proof",
		(*exported.Proof)(nil),
		&MerkleProof{},
		&SignatureProof{},
	)
	registry.RegisterImplementations(
		(*exported.Prefix)(nil),
		&MerklePrefix{},
		&SignaturePrefix{},
	)
	registry.RegisterImplementations(
		(*exported.Proof)(nil),
		&MerkleProof{},
		&SignatureProof{},
	)
}

var (
	amino = codec.New()

	// SubModuleCdc references the global x/ibc/23-commitmentl module codec. Note, the codec should
	// ONLY be used in certain instances of tests and for JSON encoding as Amino is
	// still used for that purpose.
	//
	// The actual codec used for serialization should be provided to x/ibc/23-commitmentl and
	// defined at the application level.
	SubModuleCdc = codec.NewHybridCodec(amino, cdctypes.NewInterfaceRegistry())
)

func init() {
	RegisterCodec(amino)
	amino.Seal()
}
