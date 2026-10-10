package hd

import (
	"fmt"
	"strings"

	"github.com/cosmos/go-bip39"
	"golang.org/x/text/unicode/norm"

	"github.com/cosmos/cosmos-sdk/crypto/keys/mldsa65"
	"github.com/cosmos/cosmos-sdk/crypto/types"
)

// MlDsa65 is the ML-DSA-65 (FIPS 204) account key algorithm.
var MlDsa65 = mldsa65Algo{}

// MlDsa65BIP85 uses BIP85 entropy for ML-DSA-65 key generation. It has the same
// signing algorithm and wire format as MlDsa65, but derives different keys.
// Recovery requires the original derivation mode, path, and BIP39 passphrase.
var MlDsa65BIP85 = mldsa65BIP85Algo{}

// This is an SDK-local convention, not a registered BIP85 application number
// or a Project Eleven algorithm ID: 118 identifies Cosmos, 65 ML-DSA-65.
const mlDsa65BIP85Prefix = "m/83696968'/118'/65'"

// CreateMlDsa65BIP85Path returns the fully hardened Cosmos ML-DSA-65 entropy
// path. Account and address indices must both be below 2^31.
func CreateMlDsa65BIP85Path(account, index uint32) (string, error) {
	if account >= 1<<31 || index >= 1<<31 {
		return "", fmt.Errorf("BIP85 account and index must be less than 2147483648")
	}
	return fmt.Sprintf("%s/%d'/%d'", mlDsa65BIP85Prefix, account, index), nil
}

type mldsa65BIP85Algo struct{ mldsa65Algo }

// Derive implements BIP39 -> hardened BIP32 -> BIP85 -> 32-byte keygen seed.
// Only valid 24-word English mnemonics are accepted. An empty path selects
// account 0, index 0; explicit paths must stay in the Cosmos ML-DSA-65 subtree.
func (mldsa65BIP85Algo) Derive() DeriveFn {
	return func(mnemonic, bip39Passphrase, hdPath string) ([]byte, error) {
		mnemonic = strings.Join(strings.Fields(norm.NFKD.String(mnemonic)), " ")
		if len(strings.Fields(mnemonic)) != 24 {
			return nil, fmt.Errorf("BIP85 ML-DSA-65 requires a 24-word mnemonic")
		}
		if hdPath == "" {
			hdPath, _ = CreateMlDsa65BIP85Path(0, 0)
		}
		if !strings.HasPrefix(hdPath, mlDsa65BIP85Prefix+"/") || len(strings.Split(hdPath, "/")) != 6 {
			return nil, fmt.Errorf("ML-DSA-65 BIP85 path must be %s/account'/index'", mlDsa65BIP85Prefix)
		}
		seed, err := bip39.NewSeedWithErrorChecking(mnemonic, norm.NFKD.String(bip39Passphrase))
		if err != nil {
			return nil, err
		}
		defer clear(seed)
		master, chainCode := ComputeMastersFromSeed(seed)
		defer clear(master[:])
		defer clear(chainCode[:])
		return DeriveBIP85Entropy(master, chainCode, hdPath, 32)
	}
}

type mldsa65Algo struct{}

func (mldsa65Algo) Name() PubKeyType {
	return MlDsa65Type
}

// Derive reuses the secp256k1 BIP32 derivation. It returns the 32-byte
// BIP32-derived key for the given mnemonic, passphrase, and HD path, which is
// used directly as the ML-DSA-65 keygen seed (mldsa65 SeedSize == 32). This
// gives mnemonic recovery and per-path account separation without inventing a
// new derivation scheme.
func (mldsa65Algo) Derive() DeriveFn {
	return Secp256k1.Derive()
}

// Generate builds an ML-DSA-65 private key from the 32-byte derived seed.
func (mldsa65Algo) Generate() GenerateFn {
	return func(bz []byte) types.PrivKey {
		privKey, err := mldsa65.GenPrivKeyFromSeed(bz)
		if err != nil {
			// A non-32-byte seed only reaches here from a caller passing
			// untrusted/dummy bytes directly (e.g. ImportPrivKeyHex); the BIP32
			// derivation path always supplies a 32-byte seed. Such callers must
			// guard their input — see keyring.generatePrivKey.
			panic(err)
		}
		return &privKey
	}
}
