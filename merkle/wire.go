package merkle

import (
	"github.com/cosmos/cosmos-sdk/wire"
)

// RegisterWire registers interfaces to the codec
func RegisterWire(cdc *wire.Codec) {
	cdc.RegisterInterface((*Op)(nil), nil)
	//	cdc.RegisterConcrete(TMCoreOp{}, "cosmos-sdk/TMCoreOp", nil)
	cdc.RegisterConcrete(IAVLExistsOp{}, "cosmos-sdk/IAVLExistsOp", nil)
	cdc.RegisterConcrete(SimpleExistsOp{}, "cosmos-sdk/SimpleExistsOp", nil)
}
