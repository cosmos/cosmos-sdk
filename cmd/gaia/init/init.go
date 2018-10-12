package init

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"sort"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/server"
	servercfg "github.com/cosmos/cosmos-sdk/server/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	cfg "github.com/tendermint/tendermint/config"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/libs/cli"
	"github.com/tendermint/tendermint/libs/common"
	"github.com/tendermint/tendermint/p2p"
	"github.com/tendermint/tendermint/privval"
	"github.com/tendermint/tendermint/types"
)

// get cmd to initialize all files for tendermint and application
func GenTxCmd(ctx *server.Context, cdc *codec.Codec, appInit server.AppInit) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gen-tx",
		Short: "Create genesis transaction file (under [--home]/config/gentx/gentx-[nodeID].json)",
		Args:  cobra.NoArgs,
		RunE: func(_ *cobra.Command, args []string) error {
			config := ctx.Config
			config.SetRoot(viper.GetString(cli.HomeFlag))

			ip := viper.GetString(server.FlagIP)
			if len(ip) == 0 {
				eip, err := server.ExternalIP()
				if err != nil {
					return err
				}
				ip = eip
			}

			genTxConfig := servercfg.GenTx{
				viper.GetString(server.FlagName),
				viper.GetString(server.FlagClientHome),
				viper.GetBool(server.FlagOWK),
				ip,
			}
			cliPrint, genTxFile, err := gentxWithConfig(cdc, appInit, config, genTxConfig)
			if err != nil {
				return err
			}
			toPrint := struct {
				AppMessage json.RawMessage `json:"app_message"`
				GenTxFile  json.RawMessage `json:"gen_tx_file"`
			}{
				cliPrint,
				genTxFile,
			}
			out, err := codec.MarshalJSONIndent(cdc, toPrint)
			if err != nil {
				return err
			}
			fmt.Println(string(out))
			return nil
		},
	}
	cmd.Flags().String(server.FlagIP, "", "external facing IP to use if left blank IP will be retrieved from this machine")
	cmd.Flags().AddFlagSet(appInit.FlagsAppGenTx)
	return cmd
}

// NOTE: This will update (write) the config file with
// updated name (moniker) for node.
func gentxWithConfig(cdc *codec.Codec, appInit server.AppInit, config *cfg.Config, genTxConfig servercfg.GenTx) (
	cliPrint json.RawMessage, genTxFile json.RawMessage, err error) {
	nodeKey, err := p2p.LoadOrGenNodeKey(config.NodeKeyFile())
	if err != nil {
		return
	}
	nodeID := string(nodeKey.ID())
	pubKey := readOrCreatePrivValidator(config)

	appGenTx, cliPrint, validator, err := appInit.AppGenTx(cdc, pubKey, genTxConfig)
	if err != nil {
		return
	}

	tx := server.GenesisTx{
		NodeID:    nodeID,
		IP:        genTxConfig.IP,
		Validator: validator,
		AppGenTx:  appGenTx,
	}
	bz, err := codec.MarshalJSONIndent(cdc, tx)
	if err != nil {
		return
	}
	genTxFile = json.RawMessage(bz)
	name := fmt.Sprintf("gentx-%v.json", nodeID)
	writePath := filepath.Join(config.RootDir, "config", "gentx")
	file := filepath.Join(writePath, name)
	err = common.EnsureDir(writePath, 0700)
	if err != nil {
		return
	}
	err = common.WriteFile(file, bz, 0644)
	if err != nil {
		return
	}

	// Write updated config with moniker
	config.Moniker = genTxConfig.Name
	configFilePath := filepath.Join(config.RootDir, "config", "config.toml")
	cfg.WriteConfigFile(configFilePath, config)

	return
}

// get cmd to initialize all files for tendermint and application
// nolint: golint
func InitCmd(ctx *server.Context, cdc *codec.Codec, appInit server.AppInit) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize genesis config, priv-validator file, and p2p-node file",
		Args:  cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {

			config := ctx.Config
			config.SetRoot(viper.GetString(cli.HomeFlag))
			initConfig := server.InitConfig{
				viper.GetString(server.FlagChainID),
				viper.GetBool(server.FlagWithTxs),
				filepath.Join(config.RootDir, "config", "gentx"),
				viper.GetBool(server.FlagOverwrite),
			}

			chainID, nodeID, appMessage, err := initWithConfig(cdc, appInit, config, initConfig)
			if err != nil {
				return err
			}
			// print out some key information
			toPrint := struct {
				ChainID    string          `json:"chain_id"`
				NodeID     string          `json:"node_id"`
				AppMessage json.RawMessage `json:"app_message"`
			}{
				chainID,
				nodeID,
				appMessage,
			}
			out, err := codec.MarshalJSONIndent(cdc, toPrint)
			if err != nil {
				return err
			}
			fmt.Println(string(out))
			return nil
		},
	}
	cmd.Flags().BoolP(server.FlagOverwrite, "o", false, "overwrite the genesis.json file")
	cmd.Flags().String(server.FlagChainID, "", "genesis file chain-id, if left blank will be randomly created")
	cmd.Flags().Bool(server.FlagWithTxs, false, "apply existing genesis transactions from [--home]/config/gentx/")
	cmd.Flags().AddFlagSet(appInit.FlagsAppGenState)
	cmd.Flags().AddFlagSet(appInit.FlagsAppGenTx) // need to add this flagset for when no GenTx's provided
	cmd.AddCommand(GenTxCmd(ctx, cdc, appInit))
	return cmd
}

func initWithConfig(cdc *codec.Codec, appInit server.AppInit, config *cfg.Config, initConfig server.InitConfig) (
	chainID string, nodeID string, appMessage json.RawMessage, err error) {
	nodeKey, err := p2p.LoadOrGenNodeKey(config.NodeKeyFile())
	if err != nil {
		return
	}
	nodeID = string(nodeKey.ID())
	pubKey := readOrCreatePrivValidator(config)

	if initConfig.ChainID == "" {
		initConfig.ChainID = fmt.Sprintf("test-chain-%v", common.RandStr(6))
	}
	chainID = initConfig.ChainID

	genFile := config.GenesisFile()
	if !initConfig.Overwrite && common.FileExists(genFile) {
		err = fmt.Errorf("genesis.json file already exists: %v", genFile)
		return
	}

	// process genesis transactions, or otherwise create one for defaults
	var appGenTxs []json.RawMessage
	var validators []types.GenesisValidator
	var persistentPeers string

	if initConfig.GenTxs {
		validators, appGenTxs, persistentPeers, err = processGenTxs(initConfig.GenTxsDir, cdc)
		if err != nil {
			return
		}
		config.P2P.PersistentPeers = persistentPeers
		configFilePath := filepath.Join(config.RootDir, "config", "config.toml")
		cfg.WriteConfigFile(configFilePath, config)
	} else {
		genTxConfig := servercfg.GenTx{
			viper.GetString(server.FlagName),
			viper.GetString(server.FlagClientHome),
			viper.GetBool(server.FlagOWK),
			"127.0.0.1",
		}

		// Write updated config with moniker
		config.Moniker = genTxConfig.Name
		configFilePath := filepath.Join(config.RootDir, "config", "config.toml")
		cfg.WriteConfigFile(configFilePath, config)
		appGenTx, am, validator, err := appInit.AppGenTx(cdc, pubKey, genTxConfig)
		appMessage = am
		if err != nil {
			return "", "", nil, err
		}
		validators = []types.GenesisValidator{validator}
		appGenTxs = []json.RawMessage{appGenTx}
	}

	appState, err := appInit.AppGenState(cdc, appGenTxs)
	if err != nil {
		return
	}

	err = writeGenesisFile(cdc, genFile, initConfig.ChainID, validators, appState)
	if err != nil {
		return
	}

	return
}

// append a genesis-piece
func processGenTxs(genTxsDir string, cdc *codec.Codec) (
	validators []types.GenesisValidator, appGenTxs []json.RawMessage, persistentPeers string, err error) {

	var fos []os.FileInfo
	fos, err = ioutil.ReadDir(genTxsDir)
	if err != nil {
		return
	}

	genTxs := make(map[string]server.GenesisTx)
	var nodeIDs []string
	for _, fo := range fos {
		filename := path.Join(genTxsDir, fo.Name())
		if !fo.IsDir() && (path.Ext(filename) != ".json") {
			continue
		}

		// get the genTx
		var bz []byte
		bz, err = ioutil.ReadFile(filename)
		if err != nil {
			return
		}
		var genTx server.GenesisTx
		err = cdc.UnmarshalJSON(bz, &genTx)
		if err != nil {
			return
		}

		genTxs[genTx.NodeID] = genTx
		nodeIDs = append(nodeIDs, genTx.NodeID)
	}

	sort.Strings(nodeIDs)

	for _, nodeID := range nodeIDs {
		genTx := genTxs[nodeID]

		// combine some stuff
		validators = append(validators, genTx.Validator)
		appGenTxs = append(appGenTxs, genTx.AppGenTx)

		// Add a persistent peer
		comma := ","
		if len(persistentPeers) == 0 {
			comma = ""
		}
		persistentPeers += fmt.Sprintf("%s%s@%s:26656", comma, genTx.NodeID, genTx.IP)
	}

	return
}

// read of create the private key file for this config
func readOrCreatePrivValidator(tmConfig *cfg.Config) crypto.PubKey {
	// private validator
	privValFile := tmConfig.PrivValidatorFile()
	var privValidator *privval.FilePV
	if common.FileExists(privValFile) {
		privValidator = privval.LoadFilePV(privValFile)
	} else {
		privValidator = privval.GenFilePV(privValFile)
		privValidator.Save()
	}
	return privValidator.GetPubKey()
}

// writeGenesisFile creates and writes the genesis configuration to disk. An
// error is returned if building or writing the configuration to file fails.
// nolint: unparam
func writeGenesisFile(cdc *codec.Codec, genesisFile, chainID string, validators []types.GenesisValidator, appState json.RawMessage) error {
	genDoc := types.GenesisDoc{
		ChainID:    chainID,
		Validators: validators,
		AppState:   appState,
	}

	if err := genDoc.ValidateAndComplete(); err != nil {
		return err
	}

	return genDoc.SaveAs(genesisFile)
}
