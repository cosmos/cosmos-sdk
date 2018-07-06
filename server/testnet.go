package server

import (
	"fmt"
	"net"
	"path/filepath"

	"github.com/spf13/cobra"

	gc "github.com/cosmos/cosmos-sdk/server/config"

	"os"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/spf13/viper"
	cfg "github.com/tendermint/tendermint/config"
	cmn "github.com/tendermint/tendermint/libs/common"
)

var (
	nodeDirPrefix = "node-dir-prefix"
	nValidators   = "v"
	outputDir     = "o"

	startingIPAddress = "starting-ip-address"
)

const nodeDirPerm = 0755

// get cmd to initialize all files for tendermint testnet and application
func TestnetFilesCmd(ctx *sdk.ServerContext, cdc *wire.Codec, appInit AppInit) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "testnet",
		Short: "Initialize files for a Gaiad testnet",
		Long: `testnet will create "v" number of directories and populate each with
necessary files (private validator, genesis, config, etc.).

Note, strict routability for addresses is turned off in the config file.

Example:

	gaiad testnet --v 4 --output-dir ./output --starting-ip-address 192.168.10.2
	`,
		RunE: func(_ *cobra.Command, _ []string) error {
			config := ctx.Config
			err := testnetWithConfig(config, cdc, appInit)
			return err
		},
	}
	cmd.Flags().Int(nValidators, 4,
		"Number of validators to initialize the testnet with")
	cmd.Flags().String(outputDir, "./mytestnet",
		"Directory to store initialization data for the testnet")
	cmd.Flags().String(nodeDirPrefix, "node",
		"Prefix the directory name for each node with (node results in node0, node1, ...)")

	cmd.Flags().String(startingIPAddress, "192.168.0.1",
		"Starting IP address (192.168.0.1 results in persistent peers list ID0@192.168.0.1:46656, ID1@192.168.0.2:46656, ...)")
	return cmd
}

func testnetWithConfig(config *cfg.Config, cdc *wire.Codec, appInit AppInit) error {

	outDir := viper.GetString(outputDir)
	// Generate private key, node ID, initial transaction
	for i := 0; i < viper.GetInt(nValidators); i++ {
		nodeDirName := fmt.Sprintf("%s%d", viper.GetString(nodeDirPrefix), i)
		nodeDir := filepath.Join(outDir, nodeDirName, "gaiad")
		clientDir := filepath.Join(outDir, nodeDirName, "gaiacli")
		gentxsDir := filepath.Join(outDir, "gentxs")
		config.SetRoot(nodeDir)

		err := os.MkdirAll(filepath.Join(nodeDir, "config"), nodeDirPerm)
		if err != nil {
			_ = os.RemoveAll(outDir)
			return err
		}

		err = os.MkdirAll(clientDir, nodeDirPerm)
		if err != nil {
			_ = os.RemoveAll(outDir)
			return err
		}

		config.Moniker = nodeDirName

		ip := viper.GetString(startingIPAddress)
		if len(ip) == 0 {
			ip, err = externalIP()
			if err != nil {
				return err
			}
		} else {
			ip, err = calculateIP(ip, i)
			if err != nil {
				return err
			}
		}

		genTxConfig := gc.GenTx{
			nodeDirName,
			clientDir,
			true,
			ip,
		}

		// Run `init gen-tx` and generate initial transactions
		cliPrint, genTxFile, err := gentxWithConfig(cdc, appInit, config, genTxConfig)
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

	// Generate genesis.json and config.toml
	chainID := "chain-" + cmn.RandStr(6)
	for i := 0; i < viper.GetInt(nValidators); i++ {

		nodeDirName := fmt.Sprintf("%s%d", viper.GetString(nodeDirPrefix), i)
		nodeDir := filepath.Join(outDir, nodeDirName, "gaiad")
		gentxsDir := filepath.Join(outDir, "gentxs")
		initConfig := InitConfig{
			chainID,
			true,
			gentxsDir,
			true,
		}
		config.Moniker = nodeDirName
		config.SetRoot(nodeDir)

		// Run `init` and generate genesis.json and config.toml
		_, _, _, err := initWithConfig(cdc, appInit, config, initConfig)
		if err != nil {
			return err
		}
	}

	fmt.Printf("Successfully initialized %v node directories\n", viper.GetInt(nValidators))
	return nil
}

func calculateIP(ip string, i int) (string, error) {
	ipv4 := net.ParseIP(ip).To4()
	if ipv4 == nil {
		return "", fmt.Errorf("%v: non ipv4 address", ip)
	}

	for j := 0; j < i; j++ {
		ipv4[3]++
	}
	return ipv4.String(), nil
}
