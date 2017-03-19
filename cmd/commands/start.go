package commands

import (
	"errors"
	"fmt"
	"os"
	"path"

	"github.com/urfave/cli"

	"github.com/tendermint/abci/server"
	cmn "github.com/tendermint/go-common"
	eyes "github.com/tendermint/merkleeyes/client"

	tmcfg "github.com/tendermint/tendermint/config/tendermint"
	"github.com/tendermint/tendermint/node"
	"github.com/tendermint/tendermint/proxy"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/tendermint/basecoin/app"
	"github.com/tendermint/basecoin/types"
)

const EyesCacheSize = 10000

var StartCmd = cli.Command{
	Name:      "start",
	Usage:     "Start basecoin",
	ArgsUsage: "",
	Action: func(c *cli.Context) error {
		return cmdStart(c)
	},
	Flags: []cli.Flag{
		AddrFlag,
		EyesFlag,
		WithoutTendermintFlag,
		ChainIDFlag,
	},
}

type plugin struct {
	name      string
	newPlugin func() types.Plugin
}

var plugins = []plugin{}

// RegisterStartPlugin is used to enable a plugin
func RegisterStartPlugin(name string, newPlugin func() types.Plugin) {
	plugins = append(plugins, plugin{name: name, newPlugin: newPlugin})
}

func cmdStart(c *cli.Context) error {
	basecoinDir := BasecoinRoot("")

	// Connect to MerkleEyes
	var eyesCli *eyes.Client
	if c.String("eyes") == "local" {
		eyesCli = eyes.NewLocalClient(path.Join(basecoinDir, "data", "merkleeyes.db"), EyesCacheSize)
	} else {
		var err error
		eyesCli, err = eyes.NewClient(c.String("eyes"))
		if err != nil {
			return errors.New("connect to MerkleEyes: " + err.Error())
		}
	}

	// Create Basecoin app
	basecoinApp := app.NewBasecoin(eyesCli)

	// register IBC plugn
	basecoinApp.RegisterPlugin(NewIBCPlugin())

	// register all other plugins
	for _, p := range plugins {
		basecoinApp.RegisterPlugin(p.newPlugin())
	}

	// if chain_id has not been set yet, load the genesis.
	// else, assume it's been loaded
	if basecoinApp.GetState().GetChainID() == "" {
		// If genesis file exists, set key-value options
		genesisFile := path.Join(basecoinDir, "genesis.json")
		if _, err := os.Stat(genesisFile); err == nil {
			err := basecoinApp.LoadGenesis(genesisFile)
			if err != nil {
				return errors.New(cmn.Fmt("%+v", err))
			}
		} else {
			fmt.Printf("No genesis file at %s, skipping...\n", genesisFile)
		}
	}

	chainID := basecoinApp.GetState().GetChainID()
	if c.Bool("without-tendermint") {
		log.Notice("Starting Basecoin without Tendermint", "chain_id", chainID)
		// run just the abci app/server
		return startBasecoinABCI(c, basecoinApp)
	} else {
		log.Notice("Starting Basecoin with Tendermint", "chain_id", chainID)
		// start the app with tendermint in-process
		return startTendermint(basecoinDir, basecoinApp)
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

func startTendermint(dir string, basecoinApp *app.Basecoin) error {
	// Get configuration
	tmConfig := tmcfg.GetConfig(dir)

	// logger.SetLogLevel("notice") //config.GetString("log_level"))

	// parseFlags(config, args[1:]) // Command line overrides

	// Create & start tendermint node
	privValidatorFile := tmConfig.GetString("priv_validator_file")
	privValidator := tmtypes.LoadOrGenPrivValidator(privValidatorFile)
	n := node.NewNode(tmConfig, privValidator, proxy.NewLocalClientCreator(basecoinApp))

	_, err := n.Start()
	if err != nil {
		return err
	}

	// Wait forever
	cmn.TrapSignal(func() {
		// Cleanup
		n.Stop()
	})

	return nil
}
