package app

import (
	"encoding/json"

	"github.com/pkg/errors"
	cmn "github.com/tendermint/go-common"
)

func (app *Basecoin) LoadGenesis(path string) error {
	kvz, err := loadGenesis(path)
	if err != nil {
		return err
	}
	for _, kv := range kvz {
		app.SetOption(kv.Key, kv.Value)
	}
	return nil
}

type keyValue struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func loadGenesis(filePath string) (kvz []keyValue, err error) {
	kvz_ := []json.RawMessage{}
	bytes, err := cmn.ReadFile(filePath)
	if err != nil {
		return nil, errors.Wrap(err, "loading genesis file")
	}
	err = json.Unmarshal(bytes, &kvz_)
	if err != nil {
		return nil, errors.Wrap(err, "parsing genesis file")
	}
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
