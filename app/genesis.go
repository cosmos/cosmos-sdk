package app

import (
	"encoding/json"

	"github.com/pkg/errors"

	cmn "github.com/tendermint/tmlibs/common"
)

// LoadGenesis - Load the genesis file into memory
func (app *Basecoin) LoadGenesis(path string) error {
	genDoc, err := loadGenesis(path)
	if err != nil {
		return err
	}

	// set chain_id
	app.SetOption("base/chain_id", genDoc.ChainID)

	// set accounts
	for _, acct := range genDoc.AppOptions.Accounts {
		_ = app.SetOption("coin/account", string(acct))
	}

	// set plugin options
	for _, kv := range genDoc.AppOptions.pluginOptions {
		_ = app.SetOption(kv.Key, kv.Value)
	}

	return nil
}

type keyValue struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// FullGenesisDoc - includes tendermint (in the json, we ignore here)
type FullGenesisDoc struct {
	ChainID    string      `json:"chain_id"`
	AppOptions *GenesisDoc `json:"app_options"`
}

// GenesisDoc - All genesis values
type GenesisDoc struct {
	Accounts      []json.RawMessage `json:"accounts"`
	PluginOptions []json.RawMessage `json:"plugin_options"`

	pluginOptions []keyValue // unmarshaled rawmessages
}

func loadGenesis(filePath string) (*FullGenesisDoc, error) {
	bytes, err := cmn.ReadFile(filePath)
	if err != nil {
		return nil, errors.Wrap(err, "loading genesis file")
	}

	// the basecoin genesis go-wire/data :)
	genDoc := new(FullGenesisDoc)
	err = json.Unmarshal(bytes, genDoc)
	if err != nil {
		return nil, errors.Wrap(err, "unmarshaling genesis file")
	}

	if genDoc.AppOptions == nil {
		genDoc.AppOptions = new(GenesisDoc)
	}

	pluginOpts, err := parseGenesisList(genDoc.AppOptions.PluginOptions)
	if err != nil {
		return nil, err
	}
	genDoc.AppOptions.pluginOptions = pluginOpts
	return genDoc, nil
}

func parseGenesisList(kvzIn []json.RawMessage) (kvz []keyValue, err error) {
	if len(kvzIn)%2 != 0 {
		return nil, errors.New("genesis cannot have an odd number of items.  Format = [key1, value1, key2, value2, ...]")
	}

	for i := 0; i < len(kvzIn); i += 2 {
		kv := keyValue{}
		rawK := []byte(kvzIn[i])
		err := json.Unmarshal(rawK, &(kv.Key))
		if err != nil {
			return nil, errors.Errorf("Non-string key: %s", string(rawK))
		}
		// convert value to string if possible (otherwise raw json)
		rawV := kvzIn[i+1]
		err = json.Unmarshal(rawV, &(kv.Value))
		if err != nil {
			kv.Value = string(rawV)
		}
		kvz = append(kvz, kv)
	}
	return kvz, nil
}
