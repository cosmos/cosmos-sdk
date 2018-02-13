package baseapp

import (
	"encoding/json"
	"io/ioutil"
)

// TODO: remove from here and pass the AppState
// through InitChain

// GenesisDoc defines the initial conditions for a tendermint blockchain, in particular its validator set.
type GenesisDoc struct {
	AppState json.RawMessage `json:"app_state,omitempty"`
}

// GenesisDocFromFile reads JSON data from a file and unmarshalls it into a GenesisDoc.
func GenesisDocFromFile(genDocFile string) (*GenesisDoc, error) {
	if genDocFile == "" {
		var g GenesisDoc
		return &g, nil
	}
	jsonBlob, err := ioutil.ReadFile(genDocFile)
	if err != nil {
		return nil, err
	}
	genDoc := GenesisDoc{}
	err = json.Unmarshal(jsonBlob, &genDoc)
	if err != nil {
		return nil, err
	}
	return &genDoc, nil
}
