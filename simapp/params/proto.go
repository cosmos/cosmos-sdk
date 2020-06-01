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
	txGen := signing.NewTxGenerator(marshaler, pubKeyCodec, signing.DefaultSignModeHandler())

	return EncodingConfig{
		InterfaceRegistry: interfaceRegistry,
		Marshaler:         marshaler,
		TxGenerator:       txGen,
		Amino:             cdc,
	}
}
