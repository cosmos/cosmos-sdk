package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"reflect"

	"github.com/tendermint/basecoin/app"
	. "github.com/tendermint/go-common"
	eyes "github.com/tendermint/merkleeyes/client"
	"github.com/tendermint/tmsp/server"
)

func main() {

	addrPtr := flag.String("address", "tcp://0.0.0.0:46658", "Listen address")
	eyesPtr := flag.String("eyes", "tcp://0.0.0.0:46659", "MerkleEyes address")
	genFilePath := flag.String("genesis", "", "Genesis file, if any")
	flag.Parse()

	// Connect to MerkleEyes
	eyesCli, err := eyes.NewClient(*eyesPtr)
	if err != nil {
		Exit("connect to MerkleEyes: " + err.Error())
	}

	// Create Basecoin app
	app := app.NewBasecoin(eyesCli)

	// If genesis file was specified, set key-value options
	if *genFilePath != "" {
		kvz := loadGenesis(*genFilePath)
		for _, kv := range kvz {
			log := app.SetOption(kv.Key, kv.Value)
			fmt.Println(Fmt("Set %v=%v. Log: %v", kv.Key, kv.Value, log))
		}
	}

	// Start the listener
	svr, err := server.NewServer(*addrPtr, app)
	if err != nil {
		Exit("create listener: " + err.Error())
	}

	// Wait forever
	TrapSignal(func() {
		// Cleanup
		svr.Stop()
	})

}

//----------------------------------------

type KeyValue struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func loadGenesis(filePath string) (kvz []KeyValue) {
	kvz_ := []interface{}{}
	bytes, err := ReadFile(filePath)
	if err != nil {
		Exit("loading genesis file: " + err.Error())
	}
	err = json.Unmarshal(bytes, &kvz_)
	if err != nil {
		Exit("parsing genesis file: " + err.Error())
	}
	if len(kvz_)%2 != 0 {
		Exit("genesis cannot have an odd number of items.  Format = [key1, value1, key2, value2, ...]")
	}
	for i := 0; i < len(kvz_); i += 2 {
		keyIfc := kvz_[i]
		valueIfc := kvz_[i+1]
		var key, value string
		key, ok := keyIfc.(string)
		if !ok {
			Exit(Fmt("genesis had invalid key %v of type %v", keyIfc, reflect.TypeOf(keyIfc)))
		}
		if value_, ok := valueIfc.(string); ok {
			value = value_
		} else {
			valueBytes, err := json.Marshal(valueIfc)
			if err != nil {
				Exit(Fmt("genesis had invalid value %v: %v", value_, err.Error()))
			}
			value = string(valueBytes)
		}
		kvz = append(kvz, KeyValue{key, value})
	}
	return kvz
}
