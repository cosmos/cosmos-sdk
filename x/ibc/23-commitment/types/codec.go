package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
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

// UnpackAnyPrefix unpacks a protobuf Any and sets the value to the current prefix.
func UnpackAnyPrefix(any *cdctypes.Any) (exported.Prefix, error) {
	var prefix exported.Prefix
	cachedValue := any.GetCachedValue()

	if cachedValue == nil {
		registry := cdctypes.NewInterfaceRegistry()
		RegisterInterfaces(registry)

		if err := registry.UnpackAny(any, &prefix); err != nil {
			return nil, sdkerrors.Wrapf(sdkerrors.ErrProtobufAny, err.Error())
		}

		return prefix, nil
	}

	prefix, ok := cachedValue.(exported.Prefix)
	if !ok {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrProtobufAny, "cached value %T is not a valid prefix", cachedValue)
	}

	return prefix, nil
}

// UnpackAnyProof unpacks a protobuf Any and sets the value to the current proof.
func UnpackAnyProof(any *cdctypes.Any) (exported.Proof, error) {
	var proof exported.Proof
	cachedValue := any.GetCachedValue()

	if cachedValue == nil {
		registry := cdctypes.NewInterfaceRegistry()
		RegisterInterfaces(registry)

		if err := registry.UnpackAny(any, &proof); err != nil {
			return nil, sdkerrors.Wrapf(sdkerrors.ErrProtobufAny, err.Error())
		}

		return proof, nil
	}

	proof, ok := cachedValue.(exported.Proof)
	if !ok {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrProtobufAny, "cached value %T is not a valid proof", cachedValue)
	}

	return proof, nil
}
