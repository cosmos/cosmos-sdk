package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"reflect"

	"github.com/tendermint/abci/server"
	"github.com/tendermint/basecoin-examples/mintnet"
	"github.com/tendermint/basecoin/app"
	cmn "github.com/tendermint/go-common"
	eyes "github.com/tendermint/merkleeyes/client"
)

func main() {

	addrPtr := flag.String("address", "tcp://0.0.0.0:46658", "Listen address")
	eyesPtr := flag.String("eyes", "local", "MerkleEyes address, or 'local' for embedded")
	genFilePath := flag.String("genesis", "", "Genesis file, if any")
	flag.Parse()

	// Connect to MerkleEyes
	eyesCli, err := eyes.NewClient(*eyesPtr, "socket")
	if err != nil {
		cmn.Exit("connect to MerkleEyes: " + err.Error())
	}

	// Create Basecoin app
	coin := app.NewBasecoin(eyesCli)

	// attach the plugin
	mint := mintnet.NewMintPlugin("mint")
	coin.RegisterPlugin(mint)

	// If genesis file was specified, set key-value options
	if *genFilePath != "" {
		kvz := loadGenesis(*genFilePath)
		for _, kv := range kvz {
			log := coin.SetOption(kv.Key, kv.Value)
			fmt.Println(cmn.Fmt("Set %v=%v. Log: %v", kv.Key, kv.Value, log))
		}
	}

	// Start the listener
	svr, err := server.NewServer(*addrPtr, "socket", coin)
	if err != nil {
		cmn.Exit("create listener: " + err.Error())
	}

	// Wait forever
	cmn.TrapSignal(func() {
		// Cleanup
		svr.Stop()
	})

}

//------  FIXME: all this stuff is refactored in a basecoin branch

type KeyValue struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func loadGenesis(filePath string) (kvz []KeyValue) {
	kvz_ := []interface{}{}
	bytes, err := cmn.ReadFile(filePath)
	if err != nil {
		cmn.Exit("loading genesis file: " + err.Error())
	}
	err = json.Unmarshal(bytes, &kvz_)
	if err != nil {
		cmn.Exit("parsing genesis file: " + err.Error())
	}
	if len(kvz_)%2 != 0 {
		cmn.Exit("genesis cannot have an odd number of items.  Format = [key1, value1, key2, value2, ...]")
	}
	for i := 0; i < len(kvz_); i += 2 {
		keyIfc := kvz_[i]
		valueIfc := kvz_[i+1]
		var key, value string
		key, ok := keyIfc.(string)
		if !ok {
			cmn.Exit(cmn.Fmt("genesis had invalid key %v of type %v", keyIfc, reflect.TypeOf(keyIfc)))
		}
		if value_, ok := valueIfc.(string); ok {
			value = value_
		} else {
			valueBytes, err := json.Marshal(valueIfc)
			if err != nil {
				cmn.Exit(cmn.Fmt("genesis had invalid value %v: %v", value_, err.Error()))
			}
			value = string(valueBytes)
		}
		kvz = append(kvz, KeyValue{key, value})
	}
	return kvz
}
