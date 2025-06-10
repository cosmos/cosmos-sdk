package keys

import (
	"github.com/cometbft/cometbft/v2/crypto/ed25519"
	"github.com/cometbft/cometbft/v2/crypto/encoding"
	"github.com/cometbft/cometbft/v2/crypto/secp256k1"

	cosmosed25519 "github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	cosmossecp256k1 "github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/crypto/types"
)

// PubKeyFromCometTypeAndBytes builds a crypto.PubKey from the given comet/v2 type and bytes.
// It returns ErrUnsupportedKey if the pubkey type is unsupported or
// ErrInvalidKeyLen if the key length is invalid.
func PubKeyFromCometTypeAndBytes(pkType string, bytes []byte) (types.PubKey, error) {
	var pubKey types.PubKey
	switch pkType {
	case ed25519.KeyType:
		if len(bytes) != ed25519.PubKeySize {
			return nil, encoding.ErrInvalidKeyLen{
				Key:  pkType,
				Got:  len(bytes),
				Want: ed25519.PubKeySize,
			}
		}

		pk := make([]byte, ed25519.PubKeySize)
		copy(pk, bytes)
		pubKey = &cosmosed25519.PubKey{Key: pk}

	case secp256k1.KeyType:
		if len(bytes) != secp256k1.PubKeySize {
			return nil, encoding.ErrInvalidKeyLen{
				Key:  pkType,
				Got:  len(bytes),
				Want: secp256k1.PubKeySize,
			}
		}

		pk := make([]byte, secp256k1.PubKeySize)
		copy(pk, bytes)
		pubKey = &cosmossecp256k1.PubKey{Key: pk}
	default:
		return nil, encoding.ErrUnsupportedKey{KeyType: pkType}
	}
	return pubKey, nil
}
