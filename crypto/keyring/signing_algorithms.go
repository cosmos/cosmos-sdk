package keyring

// pubKeyType defines an algorithm to derive key-pairs which can be used for cryptographic signing.
type pubKeyType string

const (
	// MultiAlgo implies that a pubkey is a multisignature
	MultiAlgo = pubKeyType("multi")
	// Secp256k1 uses the Bitcoin secp256k1 ECDSA parameters.
	Secp256k1 = pubKeyType("secp256k1")
	// Ed25519 represents the Ed25519 signature system.
	// It is currently not supported for end-user keys (wallets/ledgers).
	Ed25519 = pubKeyType("ed25519")
	// Sr25519 represents the Sr25519 signature system.
	Sr25519 = pubKeyType("sr25519")
)

// IsSupportedAlgorithm returns whether the signing algorithm is in the passed-in list of supported algorithms.
func IsSupportedAlgorithm(supported []pubKeyType, algo pubKeyType) bool {
	for _, supportedAlgo := range supported {
		if algo == supportedAlgo {
			return true
		}
	}
	return false
}
