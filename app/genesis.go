package app

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/pkg/errors"
	cmn "github.com/tendermint/go-common"
)

func (app *Basecoin) LoadGenesis(path string) error {
	kvz, err := loadGenesis(path)
	if err != nil {
		return err
	}
	for _, kv := range kvz {
		log := app.SetOption(kv.Key, kv.Value)
		// TODO: remove debug output
		fmt.Printf("Set %v=%v. Log: %v\n", kv.Key, kv.Value, log)
	}
	return nil
}

type keyValue struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func loadGenesis(filePath string) (kvz []keyValue, err error) {
	kvz_ := []interface{}{}
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
		keyIfc := kvz_[i]
		valueIfc := kvz_[i+1]
		var key, value string
		key, ok := keyIfc.(string)
		if !ok {
			return nil, errors.Errorf("genesis had invalid key %v of type %v", keyIfc, reflect.TypeOf(keyIfc))
		}
		if value_, ok := valueIfc.(string); ok {
			value = value_
		} else {
			valueBytes, err := json.Marshal(valueIfc)
			if err != nil {
				return nil, errors.Errorf("genesis had invalid value %v: %v", value_, err.Error())
			}
			value = string(valueBytes)
		}
		kvz = append(kvz, keyValue{key, value})
	}
	return kvz, nil
}
