package legacytx

import "cosmossdk.io/core/registry"

func RegisterLegacyAminoCodec(cdc registry.AminoRegistrar) {
	cdc.RegisterConcrete(StdTx{}, "cosmos-sdk/StdTx")
}
