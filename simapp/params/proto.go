// +build !test_amino

package params

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
)

func MakeEncodingConfig() EncodingConfig {
	cdc := codec.New()
	interfaceRegistry := types.NewInterfaceRegistry()
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
