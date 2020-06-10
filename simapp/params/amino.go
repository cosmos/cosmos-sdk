package params

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

// MakeEncodingConfig creates an EncodingConfig for an amino based test configuration.
//
// TODO: this file should add a "+build test_amino" flag for #6190 and a proto.go file with a protobuf configuration
func MakeEncodingConfig() EncodingConfig {
	cdc := codec.New()
	interfaceRegistry := types.NewInterfaceRegistry()
	marshaler := codec.NewHybridCodec(cdc, interfaceRegistry)

	return EncodingConfig{
		InterfaceRegistry: interfaceRegistry,
		Marshaler:         marshaler,
		TxGenerator:       authtypes.StdTxGenerator{Cdc: cdc},
		Amino:             cdc,
	}
}
