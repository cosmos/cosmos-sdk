package migration

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	cmtjson "github.com/cometbft/cometbft/libs/json"
	cmttypes "github.com/cometbft/cometbft/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/genutil/types"
)

type Migrator struct {
	// genesis file path
	filePath string
	reader   io.Reader
}

type legacyAppGenesis struct {
	AppName       string                  `json:"app_name"`
	AppVersion    string                  `json:"app_version"`
	GenesisTime   time.Time               `json:"genesis_time"`
	ChainID       string                  `json:"chain_id"`
	InitialHeight int64                   `json:"initial_height"`
	AppHash       []byte                  `json:"app_hash"`
	AppState      json.RawMessage         `json:"app_state,omitempty"`
	Consensus     *legacyConsensusGenesis `json:"consensus,omitempty"`
}

type legacyConsensusGenesis struct {
	Validators []cmttypes.GenesisValidator `json:"validators,omitempty"`
	Params     *cmttypes.ConsensusParams   `json:"params,omitempty"`
}

// NewMigrator takes in 2 file path one for the current genesis file
// and the other are the directory where the new genesis file will live.
// If you want to replace old genesis file the both path could be the same.
func NewMigrator(filePath, savePath string) (*Migrator, error) {
	file, err := os.Open(filepath.Clean(filePath))
	if err != nil {
		return nil, err
	}

	return &Migrator{
		filePath: savePath,
		reader:   file,
	}, nil
}

// MigrateGenesisFile migrate current genesis file content to match of the
// new genesis validator type.
func (m Migrator) MigrateGenesisFile() error {
	var newAg types.AppGenesis
	var ag legacyAppGenesis
	var err error

	if rs, ok := m.reader.(io.ReadSeeker); ok {
		err = json.NewDecoder(rs).Decode(&ag)
		if err != nil {
			return fmt.Errorf("error unmarshalling legacy AppGenesis: %w", err)
		}

		vals := []sdk.GenesisValidator{}
		for _, cmtVal := range ag.Consensus.Validators {
			val := sdk.GenesisValidator{
				Address: sdk.ConsAddress(cmtVal.Address).Bytes(),
				PubKey:  cmtVal.PubKey,
				Power:   cmtVal.Power,
				Name:    cmtVal.Name,
			}

			vals = append(vals, val)
		}

		newAg = types.AppGenesis{
			AppName:       ag.AppName,
			AppVersion:    ag.AppVersion,
			GenesisTime:   ag.GenesisTime,
			ChainID:       ag.ChainID,
			InitialHeight: ag.InitialHeight,
			AppHash:       ag.AppHash,
			AppState:      ag.AppState,
			Consensus: &types.ConsensusGenesis{
				Validators: vals,
				Params:     ag.Consensus.Params,
			},
		}
	}

	err = newAg.ValidateAndComplete()
	if err != nil {
		return err
	}

	return newAg.SaveAs(m.filePath)
}

// CometBFT Genesis Handling for JSON,
// this is necessary for json unmarshaling of legacyConsensusGenesis
func (cs *legacyConsensusGenesis) MarshalJSON() ([]byte, error) {
	type Alias legacyConsensusGenesis
	return cmtjson.Marshal(&Alias{
		Validators: cs.Validators,
		Params:     cs.Params,
	})
}

func (cs *legacyConsensusGenesis) UnmarshalJSON(b []byte) error {
	type Alias legacyConsensusGenesis

	result := Alias{}
	if err := cmtjson.Unmarshal(b, &result); err != nil {
		return err
	}

	cs.Params = result.Params
	cs.Validators = result.Validators

	return nil
}
