package genutil

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	cfg "github.com/cometbft/cometbft/v2/config"
	cmtcrypto "github.com/cometbft/cometbft/v2/crypto"
	cmtbls12381 "github.com/cometbft/cometbft/v2/crypto/bls12381"
	tmed25519 "github.com/cometbft/cometbft/v2/crypto/ed25519"
	"github.com/cometbft/cometbft/v2/p2p"
	"github.com/cometbft/cometbft/v2/privval"
	cmttypes "github.com/cometbft/cometbft/v2/types"
	"github.com/cosmos/go-bip39"

	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
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

// ExportGenesisFileWithTime creates and writes the genesis configuration to disk.
// An error is returned if building or writing the configuration to file fails.
func ExportGenesisFileWithTime(genFile, chainID string, validators []cmttypes.GenesisValidator, appState json.RawMessage, genTime time.Time) error {
	appGenesis := types.NewAppGenesisWithVersion(chainID, appState)
	appGenesis.GenesisTime = genTime
	appGenesis.Consensus.Validators = validators

	if err := appGenesis.ValidateAndComplete(); err != nil {
		return err
	}

	return appGenesis.SaveAs(genFile)
}

// InitializeNodeValidatorFiles creates private validator and p2p configuration files. Key type is ed25519.
func InitializeNodeValidatorFiles(config *cfg.Config) (nodeID string, valPubKey cryptotypes.PubKey, err error) {
	return InitializeNodeValidatorFilesWithKeyType(config, ed25519.KeyType)
}

// InitializeNodeValidatorFilesFromMnemonic creates private validator and p2p configuration files using the given mnemonic.
// If no valid mnemonic is given, a random one will be used instead. Key type is ed25519.
func InitializeNodeValidatorFilesFromMnemonic(config *cfg.Config, mnemonic string) (nodeID string, valPubKey cryptotypes.PubKey, err error) {
	if len(mnemonic) > 0 && !bip39.IsMnemonicValid(mnemonic) {
		return "", nil, errors.New("invalid mnemonic")
	}
	nodeKey, err := p2p.LoadOrGenNodeKey(config.NodeKeyFile())
	if err != nil {
		return "", nil, err
	}

	nodeID = nodeKey.ID()

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
		// CometBFT uses the ed25519 key generator as default if the given generator function is nil.
		filePV, err = privval.LoadOrGenFilePV(pvKeyFile, pvStateFile, nil)
		if err != nil {
			return "", nil, err
		}
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

func InitializeNodeValidatorFilesWithKeyType(config *cfg.Config, keyType string) (nodeID string, valPubKey cryptotypes.PubKey, err error) {
	return InitializeNodeValidatorFilesFromMnemonicWithKeyType(config, "", keyType)
}

func InitializeNodeValidatorFilesFromMnemonicWithKeyType(config *cfg.Config, mnemonic, keyType string) (nodeID string, valPubKey cryptotypes.PubKey, err error) {
	if len(mnemonic) > 0 && !bip39.IsMnemonicValid(mnemonic) {
		return "", nil, errors.New("invalid mnemonic")
	}
	nodeKey, err := p2p.LoadOrGenNodeKey(config.NodeKeyFile())
	if err != nil {
		return "", nil, err
	}

	nodeID = nodeKey.ID()

	pvKeyFile := config.PrivValidatorKeyFile()
	if err := os.MkdirAll(filepath.Dir(pvKeyFile), 0o777); err != nil {
		return "", nil, fmt.Errorf("could not create directory %q: %w", filepath.Dir(pvKeyFile), err)
	}

	pvStateFile := config.PrivValidatorStateFile()
	if err := os.MkdirAll(filepath.Dir(pvStateFile), 0o777); err != nil {
		return "", nil, fmt.Errorf("could not create directory %q: %w", filepath.Dir(pvStateFile), err)
	}

	var (
		filePV  *privval.FilePV
		privKey cmtcrypto.PrivKey
		keyGenF func() (cmtcrypto.PrivKey, error)
	)

	if len(mnemonic) == 0 {

		switch keyType {
		case "ed25519":
			keyGenF = nil
			// filePV = loadOrGenFilePV(tmed25519.GenPrivKey(), pvKeyFile, pvStateFile)
		case "bls12_381":
			keyGenF = func() (cmtcrypto.PrivKey, error) {
				return cmtbls12381.GenPrivKey()
			}
		default:
			keyGenF = nil
		}

		// CometBFT uses the ed25519 key generator as default if the given generator function is nil.
		filePV, err = privval.LoadOrGenFilePV(pvKeyFile, pvStateFile, keyGenF)
		if err != nil {
			return "", nil, err
		}
	} else {
		switch keyType {
		case "ed25519":
			privKey = tmed25519.GenPrivKeyFromSecret([]byte(mnemonic))
		case "bls12_381":
			// TODO: need to add support for getting from mnemonic in Comet.
			return "", nil, errors.New("BLS key type does not support mnemonic")
		default:
			privKey = tmed25519.GenPrivKeyFromSecret([]byte(mnemonic))
		}
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
