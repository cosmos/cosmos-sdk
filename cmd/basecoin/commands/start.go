package commands

import (
	"fmt"
	"os"
	"path"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/tendermint/abci/server"
	"github.com/tendermint/basecoin"
	eyesApp "github.com/tendermint/merkleeyes/app"
	eyes "github.com/tendermint/merkleeyes/client"
	"github.com/tendermint/tmlibs/cli"
	cmn "github.com/tendermint/tmlibs/common"

	tcmd "github.com/tendermint/tendermint/cmd/tendermint/commands"
	"github.com/tendermint/tendermint/node"
	"github.com/tendermint/tendermint/proxy"
	"github.com/tendermint/tendermint/types"

	"github.com/tendermint/basecoin/app"
)

// StartCmd - command to start running the basecoin node!
var StartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start basecoin",
	RunE:  startCmd,
}

// nolint TODO: move to config file
const EyesCacheSize = 10000

//nolint
const (
	FlagAddress           = "address"
	FlagEyes              = "eyes"
	FlagWithoutTendermint = "without-tendermint"
)

var (
	// Handler - use a global to store the handler, so we can set it in main.
	// TODO: figure out a cleaner way to register plugins
	Handler basecoin.Handler
)

func init() {
	flags := StartCmd.Flags()
	flags.String(FlagAddress, "tcp://0.0.0.0:46658", "Listen address")
	flags.String(FlagEyes, "local", "MerkleEyes address, or 'local' for embedded")
	flags.Bool(FlagWithoutTendermint, false, "Only run basecoin abci app, assume external tendermint process")
	// add all standard 'tendermint node' flags
	tcmd.AddNodeFlags(StartCmd)
}

func startCmd(cmd *cobra.Command, args []string) error {
	rootDir := viper.GetString(cli.HomeFlag)
	meyes := viper.GetString(FlagEyes)

	// Connect to MerkleEyes
	var eyesCli *eyes.Client
	if meyes == "local" {
		eyesApp.SetLogger(logger.With("module", "merkleeyes"))
		eyesCli = eyes.NewLocalClient(path.Join(rootDir, "data", "merkleeyes.db"), EyesCacheSize)
	} else {
		var err error
		eyesCli, err = eyes.NewClient(meyes)
		if err != nil {
			return errors.Errorf("Error connecting to MerkleEyes: %v\n", err)
		}
	}

	// Create Basecoin app
	basecoinApp := app.NewBasecoin(Handler, eyesCli, logger.With("module", "app"))

	// if chain_id has not been set yet, load the genesis.
	// else, assume it's been loaded
	if basecoinApp.GetState().GetChainID() == "" {
		// If genesis file exists, set key-value options
		genesisFile := path.Join(rootDir, "genesis.json")
		if _, err := os.Stat(genesisFile); err == nil {
			err := basecoinApp.LoadGenesis(genesisFile)
			if err != nil {
				return errors.Errorf("Error in LoadGenesis: %v\n", err)
			}
		} else {
			fmt.Printf("No genesis file at %s, skipping...\n", genesisFile)
		}
	}

	chainID := basecoinApp.GetState().GetChainID()
	if viper.GetBool(FlagWithoutTendermint) {
		logger.Info("Starting Basecoin without Tendermint", "chain_id", chainID)
		// run just the abci app/server
		return startBasecoinABCI(basecoinApp)
	}
	logger.Info("Starting Basecoin with Tendermint", "chain_id", chainID)
	// start the app with tendermint in-process
	return startTendermint(rootDir, basecoinApp)
}

func startBasecoinABCI(basecoinApp *app.Basecoin) error {
	// Start the ABCI listener
	addr := viper.GetString(FlagAddress)
	svr, err := server.NewServer(addr, "socket", basecoinApp)
	if err != nil {
		return errors.Errorf("Error creating listener: %v\n", err)
	}
	svr.SetLogger(logger.With("module", "abci-server"))
	svr.Start()

	// Wait forever
	cmn.TrapSignal(func() {
		// Cleanup
		svr.Stop()
	})
	return nil
}

func startTendermint(dir string, basecoinApp *app.Basecoin) error {
	cfg, err := tcmd.ParseConfig()
	if err != nil {
		return err
	}

	// Create & start tendermint node
	privValidator := types.LoadOrGenPrivValidator(cfg.PrivValidatorFile(), logger)
	n := node.NewNode(cfg, privValidator, proxy.NewLocalClientCreator(basecoinApp), logger.With("module", "node"))

	_, err = n.Start()
	if err != nil {
		return err
	}

	// Trap signal, run forever.
	n.RunForever()
	return nil
}
