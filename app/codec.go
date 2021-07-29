package app

import (
	"reflect"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/container"
)

type CodecOption = func(codectypes.TypeRegistry)

var CodecProvider = container.Options(
	container.AutoGroupTypes(
		reflect.TypeOf(func(codectypes.TypeRegistry) {}),
	),
	container.Provide(func(options []CodecOption) (
		codectypes.TypeRegistry,
		codec.Codec,
		codec.ProtoCodecMarshaler,
		codec.BinaryCodec,
		codec.JSONCodec,
		*codec.LegacyAmino,
	) {

		typeRegistry := codectypes.NewInterfaceRegistry()
		for _, option := range options {
			option(typeRegistry)
		}
		cdc := codec.NewProtoCodec(typeRegistry)
		amino := codec.NewLegacyAmino()
		return typeRegistry, cdc, cdc, cdc, cdc, amino
	}))
