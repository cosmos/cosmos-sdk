package std

import (
	tmjson "github.com/tendermint/tendermint/libs/json"
	tmcrypto "github.com/tendermint/tendermint/proto/tendermint/crypto"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	txtypes "github.com/cosmos/cosmos-sdk/types/tx"
	vesting "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
)

// Temporary to solve https://github.com/cosmos/cosmos-sdk/pull/7442
func init() {
	tmjson.RegisterType((*tmcrypto.PublicKey)(nil), "tendermint.crypto.PublicKey")
	tmjson.RegisterType((*tmcrypto.PublicKey_Ed25519)(nil), "tendermint.crypto.PublicKey_Ed25519")
}

// RegisterLegacyAminoCodec registers types with the Amino codec.
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	sdk.RegisterLegacyAminoCodec(cdc)
	cryptocodec.RegisterCrypto(cdc)
	vesting.RegisterLegacyAminoCodec(cdc)
}

// RegisterInterfaces registers Interfaces from sdk/types, vesting, crypto, tx.
func RegisterInterfaces(interfaceRegistry types.InterfaceRegistry) {
	sdk.RegisterInterfaces(interfaceRegistry)
	txtypes.RegisterInterfaces(interfaceRegistry)
	cryptocodec.RegisterInterfaces(interfaceRegistry)
	vesting.RegisterInterfaces(interfaceRegistry)
}
