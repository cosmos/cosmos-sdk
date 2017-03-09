package commands

import (
	"fmt"
	"os"
	"path"

	"github.com/spf13/cobra"

	"github.com/tendermint/abci/server"
	cmn "github.com/tendermint/go-common"
	eyes "github.com/tendermint/merkleeyes/client"

	tmcfg "github.com/tendermint/tendermint/config/tendermint"
	"github.com/tendermint/tendermint/node"
	"github.com/tendermint/tendermint/proxy"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/tendermint/basecoin/app"
)

var StartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start basecoin",
	Run:   startCmd,
}

//flags
var (
	addrFlag              string
	eyesFlag              string
	dirFlag               string
	withoutTendermintFlag bool
)

// TODO: move to config file
const EyesCacheSize = 10000

func init() {

	flags := []Flag2Register{
		{&addrFlag, "address", "tcp://0.0.0.0:46658", "Listen address"},
		{&eyesFlag, "eyes", "local", "MerkleEyes address, or 'local' for embedded"},
		{&dirFlag, "dir", ".", "Root directory"},
		{&withoutTendermintFlag, "without-tendermint", false, "Run Tendermint in-process with the App"},
	}
	RegisterFlags(StartCmd, flags)

	// TODO: move to config file
	// eyesCacheSizePtr := flag.Int("eyes-cache-size", 10000, "MerkleEyes db cache size, for embedded")
}

func startCmd(cmd *cobra.Command, args []string) {
	basecoinDir := BasecoinRoot("")

	// Connect to MerkleEyes
	var eyesCli *eyes.Client
	if eyesFlag == "local" {
		eyesCli = eyes.NewLocalClient(path.Join(basecoinDir, "data", "merkleeyes.db"), EyesCacheSize)
	} else {
		var err error
		eyesCli, err = eyes.NewClient(eyesFlag)
		if err != nil {
			cmn.Exit(fmt.Sprintf("Error connecting to MerkleEyes: %+v\n", err))
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
				cmn.Exit(fmt.Sprintf("Error in LoadGenesis: %+v\n", err))
			}
		} else {
			fmt.Printf("No genesis file at %s, skipping...\n", genesisFile)
		}
	}

	chainID := basecoinApp.GetState().GetChainID()
	if withoutTendermintFlag {
		log.Notice("Starting Basecoin without Tendermint", "chain_id", chainID)
		// run just the abci app/server
		startBasecoinABCI(basecoinApp)
	} else {
		log.Notice("Starting Basecoin with Tendermint", "chain_id", chainID)
		// start the app with tendermint in-process
		startTendermint(basecoinDir, basecoinApp)
	}
}

func startBasecoinABCI(basecoinApp *app.Basecoin) {

	// Start the ABCI listener
	svr, err := server.NewServer(addrFlag, "socket", basecoinApp)
	if err != nil {
		cmn.Exit(fmt.Sprintf("Error creating listener: %+v\n", err))
	}

	// Wait forever
	cmn.TrapSignal(func() {
		// Cleanup
		svr.Stop()
	})
}

func startTendermint(dir string, basecoinApp *app.Basecoin) {

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
		cmn.Exit(fmt.Sprintf("%+v\n", err))
	}

	// Wait forever
	cmn.TrapSignal(func() {
		// Cleanup
		n.Stop()
	})
}
