package types

import (
	"encoding/json"
	fmt "fmt"
	"os"
	"time"

	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	cmttypes "github.com/cometbft/cometbft/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
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

	Validators      []GenesisValidator        `json:"validators,omitempty"`
	ConsensusParams *cmtproto.ConsensusParams `json:"consensus_params,omitempty"`
}

// ToCometBFTGenesisDoc converts the AppGenesis to a CometBFT GenesisDoc.
func (ag AppGenesis) ToCometBFTGenesisDoc() (*cmttypes.GenesisDoc, error) {
	cmtValidators := make([]cmttypes.GenesisValidator, len(ag.Validators))
	for i, v := range ag.Validators {

		var pubKey cryptotypes.PubKey
		if err := json.Unmarshal(v.ConsensusPubkey.Value, pubKey); err != nil {
			return nil, fmt.Errorf("failed to unmarshal validator consensus pubkey: %v: %w", v.ConsensusPubkey, err)
		}

		cmtPk, err := cryptocodec.ToCmtPubKeyInterface(pubKey)
		if err != nil {
			return nil, fmt.Errorf("failed to convert validator consensus pubkey to cmt proto: %v: %w", v.ConsensusPubkey, err)
		}

		cmtValidators[i] = cmttypes.GenesisValidator{
			Address: sdk.ConsAddress(v.Address).Bytes(),
			PubKey:  cmtPk,
			Power:   v.VotingPower,
			Name:    v.Name,
		}
	}

	return &cmttypes.GenesisDoc{
		ChainID:       ag.ChainID,
		InitialHeight: ag.InitialHeight,
		AppHash:       ag.AppHash,
		AppState:      ag.AppState,
		ConsensusParams: &cmttypes.ConsensusParams{
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
		},
		Validators: cmtValidators,
	}, nil
}

// SaveAs is a utility method for saving AppGenesis as a JSON file.
func (ag *AppGenesis) SaveAs(file string) error {
	appGenesisBytes, err := ag.MarshalIndent("", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(file, appGenesisBytes, 0600)
}

// Marshal the AppGenesis.
func (ag *AppGenesis) MarshalJSON() ([]byte, error) {
	// TODO to fix
	return json.Marshal(&struct{}{})
}

// MarshalIndent marshals the AppGenesis with the provided prefix and indent.
func (ag *AppGenesis) MarshalIndent(prefix, indent string) ([]byte, error) {
	// TODO to fix
	return json.Marshal(&struct{}{})
}

// Unmarshal an AppGenesis from JSON.
func (ag *AppGenesis) UnmarshalJSON(bz []byte) error {
	// TODO to fix
	return nil
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
