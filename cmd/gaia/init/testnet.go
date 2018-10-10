package init

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/server"
	"net"
	"path/filepath"

	"github.com/spf13/cobra"

	gc "github.com/cosmos/cosmos-sdk/server/config"

	"os"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/spf13/viper"
	cfg "github.com/tendermint/tendermint/config"
	cmn "github.com/tendermint/tendermint/libs/common"
)

var (
	nodeDirPrefix  = "node-dir-prefix"
	nValidators    = "v"
	outputDir      = "output-dir"
	nodeDaemonHome = "node-daemon-home"
	nodeCliHome    = "node-cli-home"

	startingIPAddress = "starting-ip-address"
)

const nodeDirPerm = 0755

// get cmd to initialize all files for tendermint testnet and application
func TestnetFilesCmd(ctx *server.Context, cdc *codec.Codec, appInit server.AppInit) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "testnet",
		Short: "Initialize files for a Gaiad testnet",
		Long: `testnet will create "v" number of directories and populate each with
necessary files (private validator, genesis, config, etc.).

Note, strict routability for addresses is turned off in the config file.

Example:

	gaiad testnet --v 4 --o ./output --starting-ip-address 192.168.10.2
	`,
		RunE: func(_ *cobra.Command, _ []string) error {
			config := ctx.Config
			err := testnetWithConfig(config, cdc, appInit)
			return err
		},
	}
	cmd.Flags().Int(nValidators, 4,
		"Number of validators to initialize the testnet with")
	cmd.Flags().StringP(outputDir, "o", "./mytestnet",
		"Directory to store initialization data for the testnet")
	cmd.Flags().String(nodeDirPrefix, "node",
		"Prefix the directory name for each node with (node results in node0, node1, ...)")
	cmd.Flags().String(nodeDaemonHome, "gaiad",
		"Home directory of the node's daemon configuration")
	cmd.Flags().String(nodeCliHome, "gaiacli",
		"Home directory of the node's cli configuration")

	cmd.Flags().String(startingIPAddress, "192.168.0.1",
		"Starting IP address (192.168.0.1 results in persistent peers list ID0@192.168.0.1:46656, ID1@192.168.0.2:46656, ...)")
	return cmd
}

func testnetWithConfig(config *cfg.Config, cdc *codec.Codec, appInit server.AppInit) error {
	outDir := viper.GetString(outputDir)
	numValidators := viper.GetInt(nValidators)

	// Generate private key, node ID, initial transaction
	for i := 0; i < numValidators; i++ {
		nodeDirName := fmt.Sprintf("%s%d", viper.GetString(nodeDirPrefix), i)
		nodeDaemonHomeName := viper.GetString(nodeDaemonHome)
		nodeCliHomeName := viper.GetString(nodeCliHome)
		nodeDir := filepath.Join(outDir, nodeDirName, nodeDaemonHomeName)
		clientDir := filepath.Join(outDir, nodeDirName, nodeCliHomeName)
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
		ip, err := getIP(i)
		if err != nil {
			return err
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
		err = writeFile(name, clientDir, cliPrint)
		if err != nil {
			return err
		}

		// Gather gentxs folder
		name = fmt.Sprintf("%v.json", nodeDirName)
		err = writeFile(name, gentxsDir, genTxFile)
		if err != nil {
			return err
		}
	}

	// Generate genesis.json and config.toml
	chainID := "chain-" + cmn.RandStr(6)
	for i := 0; i < numValidators; i++ {

		nodeDirName := fmt.Sprintf("%s%d", viper.GetString(nodeDirPrefix), i)
		nodeDaemonHomeName := viper.GetString(nodeDaemonHome)
		nodeDir := filepath.Join(outDir, nodeDirName, nodeDaemonHomeName)
		gentxsDir := filepath.Join(outDir, "gentxs")
		initConfig := server.InitConfig{
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

func getIP(i int) (ip string, err error) {
	ip = viper.GetString(startingIPAddress)
	if len(ip) == 0 {
		ip, err = server.ExternalIP()
		if err != nil {
			return "", err
		}
	} else {
		ip, err = calculateIP(ip, i)
		if err != nil {
			return "", err
		}
	}
	return ip, nil
}

func writeFile(name string, dir string, contents []byte) error {
	writePath := filepath.Join(dir)
	file := filepath.Join(writePath, name)
	err := cmn.EnsureDir(writePath, 0700)
	if err != nil {
		return err
	}
	err = cmn.WriteFile(file, contents, 0600)
	if err != nil {
		return err
	}
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
