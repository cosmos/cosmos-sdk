package types

import (
	"encoding/json"
	"os"

	cmtjson "github.com/cometbft/cometbft/libs/json"
	cmttypes "github.com/cometbft/cometbft/types"
	"github.com/cosmos/cosmos-sdk/version"
)

// AppGenesisOnly defines the app's genesis.
type AppGenesisOnly struct {
	AppName    string `json:"app_name"`
	AppVersion string `json:"app_version"`
}

// AppGenesis defines the app's genesis with the consensus genesis.
type AppGenesis struct {
	AppGenesisOnly

	cmttypes.GenesisDoc
}

func NewAppGenesis(genesisDoc cmttypes.GenesisDoc) *AppGenesis {
	return &AppGenesis{
		AppGenesisOnly: AppGenesisOnly{
			AppName:    version.AppName,
			AppVersion: version.Version,
		},
		GenesisDoc: genesisDoc,
	}
}

// SaveAs is a utility method for saving AppGenesis as a JSON file.
func (ag *AppGenesis) SaveAs(file string) error {
	// appGenesisBytes, err := ag.MarshalIndent("", "  ")
	appGenesisBytes, err := json.Marshal(ag)
	if err != nil {
		return err
	}

	return os.WriteFile(file, appGenesisBytes, 0644)
}

// Marshal the AppGenesis.
func (ag *AppGenesis) MarshalJSON() ([]byte, error) {
	// marshal the genesis doc with CometBFT lib
	// if GenesisDoc was implementing MarshalJSON and UnmarshalJSON, this would be much simpler
	genDoc, err := cmtjson.Marshal(ag.GenesisDoc)
	if err != nil {
		return nil, err
	}

	appGenesis, err := json.Marshal(ag.AppGenesisOnly)
	if err != nil {
		return nil, err
	}

	out := map[string]interface{}{}
	if err = json.Unmarshal(appGenesis, &out); err != nil {
		return nil, err
	}

	if err = cmtjson.Unmarshal(genDoc, &out); err != nil {
		return nil, err
	}

	// unmarshal the genesis doc with stdlib
	return cmtjson.Marshal(out)
}

// MarshalIndent marshals the AppGenesis with the provided prefix and indent.
func (ag *AppGenesis) MarshalIndent(prefix, indent string) ([]byte, error) {
	return cmtjson.MarshalIndent(ag, prefix, indent)
}

// Unmarshal an AppGenesis from JSON.
func (ag *AppGenesis) UnmarshalJSON(bz []byte) error {
	return cmtjson.Unmarshal(bz, &ag)
}
