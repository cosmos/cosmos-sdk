package main

import (
	"flag"

	"github.com/username/bcTutorial/plugins/counter"

	"github.com/tendermint/abci/server"
	"github.com/tendermint/basecoin/app"
	cmn "github.com/tendermint/go-common"
	eyes "github.com/tendermint/merkleeyes/client"
)

func main() {

	//flags
	addrPtr := flag.String("address", "tcp://0.0.0.0:46658", "Listen address")
	eyesPtr := flag.String("eyes", "local", "MerkleEyes address, or 'local' for embedded")
	genFilePath := flag.String("genesis", "", "Genesis file, if any")
	flag.Parse()

	// Connect to MerkleEyes
	eyesCli, err := eyes.NewClient(*eyesPtr, "socket")
	if err != nil {
		cmn.Exit("connect to MerkleEyes: " + err.Error())
	}

	// Create Basecoin
	bcApp := app.NewBasecoin(eyesCli)

	// create/add plugins
	counter := counter.New("counter")
	bcApp.RegisterPlugin(counter)

	// If genesis file was specified, set key-value options
	if *genFilePath != "" {
		err := bcApp.LoadGenesis(*genFilePath)
		if err != nil {
			cmn.Exit(cmn.Fmt("%+v", err))
		}
	}

	// Start the listener
	svr, err := server.NewServer(*addrPtr, "socket", bcApp)
	if err != nil {
		cmn.Exit("create listener: " + err.Error())
	}

	// Wait forever
	cmn.TrapSignal(func() {
		// Cleanup
		svr.Stop()
	})
}
