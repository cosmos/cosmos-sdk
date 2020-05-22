// +build !test_amino

package simapp

import (
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	"github.com/cosmos/cosmos-sdk/std"
	"github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/ibc"
)

func MakeEncodingConfig() EncodingConfig {
	cdc := std.MakeCodec(ModuleBasics)
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	std.RegisterInterfaces(interfaceRegistry)
	ModuleBasics.RegisterInterfaceModules(interfaceRegistry)
	marshaler := codec.NewHybridCodec(cdc, interfaceRegistry)
	pubKeyCodec := cryptocodec.DefaultPublicKeyCodec{}

	return EncodingConfig{
		InterfaceRegistry: interfaceRegistry,
		Marshaler:         marshaler,
		TxDecoder:         signing.DefaultTxDecoder(marshaler, pubKeyCodec),
		TxGenerator:       signing.NewTxGenerator(marshaler, pubKeyCodec),
		Amino:             cdc,
	}
}

func MakeCodecs() (codec.Marshaler, codectypes.InterfaceRegistry, *codec.Codec) {
	cfg := MakeEncodingConfig()
	return cfg.Marshaler, cfg.InterfaceRegistry, cfg.Amino
}

func NewAnteHandler(ak auth.AccountKeeper, bk bank.Keeper, ibcK ibc.Keeper) types.AnteHandler {
	return ante.NewProtoAnteHandler(
		ak, bk, ibcK, ante.DefaultSigVerificationGasConsumer,
	)
}
