package keys

import (
	tmcrypto "github.com/tendermint/tendermint/crypto"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/multisig"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
)

// RegisterInterfaces registers the sdk.Tx interface.
func RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	registry.RegisterInterface("tendermint.crypto.Pubkey", (*tmcrypto.PubKey)(nil))
	registry.RegisterImplementations((*tmcrypto.PubKey)(nil), &secp256k1.PubKey{})
	registry.RegisterImplementations((*tmcrypto.PubKey)(nil), &multisig.LegacyAminoPubKey{})

	registry.RegisterInterface("crypto.Pubkey", (*cryptotypes.PubKey)(nil))
	registry.RegisterImplementations((*cryptotypes.PubKey)(nil), &secp256k1.PubKey{})
	registry.RegisterImplementations((*cryptotypes.PubKey)(nil), &multisig.LegacyAminoPubKey{})
}
