package commands

import (
	"errors"
	"os"
	"path"

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
	"github.com/tendermint/basecoin/plugins/counter"
	"github.com/tendermint/basecoin/plugins/ibc"
)

var config cfg.Config

const EyesCacheSize = 10000

var StartCmd = cli.Command{
	Name:      "start",
	Usage:     "Start basecoin",
	ArgsUsage: "",
	Action: func(c *cli.Context) error {
		return cmdStart(c)
	},
	Flags: []cli.Flag{
		addrFlag,
		eyesFlag,
		dirFlag,
		inProcTMFlag,
		chainIDFlag,
		ibcPluginFlag,
		counterPluginFlag,
	},
}

func cmdStart(c *cli.Context) error {

	// Connect to MerkleEyes
	var eyesCli *eyes.Client
	if c.String("eyes") == "local" {
		eyesCli = eyes.NewLocalClient(path.Join(c.String("dir"), "merkleeyes.db"), EyesCacheSize)
	} else {
		var err error
		eyesCli, err = eyes.NewClient(c.String("eyes"))
		if err != nil {
			return errors.New("connect to MerkleEyes: " + err.Error())
		}
	}

	// Create Basecoin app
	basecoinApp := app.NewBasecoin(eyesCli)

	if c.Bool("counter-plugin") {
		basecoinApp.RegisterPlugin(counter.New("counter"))
	}

	if c.Bool("ibc-plugin") {
		basecoinApp.RegisterPlugin(ibc.New())

	}

	// If genesis file exists, set key-value options
	genesisFile := path.Join(c.String("dir"), "genesis.json")
	if _, err := os.Stat(genesisFile); err == nil {
		err := basecoinApp.LoadGenesis(genesisFile)
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
