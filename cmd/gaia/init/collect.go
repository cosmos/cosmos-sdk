package init

// DONTCOVER

import (
	"encoding/json"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	cfg "github.com/tendermint/tendermint/config"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/libs/cli"
	"github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/cmd/gaia/app"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/x/auth"
)

const (
	flagGenTxDir = "gentx-dir"
)

type initConfig struct {
	ChainID   string
	GenTxsDir string
	Name      string
	NodeID    string
	ValPubKey crypto.PubKey
}

// nolint
func CollectGenTxsCmd(ctx *server.Context, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "collect-gentxs",
		Short: "Collect genesis txs and output a genesis.json file",
		RunE: func(_ *cobra.Command, _ []string) error {
			config := ctx.Config
			config.SetRoot(viper.GetString(cli.HomeFlag))
			name := viper.GetString(client.FlagName)
			nodeID, valPubKey, err := InitializeNodeValidatorFiles(config)
			if err != nil {
				return err
			}

			genDoc, err := LoadGenesisDoc(cdc, config.GenesisFile())
			if err != nil {
				return err
			}

			genTxsDir := viper.GetString(flagGenTxDir)
			if genTxsDir == "" {
				genTxsDir = filepath.Join(config.RootDir, "config", "gentx")
			}

			toPrint := newPrintInfo(config.Moniker, genDoc.ChainID, nodeID, genTxsDir, json.RawMessage(""))
			initCfg := newInitConfig(genDoc.ChainID, genTxsDir, name, nodeID, valPubKey)

			appMessage, err := genAppStateFromConfig(cdc, config, initCfg, genDoc)
			if err != nil {
				return err
			}

			toPrint.AppMessage = appMessage

			// print out some key information
			return displayInfo(cdc, toPrint)
		},
	}

	cmd.Flags().String(cli.HomeFlag, app.DefaultNodeHome, "node's home directory")
	cmd.Flags().String(flagGenTxDir, "",
		"override default \"gentx\" directory from which collect and execute "+
			"genesis transactions; default [--home]/config/gentx/")
	return cmd
}

func genAppStateFromConfig(
	cdc *codec.Codec, config *cfg.Config, initCfg initConfig, genDoc types.GenesisDoc,
) (appState json.RawMessage, err error) {

	genFile := config.GenesisFile()
	var (
		appGenTxs       []auth.StdTx
		persistentPeers string
		genTxs          []json.RawMessage
		jsonRawTx       json.RawMessage
	)

	// process genesis transactions, else create default genesis.json
	appGenTxs, persistentPeers, err = app.CollectStdTxs(
		cdc, config.Moniker, initCfg.GenTxsDir, genDoc,
	)
	if err != nil {
		return
	}

	genTxs = make([]json.RawMessage, len(appGenTxs))
	config.P2P.PersistentPeers = persistentPeers

	for i, stdTx := range appGenTxs {
		jsonRawTx, err = cdc.MarshalJSON(stdTx)
		if err != nil {
			return
		}
		genTxs[i] = jsonRawTx
	}

	cfg.WriteConfigFile(filepath.Join(config.RootDir, "config", "config.toml"), config)

	appState, err = app.GaiaAppGenStateJSON(cdc, genDoc, genTxs)
	if err != nil {
		return
	}

	err = ExportGenesisFile(genFile, initCfg.ChainID, nil, appState)
	return
}

func newInitConfig(chainID, genTxsDir, name, nodeID string,
	valPubKey crypto.PubKey) initConfig {

	return initConfig{
		ChainID:   chainID,
		GenTxsDir: genTxsDir,
		Name:      name,
		NodeID:    nodeID,
		ValPubKey: valPubKey,
	}
}

func newPrintInfo(moniker, chainID, nodeID, genTxsDir string,
	appMessage json.RawMessage) printInfo {

	return printInfo{
		Moniker:    moniker,
		ChainID:    chainID,
		NodeID:     nodeID,
		GenTxsDir:  genTxsDir,
		AppMessage: appMessage,
	}
}
