package hd

import (
	"crypto/sha3"
	"errors"
	"fmt"
	"strings"

	"github.com/cosmos/go-bip39"
	"golang.org/x/text/unicode/norm"

	"github.com/cosmos/cosmos-sdk/crypto/keys/mldsa65"
	"github.com/cosmos/cosmos-sdk/crypto/types"
)

// MlDsa65 is the ML-DSA-65 (FIPS 204) account key algorithm.
var MlDsa65 = mldsa65Algo{}

// MlDsa65BIP85 uses the draft PQHD BIP85/SHAKE256 derivation for ML-DSA-65.
// It has the same signing algorithm and wire format as MlDsa65, but derives different keys.
// Recovery requires the original derivation mode, path, and BIP39 passphrase.
// The pre-v1 proposal must not be used to protect real assets.
var MlDsa65BIP85 = mldsa65BIP85Algo{}

// PQHD draft framework, family-block registry, and vectors:
// https://github.com/pqdh/pqhd/tree/befdce14cecbb5d485d3239982bfa5b207cdb922
// Application 505153 ("PQS") is not a registered BIP85 assignment. Scheme 201
// is active ML-DSA-65 in family block 2; no other PQHD scheme is supported here.
const mlDsa65BIP85Prefix = "m/83696968'/505153'"

var (
	// ErrMlDsa65BIP85Mnemonic is returned when a PQHD ML-DSA-65 key is derived
	// from a mnemonic that is not 24 words long.
	ErrMlDsa65BIP85Mnemonic = errors.New("BIP85 ML-DSA-65 requires a 24-word mnemonic")
	// ErrMlDsa65BIP85Path is returned when a PQHD ML-DSA-65 key is derived on a
	// path outside the PQHD ML-DSA-65 namespace.
	ErrMlDsa65BIP85Path = errors.New("ML-DSA-65 BIP85 path must be " + mlDsa65BIP85Prefix + "/coin_type'/201'/account'/index'")
)

// CreateMlDsa65BIP85Path returns the fully hardened PQHD ML-DSA-65 entropy
// path. The SLIP-44 coin type, account, and address index must be below 2^31.
func CreateMlDsa65BIP85Path(coinType, account, index uint32) (string, error) {
	if coinType >= 1<<31 || account >= 1<<31 || index >= 1<<31 {
		return "", fmt.Errorf("BIP85 coin type, account, and index must be less than 2147483648")
	}
	return fmt.Sprintf("%s/%d'/201'/%d'/%d'", mlDsa65BIP85Prefix, coinType, account, index), nil
}

type mldsa65BIP85Algo struct{ mldsa65Algo }

// Derive implements BIP39 -> hardened BIP32 -> BIP85 -> SHAKE256 -> 32-byte xi.
// Only valid 24-word English mnemonics are accepted. An empty path selects
// coin type 118, account 0, index 0; explicit paths must use PQHD's scoped
// ML-DSA-65 namespace.
func (mldsa65BIP85Algo) Derive() DeriveFn {
	return func(mnemonic, bip39Passphrase, hdPath string) ([]byte, error) {
		mnemonic = strings.Join(strings.Fields(norm.NFKD.String(mnemonic)), " ")
		if len(strings.Fields(mnemonic)) != 24 {
			return nil, ErrMlDsa65BIP85Mnemonic
		}
		if hdPath == "" {
			hdPath, _ = CreateMlDsa65BIP85Path(118, 0, 0)
		}
		parts := strings.Split(hdPath, "/")
		if !strings.HasPrefix(hdPath, mlDsa65BIP85Prefix+"/") || len(parts) != 7 || parts[4] != "201'" {
			return nil, ErrMlDsa65BIP85Path
		}
		seed, err := bip39.NewSeedWithErrorChecking(mnemonic, norm.NFKD.String(bip39Passphrase))
		if err != nil {
			return nil, err
		}
		defer clear(seed)
		master, chainCode := ComputeMastersFromSeed(seed)
		defer clear(master[:])
		defer clear(chainCode[:])
		entropy, err := DeriveBIP85Entropy(master, chainCode, hdPath, 64)
		if err != nil {
			return nil, err
		}
		defer clear(entropy)
		// PQHD expands the full HMAC output, even when xi is shorter than it.
		return sha3.SumSHAKE256(entropy, 32), nil
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
