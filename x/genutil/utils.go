package genutil

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"time"

	cfg "github.com/cometbft/cometbft/config"
	cmtcrypto "github.com/cometbft/cometbft/crypto"
	"github.com/cometbft/cometbft/crypto/bls12381"
	tmed25519 "github.com/cometbft/cometbft/crypto/ed25519"
	"github.com/cometbft/cometbft/crypto/mldsa65"
	"github.com/cometbft/cometbft/crypto/secp256k1"
	"github.com/cometbft/cometbft/p2p"
	"github.com/cometbft/cometbft/privval"
	"github.com/cosmos/go-bip39"

	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/x/genutil/types"
)

// ExportGenesisFile creates and writes the genesis configuration to disk. An
// error is returned if building or writing the configuration to file fails.
func ExportGenesisFile(genesis *types.AppGenesis, genFile string) error {
	if err := genesis.ValidateAndComplete(); err != nil {
		return err
	}

	return genesis.SaveAs(genFile)
}

// ExportGenesisFileWithTime overwrites the genesis file at genFile with
// appGenesis, applying genTime as the genesis time. All other fields on
// appGenesis (validators, consensus params, app state, app hash, initial
// height, etc.) are preserved as the caller set them.
func ExportGenesisFileWithTime(genFile string, appGenesis *types.AppGenesis, genTime time.Time) error {
	appGenesis.GenesisTime = genTime
	return ExportGenesisFile(appGenesis, genFile)
}

// InitializeNodeValidatorFiles creates private validator and p2p configuration files.
func InitializeNodeValidatorFiles(config *cfg.Config) (nodeID string, valPubKey cryptotypes.PubKey, err error) {
	return InitializeNodeValidatorFilesFromMnemonic(config, "")
}

// InitializeNodeValidatorFilesFromMnemonic creates private validator and p2p configuration files using the given mnemonic.
// If no valid mnemonic is given, a random one will be used instead.
func InitializeNodeValidatorFilesFromMnemonic(config *cfg.Config, mnemonic string) (nodeID string, valPubKey cryptotypes.PubKey, err error) {
	if len(mnemonic) > 0 && !bip39.IsMnemonicValid(mnemonic) {
		return "", nil, fmt.Errorf("invalid mnemonic")
	}
	nodeKey, err := p2p.LoadOrGenNodeKey(config.NodeKeyFile())
	if err != nil {
		return "", nil, err
	}

	nodeID = string(nodeKey.ID())

	pvKeyFile := config.PrivValidatorKeyFile()
	if err := os.MkdirAll(filepath.Dir(pvKeyFile), 0o777); err != nil {
		return "", nil, fmt.Errorf("could not create directory %q: %w", filepath.Dir(pvKeyFile), err)
	}

	pvStateFile := config.PrivValidatorStateFile()
	if err := os.MkdirAll(filepath.Dir(pvStateFile), 0o777); err != nil {
		return "", nil, fmt.Errorf("could not create directory %q: %w", filepath.Dir(pvStateFile), err)
	}

	var filePV *privval.FilePV
	if len(mnemonic) == 0 {
		filePV = privval.LoadOrGenFilePV(pvKeyFile, pvStateFile)
	} else {
		privKey := tmed25519.GenPrivKeyFromSecret([]byte(mnemonic))
		filePV = privval.NewFilePV(privKey, pvKeyFile, pvStateFile)
		filePV.Save()
	}

	tmValPubKey, err := filePV.GetPubKey()
	if err != nil {
		return "", nil, err
	}

	valPubKey, err = cryptocodec.FromCmtPubKeyInterface(tmValPubKey)
	if err != nil {
		return "", nil, err
	}

	return nodeID, valPubKey, nil
}

// InitializeNodeValidatorFilesFromMnemonicWithKeyType is a sibling of
// InitializeNodeValidatorFilesFromMnemonic that lets callers choose the
// validator consensus-key algorithm. Recognized keyType values are
// "ed25519", "secp256k1", "bls12_381", "ml_dsa_65", and "" (which is
// treated as ed25519, delegating to the original function so behavior is
// byte-for-byte unchanged for callers that don't opt in).
//
// This function exists separately from InitializeNodeValidatorFilesFromMnemonic
// so the latter remains backwards compatible. Use this from test harnesses
// (e.g. testutil/network) that need to spin up validators backed by a
// non-default signature scheme.
func InitializeNodeValidatorFilesFromMnemonicWithKeyType(
	config *cfg.Config, mnemonic, keyType string,
) (nodeID string, valPubKey cryptotypes.PubKey, err error) {
	if keyType == "" || keyType == tmed25519.KeyType {
		return InitializeNodeValidatorFilesFromMnemonic(config, mnemonic)
	}
	if len(mnemonic) > 0 && !bip39.IsMnemonicValid(mnemonic) {
		return "", nil, fmt.Errorf("invalid mnemonic")
	}

	nodeKey, err := p2p.LoadOrGenNodeKey(config.NodeKeyFile())
	if err != nil {
		return "", nil, err
	}
	nodeID = string(nodeKey.ID())

	pvKeyFile := config.PrivValidatorKeyFile()
	if err := os.MkdirAll(filepath.Dir(pvKeyFile), 0o777); err != nil {
		return "", nil, fmt.Errorf("could not create directory %q: %w", filepath.Dir(pvKeyFile), err)
	}
	pvStateFile := config.PrivValidatorStateFile()
	if err := os.MkdirAll(filepath.Dir(pvStateFile), 0o777); err != nil {
		return "", nil, fmt.Errorf("could not create directory %q: %w", filepath.Dir(pvStateFile), err)
	}

	privKey, err := generateValidatorPrivKey(keyType, mnemonic)
	if err != nil {
		return "", nil, err
	}
	filePV := privval.NewFilePV(privKey, pvKeyFile, pvStateFile)
	filePV.Save()

	tmValPubKey, err := filePV.GetPubKey()
	if err != nil {
		return "", nil, err
	}

	valPubKey, err = cryptocodec.FromCmtPubKeyInterface(tmValPubKey)
	if err != nil {
		return "", nil, err
	}
	return nodeID, valPubKey, nil
}

// generateValidatorPrivKey produces a consensus private key of the requested
// type.
//
// Input contract:
//   - mnemonic == "": the key is sampled from OS randomness.
//   - mnemonic != "": SHA-256(mnemonic) is used as the 32-byte seed for every
//     supported scheme, so two calls with the same (keyType, mnemonic) always
//     produce the same key, and two calls with the same mnemonic but
//     different keyType derive from the same underlying seed bytes.
//
// Note: BLS12-381's underlying GenPrivKeyFromSecret hashes its input
// internally when the input is not exactly 32 bytes; we pass the 32-byte
// SHA-256 digest directly so it's used as-is. This keeps the
// mnemonic -> seed mapping uniform across schemes.
func generateValidatorPrivKey(keyType, mnemonic string) (cmtcrypto.PrivKey, error) {
	var seed []byte
	if mnemonic != "" {
		h := sha256.Sum256([]byte(mnemonic))
		seed = h[:]
	}

	switch keyType {
	case secp256k1.KeyType:
		if seed == nil {
			return secp256k1.GenPrivKey(), nil
		}
		return secp256k1.GenPrivKeySecp256k1(seed), nil

	case bls12381.KeyType:
		if seed == nil {
			pk, err := bls12381.GenPrivKey()
			if err != nil {
				return nil, fmt.Errorf("bls12_381 GenPrivKey: %w", err)
			}
			return pk, nil
		}
		pk, err := bls12381.GenPrivKeyFromSecret(seed)
		if err != nil {
			return nil, fmt.Errorf("bls12_381 GenPrivKeyFromSecret: %w", err)
		}
		return pk, nil

	case mldsa65.KeyType:
		if seed == nil {
			pk, err := mldsa65.GenPrivKey()
			if err != nil {
				return nil, fmt.Errorf("ml_dsa_65 GenPrivKey: %w", err)
			}
			return pk, nil
		}
		pk, err := mldsa65.GenPrivKeyFromSeed(seed)
		if err != nil {
			return nil, fmt.Errorf("ml_dsa_65 GenPrivKeyFromSeed: %w", err)
		}
		return pk, nil

	default:
		return nil, fmt.Errorf("unsupported validator consensus key type %q", keyType)
	}
}
