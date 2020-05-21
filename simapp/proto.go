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

func SimappTxDecoder() types.TxDecoder {
	cdc, _, _ := MakeCodecs()
	return signing.DefaultTxDecoder(cdc, cryptocodec.DefaultPublicKeyCodec{})
}

// MakeCodecs constructs the *std.Codec and *codec.Codec instances used by
// simapp. It is useful for tests and clients who do not want to construct the
// full simapp
func MakeCodecs() (codec.Marshaler, codectypes.InterfaceRegistry, *codec.Codec) {
	cdc := std.MakeCodec(ModuleBasics)
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	std.RegisterInterfaces(interfaceRegistry)
	ModuleBasics.RegisterInterfaceModules(interfaceRegistry)
	appCodec := codec.NewHybridCodec(cdc, interfaceRegistry)
	return appCodec, nil, cdc
}

func NewAnteHandler(ak auth.AccountKeeper, bk bank.Keeper, ibcK ibc.Keeper) types.AnteHandler {
	return ante.NewProtoAnteHandler(
		ak, bk, ibcK, ante.DefaultSigVerificationGasConsumer,
	)
}
