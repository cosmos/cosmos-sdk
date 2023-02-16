package types

import (
	"encoding/json"
	fmt "fmt"
	"os"
	"time"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	cmttypes "github.com/cometbft/cometbft/types"
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

// ToCometBFTGenesisDoc converts the AppGenesis to a CometBFT GenesisDoc.
func (ag AppGenesis) ToCometBFTGenesisDoc() (*cmttypes.GenesisDoc, error) {
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

	return os.WriteFile(file, appGenesisBytes, 0600)
}

// AppGenesisFromFile reads the AppGenesis from the provided file.
func AppGenesisFromFile(genFile string) (AppGenesis, error) {
	jsonBlob, err := os.ReadFile(genFile)
	if err != nil {
		return AppGenesis{}, fmt.Errorf("couldn't read AppGenesis file (%s): %w", genFile, err)
	}

	var appGenesis AppGenesis
	if err := json.Unmarshal(jsonBlob, &appGenesis); err != nil {
		return AppGenesis{}, fmt.Errorf("error unmarshalling AppGenesis at %s: %w", genFile, err)
	}

	return appGenesis, nil
}
