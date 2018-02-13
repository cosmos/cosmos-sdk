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

// read app state from the genesis file
//func GenesisAppState(genesisFile string) (state json.RawMessage, err error) {
//if genesisFile == "" {
//return
//}
//jsonBlob, err := ioutil.ReadFile(genesisFile)
//if err != nil {
//return nil, err
//}
//data := make(map[string]interface{})
//err = json.Unmarshal(jsonBlob, &data)
//if err != nil {
//return nil, err
//}
//state, ok := data["app_state"].(json.RawMessage)
//if !ok {
//return nil, errors.New("app state genesis parse error")
//}
//return state, nil
//}
