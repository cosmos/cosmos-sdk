// +build test_amino

package simapp

import (
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/std"
	"github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/ibc"
)

func MakeEncodingConfig() EncodingConfig {
	cdc := std.MakeCodec(ModuleBasics)
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	std.RegisterInterfaces(interfaceRegistry)
	ModuleBasics.RegisterInterfaceModules(interfaceRegistry)
	marshaler := codec.NewHybridCodec(cdc, interfaceRegistry)
	return EncodingConfig{
		InterfaceRegistry: interfaceRegistry,
		Marshaler:         marshaler,
		TxDecoder:         authtypes.DefaultTxDecoder(cdc),
		TxGenerator:       authtypes.StdTxGenerator{Cdc: cdc},
		Amino:             cdc,
	}
}

func MakeCodecs() (codec.Marshaler, codectypes.InterfaceRegistry, *codec.Codec) {
	cfg := MakeEncodingConfig()
	return cfg.Marshaler, cfg.InterfaceRegistry, cfg.Amino
}

func NewAnteHandler(ak auth.AccountKeeper, bk bank.Keeper, ibcK ibc.Keeper) types.AnteHandler {
	return ante.NewAnteHandler(
		ak, bk, ibcK, ante.DefaultSigVerificationGasConsumer,
	)
}
