package main

import (
	"errors"

	"github.com/urfave/cli"

	"github.com/tendermint/abci/server"
	cmn "github.com/tendermint/go-common"
	cfg "github.com/tendermint/go-config"
	//logger "github.com/tendermint/go-logger"
	eyes "github.com/tendermint/merkleeyes/client"

	tmcfg "github.com/tendermint/tendermint/config/tendermint"
	"github.com/tendermint/tendermint/node"
	"github.com/tendermint/tendermint/proxy"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/tendermint/basecoin/app"
)

var config cfg.Config

const EyesCacheSize = 10000

func cmdStart(c *cli.Context) error {

	// Connect to MerkleEyes
	var eyesCli *eyes.Client
	if c.String("eyes") == "local" {
		eyesCli = eyes.NewLocalClient(c.String("eyes-db"), EyesCacheSize)
	} else {
		var err error
		eyesCli, err = eyes.NewClient(c.String("eyes"))
		if err != nil {
			return errors.New("connect to MerkleEyes: " + err.Error())
		}
	}

	// Create Basecoin app
	basecoinApp := app.NewBasecoin(eyesCli)

	// If genesis file was specified, set key-value options
	if c.String("genesis") != "" {
		err := basecoinApp.LoadGenesis(c.String("genesis"))
		if err != nil {
			return errors.New(cmn.Fmt("%+v", err))
		}
	}

	if c.Bool("in-proc") {
		startTendermint(c, basecoinApp)
	} else {
		startBasecoinABCI(c, basecoinApp)
	}

	return nil
}

func startBasecoinABCI(c *cli.Context, basecoinApp *app.Basecoin) error {
	// Start the ABCI listener
	svr, err := server.NewServer(c.String("address"), "socket", basecoinApp)
	if err != nil {
		return errors.New("create listener: " + err.Error())
	}
	// Wait forever
	cmn.TrapSignal(func() {
		// Cleanup
		svr.Stop()
	})
	return nil

}

func startTendermint(c *cli.Context, basecoinApp *app.Basecoin) {
	// Get configuration
	config = tmcfg.GetConfig("")
	// logger.SetLogLevel("notice") //config.GetString("log_level"))

	// parseFlags(config, args[1:]) // Command line overrides

	// Create & start tendermint node
	privValidatorFile := config.GetString("priv_validator_file")
	privValidator := tmtypes.LoadOrGenPrivValidator(privValidatorFile)
	n := node.NewNode(config, privValidator, proxy.NewLocalClientCreator(basecoinApp))

	n.Start()

	// Wait forever
	cmn.TrapSignal(func() {
		// Cleanup
		n.Stop()
	})
}
