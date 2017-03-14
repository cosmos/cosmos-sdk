package app

import (
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
	"github.com/tendermint/basecoin/types"
	cmn "github.com/tendermint/go-common"
	"github.com/tendermint/go-wire"
	tmtypes "github.com/tendermint/tendermint/types"
)

func (app *Basecoin) LoadGenesis(path string) error {
	tmDoc, appDoc, err := loadGenesis(path)
	if err != nil {
		return err
	}
	fmt.Println("TMGendoc", tmDoc)
	fmt.Println("AppGendoc", appDoc)

	app.SetOption("base/chain_id", appDoc.ChainID)
	for _, acc := range appDoc.Accounts {
		accBytes, err := json.Marshal(acc)
		if err != nil {
			return err
		}
		r := app.SetOption("base/account", string(accBytes))
		// TODO: SetOption returns an error
		log.Notice("SetOption", "result", r)
	}
	return nil
}

type keyValue struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// includes tendermint (in the json, we ignore here)
type FullGenesisDoc struct {
	AppOptions *GenesisDoc `json:"app_options"`
}

type GenesisDoc struct {
	ChainID  string          `json:"chain_id"`
	Accounts []types.Account `json:"accounts"`
}

func loadGenesis(filePath string) (*tmtypes.GenesisDoc, *GenesisDoc, error) {
	bytes, err := cmn.ReadFile(filePath)
	if err != nil {
		return nil, nil, errors.Wrap(err, "loading genesis file")
	}

	tmGenesis := new(tmtypes.GenesisDoc)
	appGenesis := new(FullGenesisDoc)

	// the tendermint genesis is go-wire
	err = wire.ReadJSONBytes(bytes, tmGenesis)

	// the basecoin genesis go-data :)
	err = json.Unmarshal(bytes, appGenesis)
	if err != nil {
		return nil, nil, errors.Wrap(err, "unmarshaling genesis file")
	}
	return tmGenesis, appGenesis.AppOptions, nil
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
