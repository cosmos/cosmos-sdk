package export

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	app "github.com/cosmos/cosmos-sdk/cmd/gaia/app"
	"github.com/cosmos/cosmos-sdk/codec"
	tmtypes "github.com/tendermint/tendermint/types"
)

// GenesisFile defines the Gaia genesis format
type GenesisFile struct {
	GenesisTime     string                   `json:"genesis_time"`
	ChainID         string                   `json:"chain_id"`
	ConsensusParams *tmtypes.ConsensusParams `json:"consensus_params"`
	AppHash         string                   `json:"app_hash"`
	AppState        app.GenesisState         `json:"app_state"`
}

// NewGenesisFile builds a default GenesisDoc and creates a GenesisFile from it
func NewGenesisFile(cdc *codec.Codec, path string) (GenesisFile, error) {

	genDoc, err := importGenesis(path)
	if err != nil {
		return GenesisFile{}, err
	}

	var appState app.GenesisState
	if genDoc.AppState == nil {
		appState = app.GenesisState{}
	} else {
		if err = cdc.UnmarshalJSON(genDoc.AppState, &appState); err != nil {
			return GenesisFile{}, err
		}
	}

	return GenesisFile{
		GenesisTime:     genDoc.GenesisTime.String(),
		ChainID:         genDoc.ChainID,
		ConsensusParams: genDoc.ConsensusParams,
		AppHash:         genDoc.AppHash.String(),
		AppState:        appState,
	}, nil
}

// validateBasic validates each of the arguments passed to the script
func ValidateBasic(path, genesisTime string) error {
	if path == "" {
		return fmt.Errorf("path to genesis file required")
	}

	if genesisTime == "" {
		return fmt.Errorf("genesis start time required")
	}

	_, err := time.Parse(time.RFC3339, genesisTime)
	if err != nil {
		return err
	}

	if ext := filepath.Ext(path); ext != ".json" {
		return fmt.Errorf("%s is not a JSON file", path)
	}

	if _, err = os.Stat(path); err != nil {
		return err
	}
	return nil
}

// importGenesis imports genesis from JSON and completes missing fields
func importGenesis(path string) (genDoc *tmtypes.GenesisDoc, err error) {
	genDoc, err = tmtypes.GenesisDocFromFile(path)
	if err != nil {
		return
	}

	err = genDoc.ValidateAndComplete()
	if err != nil {
		return
	}
	return
}
