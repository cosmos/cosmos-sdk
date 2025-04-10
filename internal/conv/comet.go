package conv

import (
	"fmt"

	abci "github.com/cometbft/cometbft/abci/types"
	cmtprotocrypto "github.com/cometbft/cometbft/api/cometbft/crypto/v1"
)

// Convenience conversion from PublicKey and power to ABCI ValidatorUpdate
func ValidatorUpdateFromPublicKey(publicKey cmtprotocrypto.PublicKey, power int64) (abci.ValidatorUpdate, error) {
	var pkBytes []byte
	var pkType string

	switch publicKey.Sum.(type) {
	case *cmtprotocrypto.PublicKey_Ed25519:
		pkBytes = publicKey.GetEd25519()
		pkType = "tendermint/PubKeyEd25519"
	case *cmtprotocrypto.PublicKey_Secp256K1:
		pkBytes = publicKey.GetSecp256K1()
		pkType = "tendermint/PubKeySecp256k1"
	case *cmtprotocrypto.PublicKey_Bls12381:
		pkBytes = publicKey.GetBls12381()
		pkType = "tendermint/PubKeyBls12_381"
	default:
		return abci.ValidatorUpdate{}, fmt.Errorf("no known key is set")
	}

	return abci.ValidatorUpdate{
		PubKeyBytes: pkBytes,
		PubKeyType:  pkType,
		Power:       power,
	}, nil
}
