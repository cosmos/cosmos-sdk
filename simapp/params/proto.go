// +build test_proto

// TODO switch to !test_amino build flag once proto Tx's are ready
package params

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/std"
	"github.com/cosmos/cosmos-sdk/x/auth/tx"
)

// MakeEncodingConfig creates an EncodingConfig for an amino based test configuration.
func MakeEncodingConfig() EncodingConfig {
	cdc := codec.New()
	interfaceRegistry := types.NewInterfaceRegistry()
	protoCodec := codec.NewProtoCodec(interfaceRegistry)
	txGen := tx.NewTxConfig(protoCodec, std.DefaultPublicKeyCodec{}, tx.DefaultSignModeHandler())

	return EncodingConfig{
		InterfaceRegistry: interfaceRegistry,
		Marshaler:         protoCodec,
		TxConfig:          txGen,
		Amino:             cdc,
	}
}
