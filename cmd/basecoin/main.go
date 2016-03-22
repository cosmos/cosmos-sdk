package main

import (
	"encoding/json"
	"flag"
	"fmt"

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
	bytes, err := ReadFile(filePath)
	if err != nil {
		Exit("loading genesis file: " + err.Error())
	}
	fmt.Println(">>", string(bytes))
	err = json.Unmarshal(bytes, &kvz)
	if err != nil {
		Exit("parsing genesis file: " + err.Error())
	}
	return
}
