// +build test_amino

package params

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

// MakeEncodingConfig creates an EncodingConfig for an amino based test configuration.
func MakeEncodingConfig() EncodingConfig {
	cdc := codec.New()
	interfaceRegistry := types.NewInterfaceRegistry()
	// TODO: switch to using AminoCodec here once amino compatibility is fixed
	marshaler := codec.NewHybridCodec(cdc, interfaceRegistry)

	return EncodingConfig{
		InterfaceRegistry: interfaceRegistry,
		Marshaler:         marshaler,
		TxConfig:          authtypes.StdTxConfig{Cdc: cdc},
		Amino:             cdc,
	}
}
