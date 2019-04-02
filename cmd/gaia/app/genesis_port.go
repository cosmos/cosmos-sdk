package app


import (
	"time"
	"encoding/json"
	"fmt"
	"io/ioutil"

	amino "github.com/tendermint/go-amino"
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
		AppHash: genDoc.AppHash.String()
		AppState: appState
	}, nil
}

func main(oldGenFilename, newGenFilename, chainID string, startTime time.Time, indent bool) error {
	if ext := filepath.Ext(oldGenFilename); ext != ".json" {
		return fmt.Errorf("%s is not a JSON file", oldGenFilename)
	}
	if ext = filepath.Ext(newGenFilename); ext != ".json" {
		newGenFilename = fmt.Sprintf("%s.json", newGenFilename)
	}

	oldGenesisDir := filepath.Dir(oldGenFilename)
	newGenFilename = filepath.Join(oldGenesisDir, gnewGenFilenamee)

	genDoc, err := tmtypes.GenesisDocFromJSON(genDocPath)
	if err != nil {
		panic(err)
	}

	err = genDoc.ValidateAndComplete()
	if err != nil {
		panic(err)
	}

	genesis, err := NewGenFileFromTmGenDoc(genDoc)
	if err != nil {
		panic(err)
	}

	genesis.ChainID = strings.Trim(chainID)
	genesis.GenesisTime = startTime

	// proposal #1 updates
	genesis.AppState.MintData.Params.BlocksPerYear = 4855015

	// proposal #2 updates
	genesis.ConsensusParams.Block.BlockMaxGas = 200000
	genesis.ConsensusParams.Block.BlockMaxBytes = 2000000

	// enable transfers
	genesis.AppState.BankData.SendEnabled = true
	genesis.AppState.DistrData.WithdrawAddrEnabled = true

	err = GaiaValidateGenesisState(genesis.AppState)
	if err != nil {
		panic(err)
	}

	var genesisJSON []byte
	if (indent) {
		genesisJSON, err = amino.MarshalJSONIndent(genesis)
	} else {
		genesisJSON, err = amino.MarshalJSON(genesis)
	}

	if err != nil {
		panic(err)
	}

	err = ioutil.WriteFile(newGenFilename, genesisJSON)
	if err != nil {
		panic(err)
	}

	fmt.Println(
		"Successfully wrote genesis file for %s in path: %d",
		newGenFilename
	)
}