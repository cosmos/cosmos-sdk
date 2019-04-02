package app


import (
	"time"
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/cosmos/cosmos-sdk/codec"
	tmtypes "github.com/tendermint/tendermint/types"
)

type GenesisFile struct {
	GenesisTime time.Time `json:"genesis_time"`
	ChainID string `json:"chain_id"` 
	ConsensusParams tmtypes.ConsensusParams `json:"consensus_params"`
	AppHash string `json:"app_hash"`
	AppState GenesisState `json:"app_state"`
}

// NewGenFileFromTmGenDoc unmarshals a tendermint GenesisDoc and creates a GenesisFile from it
func NewGenFileFromTmGenDoc(genDoc tmtypes.GenesisDoc) (GenesisFile, error){
	
	var appState GenesisState
	err := json.Unmarshal(genDoc.AppHash, &appState)
	if err != nil {
		return nil, err
	} 
	
	return GenesisFile{
		GenesisTime: genDoc.GenesisTime,
		ChainID: genDoc.ChainID,
		ConsensusParams: genDoc.ConsensusParams,
		AppHash: genDoc.AppHash.String(),
		AppState: appState
	}, nil
}

func (genFile *GenesisFile) applyLatestChanges() {
		// proposal #1 updates
		*genFile.AppState.MintData.Params.BlocksPerYear = 4855015

		// proposal #2 updates
		*genFile.ConsensusParams.Block.BlockMaxGas = 200000
		*genFile.ConsensusParams.Block.BlockMaxBytes = 2000000
	
		// enable transfers
		*genFile.AppState.BankData.SendEnabled = true
		*genFile.AppState.DistrData.WithdrawAddrEnabled = true
}

func GetUpdatedGenesis(cdc *codec.Codec, oldGenFilename, chainID string, startTime time.Time)
(GenesisFile, error) {
	genDoc, err := tmtypes.GenesisDocFromJSON(genDocPath)
	if err != nil {
		return err
	}

	err = genDoc.ValidateAndComplete()
	if err != nil {
		return err
	}

	genesis, err := NewGenFileFromTmGenDoc(genDoc)
	if err != nil {
		return err
	}

	genesis.ChainID = strings.Trim(chainID)
	genesis.GenesisTime = startTime

	genesis.applyLatestChanges()

	err = GaiaValidateGenesisState(genesis.AppState)
	if err != nil {
		return err
	}

	

}