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

	type ConsensusGenesis []byte

	// unmarshal the genesis doc with stdlib
	return json.Marshal(&struct {
		AppGenesisOnly
		ConsensusGenesis
	}{
		AppGenesisOnly:   ag.AppGenesisOnly,
		ConsensusGenesis: genDoc,
	})
}

// MarshalIndent marshals the AppGenesis with the provided prefix and indent.
func (ag *AppGenesis) MarshalIndent(prefix, indent string) ([]byte, error) {
	return cmtjson.MarshalIndent(ag, prefix, indent)
}

// Unmarshal an AppGenesis from JSON.
func (ag *AppGenesis) UnmarshalJSON(bz []byte) error {
	type Alias AppGenesis // we alias for avoiding recursion in UnmarshalJSON
	var result Alias

	if err := cmtjson.Unmarshal(bz, &result); err != nil {
		return err
	}

	ag = (*AppGenesis)(&result)
	return nil
}
