package app

import (
	"strings"

	"github.com/cosmos/cosmos-sdk/codec"
	tmtypes "github.com/tendermint/tendermint/types"
)

// GenesisFile defines the Gaia genesis format
type GenesisFile struct {
	GenesisTime     string                   `json:"genesis_time"`
	ChainID         string                   `json:"chain_id"`
	ConsensusParams *tmtypes.ConsensusParams `json:"consensus_params"`
	AppHash         string                   `json:"app_hash"`
	AppState        GenesisState             `json:"app_state"`
}

// NewGenFileFromTmGenDoc unmarshals a tendermint GenesisDoc and creates a GenesisFile from it
func NewGenFileFromTmGenDoc(cdc *codec.Codec, genDoc *tmtypes.GenesisDoc) (GenesisFile, error) {

	var appState GenesisState
	err := cdc.UnmarshalJSON(genDoc.AppState, &appState)
	if err != nil {
		return GenesisFile{}, err
	}

	return GenesisFile{
		GenesisTime:     genDoc.GenesisTime.String(),
		ChainID:         genDoc.ChainID,
		ConsensusParams: genDoc.ConsensusParams,
		AppHash:         genDoc.AppHash.String(),
		AppState:        appState,
	}, nil
}

// applyLatestChanges for software upgrade. Nees to be updated on every state export.
func (genFile GenesisFile) applyLatestChanges() GenesisFile {
	// proposal #1 updates
	genFile.AppState.MintData.Params.BlocksPerYear = 4855015

	// proposal #2 updates
	genFile.ConsensusParams.Block.MaxGas = 200000
	genFile.ConsensusParams.Block.MaxBytes = 2000000

	// enable transfers
	genFile.AppState.BankData.SendEnabled = true
	genFile.AppState.DistrData.WithdrawAddrEnabled = true

	return genFile
}

// GetUpdatedGenesis creates a tendermint genesis doc, updates latests release
// changes and validates correctness of the new genesis state
func GetUpdatedGenesis(
	cdc *codec.Codec, oldGenesisPath, chainID, startTime string,
) (GenesisFile, error) {

	genDoc, err := tmtypes.GenesisDocFromFile(oldGenesisPath)
	if err != nil {
		return GenesisFile{}, err
	}

	err = genDoc.ValidateAndComplete()
	if err != nil {
		return GenesisFile{}, err
	}

	genesis, err := NewGenFileFromTmGenDoc(cdc, genDoc)
	if err != nil {
		return GenesisFile{}, err
	}

	genesis.ChainID = strings.Trim(chainID, " ")
	genesis.GenesisTime = startTime

	genesis = genesis.applyLatestChanges()

	err = GaiaValidateGenesisState(genesis.AppState)
	if err != nil {
		return GenesisFile{}, err
	}

	return genesis, nil
}
