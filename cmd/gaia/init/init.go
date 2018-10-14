package init

import (
	"encoding/json"
	"fmt"
	"github.com/cosmos/cosmos-sdk/cmd/gaia/app"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/privval"
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

const (
	flagChainID = "chain-id"
	flagWithTxs = "with-txs"
	flagMoniker = "moniker"
	flagOverwrite = "overwrite"
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
			chainID := viper.GetString(flagChainID)
			if chainID == "" {
				chainID = fmt.Sprintf("test-chain-%v", common.RandStr(6))
			}
			genTxsDir := filepath.Join(config.RootDir, "config", "gentx")
			moniker := viper.GetString(flagMoniker)
			withTxs := viper.GetBool(flagWithTxs)
			overwriteGenesis := viper.GetBool(flagOverwrite)
			nodeID, appMessage, err := initWithConfig(cdc, appInit, config, chainID, moniker, genTxsDir, withTxs, overwriteGenesis)
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

	cmd.Flags().BoolP(flagOverwrite, "o", false, "overwrite the genesis.json file")
	cmd.Flags().String(flagChainID, "", "genesis file chain-id, if left blank will be randomly created")
	cmd.Flags().Bool(flagWithTxs, false, "apply existing genesis transactions from [--home]/config/gentx/")
	cmd.Flags().String(flagMoniker, "", "moniker")
	//cmd.Flags().AddFlagSet(appInit.FlagsAppGenState)
	return cmd
}

func initWithConfig(cdc *codec.Codec, appInit server.AppInit, config *cfg.Config, chainID, moniker, genTxsDir string, withGenTxs, overwriteGenesis bool) (
	nodeID string, appMessage json.RawMessage, err error) {
	nodeKey, err := p2p.LoadOrGenNodeKey(config.NodeKeyFile())
	if err != nil {
		return
	}
	nodeID = string(nodeKey.ID())

	genFile := config.GenesisFile()
	if !overwriteGenesis && common.FileExists(genFile) {
		err = fmt.Errorf("genesis.json file already exists: %v", genFile)
		return
	}

	// process genesis transactions, or otherwise create one for defaults
	var appGenTxs []auth.StdTx
	var persistentPeers string
	var genTxs []json.RawMessage
	var appState json.RawMessage
	var validators []types.GenesisValidator

	if withGenTxs {
		_, appGenTxs, persistentPeers, err = app.ProcessStdTxs(moniker, genTxsDir, cdc)
		if err != nil {
			return
		}
		genTxs = make([]json.RawMessage, len(appGenTxs))
		config.P2P.PersistentPeers = persistentPeers
		for i, stdTx := range appGenTxs {
			var jsonRawTx json.RawMessage
			jsonRawTx, err = cdc.MarshalJSON(stdTx)
			if err != nil {
				return
			}
			genTxs[i] = jsonRawTx
		}

		appState, err = app.GaiaAppGenStateJSON(cdc, genTxs)
		if err != nil {
			return
		}
	} else {
		var genesisState app.GenesisState
		pubKey := readOrCreatePrivValidator(config)
		config.Moniker = moniker
		genesisState, genValidator := app.DefaultState(config.Moniker, pubKey)
		appState, err = codec.MarshalJSONIndent(cdc, genesisState)
		if err != nil {
			return
		}
		validators = []types.GenesisValidator{genValidator}
	}

	cfg.WriteConfigFile(filepath.Join(config.RootDir, "config", "config.toml"), config)
	err = writeGenesisFile(cdc, genFile, chainID, validators, appState)
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
