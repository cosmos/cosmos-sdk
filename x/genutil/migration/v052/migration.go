package migration

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	cmtjson "github.com/cometbft/cometbft/libs/json"
	cmttypes "github.com/cometbft/cometbft/types"

	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
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

func NewMigrator(filePath string) (*Migrator, error) {
	file, err := os.Open(filepath.Clean(filePath))
	if err != nil {
		return nil, err
	}

	return &Migrator{
		reader: file,
	}, nil
}

// MigrateGenesisFile migrate current genesis file content to match of the
// new genesis validator type.
func (m Migrator) MigrateGenesisFile() (*types.AppGenesis, error) {
	var newAg types.AppGenesis
	var ag legacyAppGenesis
	var err error

	if rs, ok := m.reader.(io.ReadSeeker); ok {
		err = json.NewDecoder(rs).Decode(&ag)
		if err == nil {
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

			return &newAg, nil
		}

		err = fmt.Errorf("error unmarshalling legacy AppGenesis: %w", err)
		if _, serr := rs.Seek(0, io.SeekStart); serr != nil {
			err = errors.Join(err, fmt.Errorf("error seeking back to the front: %w", serr))
			return nil, err
		}
	}

	jsonBlob, ioerr := io.ReadAll(m.reader)
	if ioerr != nil {
		err = errors.Join(err, fmt.Errorf("failed to read file completely: %w", ioerr))
		return nil, err
	}

	// fallback to comet genesis parsing
	var ctmGenesis cmttypes.GenesisDoc
	if uerr := cmtjson.Unmarshal(jsonBlob, &ctmGenesis); uerr != nil {
		err = errors.Join(err, fmt.Errorf("failed fallback to CometBFT GenDoc: %w", uerr))
		return nil, err
	}

	vals := []sdk.GenesisValidator{}
	for _, cmtVal := range ctmGenesis.Validators {
		val := sdk.GenesisValidator{
			Address: cmtVal.Address.Bytes(),
			PubKey:  cmtVal.PubKey,
			Power:   cmtVal.Power,
			Name:    cmtVal.Name,
		}

		vals = append(vals, val)
	}

	newAg = types.AppGenesis{
		AppName: version.AppName,
		// AppVersion is not filled as we do not know it from a CometBFT genesis
		GenesisTime:   ctmGenesis.GenesisTime,
		ChainID:       ctmGenesis.ChainID,
		InitialHeight: ctmGenesis.InitialHeight,
		AppHash:       ctmGenesis.AppHash,
		AppState:      ctmGenesis.AppState,
		Consensus: &types.ConsensusGenesis{
			Validators: vals,
			Params:     ctmGenesis.ConsensusParams,
		},
	}

	return &newAg, nil
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

// since we only need migrate the consensus validators content so there is no
// exported state migration.
func Migrate(appState types.AppMap, _ client.Context) (types.AppMap, error) {
	return appState, nil
}
