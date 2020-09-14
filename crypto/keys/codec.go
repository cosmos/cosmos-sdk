package keys

import (
	"github.com/tendermint/tendermint/crypto"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
)

// RegisterInterfaces registers the sdk.Tx interface.
func RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	registry.RegisterInterface("crypto.Pubkey", (*crypto.PubKey)(nil))
	registry.RegisterImplementations((*crypto.PubKey)(nil), &secp256k1.PubKey{})
}
