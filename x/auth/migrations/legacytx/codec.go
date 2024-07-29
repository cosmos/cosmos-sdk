package legacytx

import (
	"cosmossdk.io/core/legacy"
)

func RegisterLegacyAminoCodec(cdc legacy.Amino) {
	cdc.RegisterConcrete(StdTx{}, "cosmos-sdk/StdTx")
}
