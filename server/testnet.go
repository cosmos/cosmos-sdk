/*
 * Ported from Tendermint
 */
package server

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	cfg "github.com/tendermint/tendermint/config"
	gaiacfg "github.com/cosmos/cosmos-sdk/config"
	"github.com/tendermint/tendermint/p2p"
	"github.com/tendermint/tendermint/types"
	pvm "github.com/tendermint/tendermint/types/priv_validator"
	cmn "github.com/tendermint/tmlibs/common"
	"github.com/cosmos/cosmos-sdk/wire"
)

var (
	nValidators    int
	nNonValidators int
	outputDir      string
	nodeDirPrefix  string

	populatePersistentPeers bool
	hostnamePrefix          string
	startingIPAddress       string
	p2pPort                 int
)

const (
	nodeDirPerm = 0755
)

func init() {
}

// get cmd to initialize all files for tendermint testnet and application
func TestnetFilesCmd(ctx *Context, cdc *wire.Codec, appInit AppInit) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "testnet",
		Short: "Initialize files for a Gaiad testnet",
		Long: `testnet will create "v" + "n" number of directories and populate each with
necessary files (private validator, genesis, config, etc.).

Note, strict routability for addresses is turned off in the config file.

Optionally, it will fill in persistent_peers list in config file using either hostnames or IPs.

Example:

	gaiad testnet --v 4 --o ./output --populate-persistent-peers --starting-ip-address 192.168.10.2
	`,
		RunE: func(_ *cobra.Command, _ []string) error {
			config := ctx.Config
			genVals := make([]types.GenesisValidator, nValidators)

			// Generate private key, node ID, initial transaction
			for i := 0; i < nValidators; i++ {
				nodeDirName := cmn.Fmt("%s%d", nodeDirPrefix, i)
				nodeDir := filepath.Join(outputDir, nodeDirName, "gaiad")
				clientDir := filepath.Join(outputDir, nodeDirName, "gaiacli")
				gentxsDir := filepath.Join(outputDir, "gentxs")
				config.SetRoot(nodeDir)

				err := os.MkdirAll(filepath.Join(nodeDir, "config"), nodeDirPerm)
				if err != nil {
					_ = os.RemoveAll(outputDir)
					return err
				}

				err = os.MkdirAll(clientDir, nodeDirPerm)
				if err != nil {
					_ = os.RemoveAll(outputDir)
					return err
				}

				config.Moniker = nodeDirName
				gaiaConfig := gaiacfg.DefaultConfig()
				gaiaConfig.Name = nodeDirName
				gaiaConfig.CliRoot = clientDir
				gaiaConfig.GenTx.Overwrite = true

				// Run `init gen-tx` and generate initial transactions
				cliPrint, genTxFile, err := gentxWithConfig(config, gaiaConfig, ctx, cdc, appInit)
				if err != nil {
					return err
				}

				// Save private key seed words
				name := fmt.Sprintf("%v.json", "key_seed")
				writePath := filepath.Join(clientDir)
				file := filepath.Join(writePath, name)
				err = cmn.EnsureDir(writePath, 0700)
				if err != nil {
					return err
				}
				err = cmn.WriteFile(file, cliPrint, 0600)
				if err != nil {
					return err
				}

				// Gather gentxs folder
				name = fmt.Sprintf("%v.json", nodeDirName)
				writePath = filepath.Join(gentxsDir)
				file = filepath.Join(writePath, name)
				err = cmn.EnsureDir(writePath, 0700)
				if err != nil {
					return err
				}
				err = cmn.WriteFile(file, genTxFile, 0644)
				if err != nil {
					return err
				}
			}

			// Generate genesis.json
			for i := 0; i < nValidators; i++ {

				nodeDirName := cmn.Fmt("%s%d", nodeDirPrefix, i)
				nodeDir := filepath.Join(outputDir, nodeDirName, "gaiad")
				clientDir := filepath.Join(outputDir, nodeDirName, "gaiacli")
				gentxsDir := filepath.Join(outputDir, "gentxs")
				gaiaConfig := gaiacfg.DefaultConfig()
				gaiaConfig.Name = nodeDirName
				gaiaConfig.CliRoot = clientDir
				gaiaConfig.Init.ChainID = "something-something"
				gaiaConfig.Init.GenTxs = true
				gaiaConfig.Init.Overwrite = true
				gaiaConfig.Init.GenTxsDir = gentxsDir
				config.SetRoot(nodeDir)

				// Run `init` and generate genesis.json
				_, _, _, err := initWithConfig(config,gaiaConfig, ctx,cdc,appInit)
				if err != nil {
					return err
				}

				pvFile := filepath.Join(nodeDir, config.BaseConfig.PrivValidator)
				pv := pvm.LoadFilePV(pvFile)
				genVals[i] = types.GenesisValidator{
					PubKey: pv.GetPubKey(),
					Power:  1,
					Name:   nodeDirName,
				}
			}

			for i := 0; i < nNonValidators; i++ {
				nodeDirName := cmn.Fmt("%s%d", nodeDirPrefix, i+nValidators)
				nodeDir := filepath.Join(outputDir, nodeDirName, "gaiad")
				config.SetRoot(nodeDir)

				err := os.MkdirAll(filepath.Join(nodeDir, "config"), nodeDirPerm)
				if err != nil {
					_ = os.RemoveAll(outputDir)
					return err
				}

				config.Moniker = nodeDirName
				//TODO: Fix it!v
				initWithConfig(config,nil,ctx,cdc,appInit)
			}

			if populatePersistentPeers {
				err := populatePersistentPeersInConfigAndWriteIt(config)
				if err != nil {
					_ = os.RemoveAll(outputDir)
					return err
				}
			}

			fmt.Printf("Successfully initialized %v node directories\n", nValidators+nNonValidators)
			return nil
			},
	}
	cmd.Flags().IntVar(&nValidators, "v", 4,
		"Number of validators to initialize the testnet with")
	cmd.Flags().IntVar(&nNonValidators, "n", 0,
		"Number of non-validators to initialize the testnet with")
	cmd.Flags().StringVar(&outputDir, "o", "./mytestnet",
		"Directory to store initialization data for the testnet")
	cmd.Flags().StringVar(&nodeDirPrefix, "node-dir-prefix", "node",
		"Prefix the directory name for each node with (node results in node0, node1, ...)")

	cmd.Flags().BoolVar(&populatePersistentPeers, "populate-persistent-peers", false,
		"Update config of each node with the list of persistent peers build using either hostname-prefix or starting-ip-address")
	cmd.Flags().StringVar(&hostnamePrefix, "hostname-prefix", "node",
		"Hostname prefix (node results in persistent peers list ID0@node0:46656, ID1@node1:46656, ...)")
	cmd.Flags().StringVar(&startingIPAddress, "starting-ip-address", "",
		"Starting IP address (192.168.0.1 results in persistent peers list ID0@192.168.0.1:46656, ID1@192.168.0.2:46656, ...)")
	cmd.Flags().IntVar(&p2pPort, "p2p-port", 46656,
		"P2P Port")
	cmd.AddCommand(GenTxCmd(ctx, cdc, appInit))
	return cmd
}

func hostnameOrIP(i int) string {
	if startingIPAddress != "" {
		ip := net.ParseIP(startingIPAddress)
		ip = ip.To4()
		if ip == nil {
			fmt.Printf("%v: non ipv4 address\n", startingIPAddress)
			os.Exit(1)
		}

		for j := 0; j < i; j++ {
			ip[3]++
		}
		return ip.String()
	}

	return fmt.Sprintf("%s%d", hostnamePrefix, i)
}

func populatePersistentPeersInConfigAndWriteIt(config *cfg.Config) error {
	persistentPeers := make([]string, nValidators+nNonValidators)
	for i := 0; i < nValidators+nNonValidators; i++ {
		nodeKey, err := p2p.LoadNodeKey(config.NodeKeyFile())
		if err != nil {
			return err
		}
		persistentPeers[i] = p2p.IDAddressString(nodeKey.ID(), fmt.Sprintf("%s:%d", hostnameOrIP(i), p2pPort))
	}
	persistentPeersList := strings.Join(persistentPeers, ",")

	for i := 0; i < nValidators+nNonValidators; i++ {
		config.P2P.PersistentPeers = persistentPeersList
		config.P2P.AddrBookStrict = false

		// overwrite default config
		cfg.WriteConfigFile(filepath.Join(config.RootDir, "config", "config.toml"), config)
	}

	return nil
}
