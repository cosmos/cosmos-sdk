package app

import (
	"bytes"
	"encoding/json"

	"github.com/pkg/errors"

	"github.com/tendermint/basecoin/types"
	crypto "github.com/tendermint/go-crypto"
	"github.com/tendermint/go-wire/data"
	cmn "github.com/tendermint/tmlibs/common"
)

func (app *Basecoin) LoadGenesis(path string) error {
	genDoc, err := loadGenesis(path)
	if err != nil {
		return err
	}

	// set chain_id
	app.SetOption("base/chain_id", genDoc.ChainID)

	// set accounts
	for _, acc := range genDoc.AppOptions.Accounts {
		accBytes, err := json.Marshal(acc)
		if err != nil {
			return err
		}
		r := app.SetOption("base/account", string(accBytes))
		// TODO: SetOption returns an error
		app.logger.Info("Done setting Account via SetOption", "result", r)
	}

	// set plugin options
	for _, kv := range genDoc.AppOptions.pluginOptions {
		r := app.SetOption(kv.Key, kv.Value)
		app.logger.Info("Done setting Plugin key-value pair via SetOption", "result", r, "k", kv.Key, "v", kv.Value)
	}

	return nil
}

type keyValue struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// includes tendermint (in the json, we ignore here)
type FullGenesisDoc struct {
	ChainID    string      `json:"chain_id"`
	AppOptions *GenesisDoc `json:"app_options"`
}

type GenesisDoc struct {
	Accounts      []GenesisAccount  `json:"accounts"`
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

func parseGenesisList(kvz_ []json.RawMessage) (kvz []keyValue, err error) {
	if len(kvz_)%2 != 0 {
		return nil, errors.New("genesis cannot have an odd number of items.  Format = [key1, value1, key2, value2, ...]")
	}

	for i := 0; i < len(kvz_); i += 2 {
		kv := keyValue{}
		rawK := []byte(kvz_[i])
		err := json.Unmarshal(rawK, &(kv.Key))
		if err != nil {
			return nil, errors.Errorf("Non-string key: %s", string(rawK))
		}
		// convert value to string if possible (otherwise raw json)
		rawV := kvz_[i+1]
		err = json.Unmarshal(rawV, &(kv.Value))
		if err != nil {
			kv.Value = string(rawV)
		}
		kvz = append(kvz, kv)
	}
	return kvz, nil
}

/**** code to parse accounts from genesis docs ***/

type GenesisAccount struct {
	Address data.Bytes `json:"address"`
	// this from types.Account (don't know how to embed this properly)
	PubKey   crypto.PubKey `json:"pub_key"` // May be nil, if not known.
	Sequence int           `json:"sequence"`
	Balance  types.Coins   `json:"coins"`
}

func (g GenesisAccount) ToAccount() *types.Account {
	return &types.Account{
		PubKey:   g.PubKey,
		Sequence: g.Sequence,
		Balance:  g.Balance,
	}
}

func (g GenesisAccount) GetAddr() ([]byte, error) {
	noAddr, noPk := len(g.Address) == 0, g.PubKey.Empty()

	if noAddr {
		if noPk {
			return nil, errors.New("No address given")
		}
		return g.PubKey.Address(), nil
	}
	if noPk { // but is addr...
		return g.Address, nil
	}
	// now, we have both, make sure they check out
	if bytes.Equal(g.Address, g.PubKey.Address()) {
		return g.Address, nil
	}
	return nil, errors.New("Address and pubkey don't match")
}
