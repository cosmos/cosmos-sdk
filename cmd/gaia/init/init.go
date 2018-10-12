package init

import (
	"encoding/json"
	"fmt"
	"errors"
	"github.com/cosmos/cosmos-sdk/cmd/gaia/app"
	"github.com/cosmos/cosmos-sdk/x/auth"

	"path/filepath"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	cfg "github.com/tendermint/tendermint/config"
	"github.com/tendermint/tendermint/libs/cli"
	"github.com/tendermint/tendermint/libs/common"
	"github.com/tendermint/tendermint/p2p"
	"github.com/tendermint/tendermint/types"
)


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
				viper.GetString(server.FlagName),
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
	return cmd
}

func initWithConfig(cdc *codec.Codec, appInit server.AppInit, config *cfg.Config, initConfig server.InitConfig) (
	chainID string, nodeID string, appMessage json.RawMessage, err error) {
	nodeKey, err := p2p.LoadOrGenNodeKey(config.NodeKeyFile())
	if err != nil {
		return
	}
	nodeID = string(nodeKey.ID())
	//pubKey := readOrCreatePrivValidator(config)

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
	var appGenTxs []auth.StdTx
	var persistentPeers string
	var genTxs []json.RawMessage

	if initConfig.GenTxs {
		_, appGenTxs, persistentPeers, err = app.ProcessStdTxs(initConfig.Moniker, initConfig.GenTxsDir, cdc)
		if err != nil {
			return
		}
		genTxs = make([]json.RawMessage, len(appGenTxs))
		config.P2P.PersistentPeers = persistentPeers
		configFilePath := filepath.Join(config.RootDir, "config", "config.toml")
		cfg.WriteConfigFile(configFilePath, config)
		for i, stdTx := range appGenTxs {
			var jsonRawTx json.RawMessage
			jsonRawTx, err = cdc.MarshalJSON(stdTx)
			if err != nil {
				return
			}
			genTxs[i] = jsonRawTx
		}
	} else {
		panic(errors.New("WIP"))
		//genTxConfig := servercfg.GenTx{
		//	viper.GetString(server.FlagName),
		//	viper.GetString(server.FlagClientHome),
		//	viper.GetBool(server.FlagOWK),
		//	"127.0.0.1",
		//}
		//
		//// Write updated config with moniker
		//config.Moniker = genTxConfig.Name
		//configFilePath := filepath.Join(config.RootDir, "config", "config.toml")
		//cfg.WriteConfigFile(configFilePath, config)
		//_, am, err := appInit.AppGenTx(cdc, pubKey, genTxConfig)
		//appMessage = am
		//if err != nil {
		//	return "", "", nil, err
		//}
		//validators = []types.GenesisValidator{validator)
		//jsonMsg, err := json.Marshal(appGenTx)
		//if err != nil {
		//	return
		//}
		//appGenTxs = []json.RawMessage{jsonMsg}
	}

	appState, err := app.GaiaAppGenStateJSON(cdc, genTxs)
	if err != nil {
		return
	}
	
	err = writeGenesisFile(cdc, genFile, initConfig.ChainID, nil, appState)
	if err != nil {
		return
	}

	return
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
