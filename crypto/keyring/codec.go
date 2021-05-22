package keyring

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/legacy"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
)

func init() {
	RegisterLegacyAminoCodec(legacy.Cdc)
}

// RegisterLegacyAminoCodec registers concrete types and interfaces on the given codec.
// TODO how to remove Info entirely?
// rename to LegacyInfo
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterInterface((*Info)(nil), nil)
	cdc.RegisterConcrete(hd.BIP44Params{}, "crypto/keys/hd/BIP44Params", nil)
	cdc.RegisterConcrete(LocalInfo{}, "crypto/keys/LocalInfo", nil)
	cdc.RegisterConcrete(LedgerInfo{}, "crypto/keys/LedgerInfo", nil)
	cdc.RegisterConcrete(OfflineInfo{}, "crypto/keys/OfflineInfo", nil)
	cdc.RegisterConcrete(MultiInfo{}, "crypto/keys/MultiInfo", nil)
}
