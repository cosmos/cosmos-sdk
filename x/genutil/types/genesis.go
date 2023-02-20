package types

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	cmtjson "github.com/cometbft/cometbft/libs/json"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	cmttypes "github.com/cometbft/cometbft/types"
	cmttime "github.com/cometbft/cometbft/types/time"

	"github.com/cosmos/cosmos-sdk/version"
)

const (
	// MaxChainIDLen is the maximum length of a chain ID.
	MaxChainIDLen = cmttypes.MaxChainIDLen
)

// AppGenesisOnly defines the app's genesis.
type AppGenesis struct {
	AppName       string          `json:"app_name"`
	AppVersion    string          `json:"app_version"`
	GenesisTime   time.Time       `json:"genesis_time"`
	ChainID       string          `json:"chain_id"`
	InitialHeight int64           `json:"initial_height"`
	AppHash       []byte          `json:"app_hash"`
	AppState      json.RawMessage `json:"app_state,omitempty"`

	// TODO eventually abstract from CometBFT types
	Validators      []cmttypes.GenesisValidator `json:"validators,omitempty"`
	ConsensusParams *cmtproto.ConsensusParams   `json:"consensus_params,omitempty"`
}

// NewAppGenesisWithVersion returns a new AppGenesis with the app name and app version already.
func NewAppGenesisWithVersion(chainID string, appState json.RawMessage) *AppGenesis {
	return &AppGenesis{
		AppName:    version.AppName,
		AppVersion: version.Version,
		ChainID:    chainID,
		AppState:   appState,
		Validators: nil,
	}
}

// ToCometBFTGenesisDoc converts the AppGenesis to a CometBFT GenesisDoc.
func (ag *AppGenesis) ToCometBFTGenesisDoc() (*cmttypes.GenesisDoc, error) {
	var consensusParams *cmttypes.ConsensusParams
	if ag.ConsensusParams != nil {
		consensusParams = &cmttypes.ConsensusParams{
			Block: cmttypes.BlockParams{
				MaxBytes: ag.ConsensusParams.Block.MaxBytes,
				MaxGas:   ag.ConsensusParams.Block.MaxGas,
			},
			Evidence: cmttypes.EvidenceParams{
				MaxAgeNumBlocks: ag.ConsensusParams.Evidence.MaxAgeNumBlocks,
				MaxAgeDuration:  ag.ConsensusParams.Evidence.MaxAgeDuration,
				MaxBytes:        ag.ConsensusParams.Evidence.MaxBytes,
			},
			Validator: cmttypes.ValidatorParams{
				PubKeyTypes: ag.ConsensusParams.Validator.PubKeyTypes,
			},
		}
	}

	return &cmttypes.GenesisDoc{
		ChainID:         ag.ChainID,
		InitialHeight:   ag.InitialHeight,
		AppHash:         ag.AppHash,
		AppState:        ag.AppState,
		ConsensusParams: consensusParams,
		Validators:      ag.Validators,
	}, nil
}

// SaveAs is a utility method for saving AppGenesis as a JSON file.
func (ag *AppGenesis) SaveAs(file string) error {
	appGenesisBytes, err := json.MarshalIndent(ag, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(file, appGenesisBytes, 0o600)
}

// ValidateAndComplete performs validation and completes the AppGenesis.
func (ag *AppGenesis) ValidateAndComplete() error {
	if ag.ChainID == "" {
		return errors.New("genesis doc must include non-empty chain_id")
	}

	if len(ag.ChainID) > MaxChainIDLen {
		return fmt.Errorf("chain_id in genesis doc is too long (max: %d)", MaxChainIDLen)
	}

	if ag.InitialHeight < 0 {
		return fmt.Errorf("initial_height cannot be negative (got %v)", ag.InitialHeight)
	}

	if ag.InitialHeight == 0 {
		ag.InitialHeight = 1
	}

	if ag.GenesisTime.IsZero() {
		ag.GenesisTime = cmttime.Now()
	}

	// verify that consesus and validators parameters are valid for CometBFT
	// TODO eventually generalize this for every consensus engine
	cmtGenesis, err := ag.ToCometBFTGenesisDoc()
	if err != nil {
		return err
	}

	if err := cmtGenesis.ValidateAndComplete(); err != nil {
		return err
	}

	ag.Validators = cmtGenesis.Validators
	consensusParams := cmtGenesis.ConsensusParams.ToProto()
	ag.ConsensusParams = &consensusParams

	return nil
}

// AppGenesisFromFile reads the AppGenesis from the provided file.
func AppGenesisFromFile(genFile string) (*AppGenesis, error) {
	jsonBlob, err := os.ReadFile(genFile) //nolint:gosec
	if err != nil {
		return nil, fmt.Errorf("couldn't read AppGenesis file (%s): %w", genFile, err)
	}

	var appGenesis AppGenesis
	if err := json.Unmarshal(jsonBlob, &appGenesis); err != nil {
		// fallback to CometBFT genesis
		var ctmGenesis cmttypes.GenesisDoc
		if err2 := cmtjson.Unmarshal(jsonBlob, &ctmGenesis); err2 != nil {
			return nil, fmt.Errorf("error unmarshalling AppGenesis at %s: %w and failed fallback to CometBFT GenDoc: %w", genFile, err, err2)
		}

		consensusParams := ctmGenesis.ConsensusParams.ToProto()
		appGenesis = AppGenesis{
			GenesisTime:     ctmGenesis.GenesisTime,
			ChainID:         ctmGenesis.ChainID,
			InitialHeight:   ctmGenesis.InitialHeight,
			AppHash:         ctmGenesis.AppHash,
			AppState:        ctmGenesis.AppState,
			Validators:      ctmGenesis.Validators,
			ConsensusParams: &consensusParams,
		}
	}

	return &appGenesis, nil
}
