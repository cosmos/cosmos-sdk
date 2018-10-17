package init

import (
	"encoding/json"
	"fmt"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/cmd/gaia/app"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/client/txbuilder"
	"github.com/cosmos/cosmos-sdk/x/stake"
	"github.com/tendermint/tendermint/crypto"
	"github.com/tendermint/tendermint/privval"
	"os"
	"path/filepath"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	cfg "github.com/tendermint/tendermint/config"
	"github.com/tendermint/tendermint/libs/cli"
	"github.com/tendermint/tendermint/libs/common"
	"github.com/tendermint/tendermint/p2p"
	"github.com/tendermint/tendermint/types"
)

const (
	flagChainID              = "chain-id"
	flagWithTxs              = "with-txs"
	flagOverwrite            = "overwrite"
	flagClientHome           = "home-client"
	flagOWK                  = "owk"
)

type InitConfig struct {
	ChainID string
	GenTxsDir string
	Moniker string
	ClientHome string
	WithTxs bool
	Overwrite bool
	OverwriteKeys bool
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
			chainID := viper.GetString(flagChainID)
			if chainID == "" {
				chainID = fmt.Sprintf("test-chain-%v", common.RandStr(6))
			}
			initCfg := InitConfig{
				ChainID: chainID,
				GenTxsDir: viper.GetString(filepath.Join(config.RootDir, "config", "gentx")),
				Moniker: viper.GetString(client.FlagName),
				ClientHome: viper.GetString(flagClientHome),
				WithTxs: viper.GetBool(flagWithTxs),
				Overwrite: viper.GetBool(flagOverwrite),
				OverwriteKeys: viper.GetBool(flagOWK),
			}
			nodeID, appMessage, err := initWithConfig(cdc, config, initCfg)
			// print out some key information
			if err != nil {
				return err
			}

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
			fmt.Fprintf(os.Stderr, "%s\n", string(out))
			return nil
		},
	}

	cmd.Flags().String(cli.HomeFlag, app.DefaultNodeHome, "node's home directory")
	cmd.Flags().BoolP(flagOverwrite, "o", false, "overwrite the genesis.json file")
	cmd.Flags().String(flagChainID, "", "genesis file chain-id, if left blank will be randomly created")
	cmd.Flags().Bool(flagWithTxs, false, "apply existing genesis transactions from [--home]/config/gentx/")
	cmd.Flags().String(client.FlagName, "", "moniker")
	cmd.Flags().String(flagClientHome, app.DefaultCLIHome, "client's home directory")
	cmd.Flags().Bool(flagOWK, false, "overwrite client's keys")
	return cmd
}

func initWithConfig(cdc *codec.Codec, config *cfg.Config, initCfg InitConfig) (
	nodeID string, appMessage json.RawMessage, err error) {
	nodeKey, err := p2p.LoadOrGenNodeKey(config.NodeKeyFile())
	if err != nil {
		return
	}
	nodeID = string(nodeKey.ID())
	privValFile := config.PrivValidatorFile()
	valPubKey := readOrCreatePrivValidator(privValFile)
	genFile := config.GenesisFile()
	if !initCfg.Overwrite && common.FileExists(genFile) {
		err = fmt.Errorf("genesis.json file already exists: %v", genFile)
		return
	}

	// process genesis transactions, else create default genesis.json
	var appGenTxs []auth.StdTx
	var persistentPeers string
	var genTxs []json.RawMessage
	var appState json.RawMessage
	var jsonRawTx json.RawMessage
	moniker := initCfg.Moniker
	chainID := initCfg.ChainID

	if initCfg.WithTxs {
		_, appGenTxs, persistentPeers, err = app.ProcessStdTxs(moniker, initCfg.GenTxsDir, cdc)
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
	} else {
		var ip, keyPass, secret string
		var addr sdk.AccAddress
		var signedTx auth.StdTx

		config.Moniker = moniker
		ip, err = server.ExternalIP()
		if err != nil {
			return
		}
		memo := fmt.Sprintf("%s@%s:26656", nodeID, ip)
		buf := client.BufferStdin()
		prompt := fmt.Sprintf("Password for account '%s' (default %s):", moniker, server.DefaultKeyPass)
		keyPass, err = client.GetPassword(prompt, buf)
		if err != nil && keyPass != "" {
			// An error was returned that either failed to read the password from
			// STDIN or the given password is not empty but failed to meet minimum
			// length requirements.
			return
		}
		if keyPass == "" {
			keyPass = server.DefaultKeyPass
		}

		addr, secret, err = server.GenerateSaveCoinKey(initCfg.ClientHome, moniker, keyPass, initCfg.OverwriteKeys)
		if err != nil {
			return
		}
		appMessage, err = json.Marshal(map[string]string{"secret": secret})
		if err != nil {
			return
		}

		msg := stake.NewMsgCreateValidator(
			sdk.ValAddress(addr),
			valPubKey,
			sdk.NewInt64Coin("steak", 100),
			stake.NewDescription(moniker, "", "", ""),
			stake.NewCommissionMsg(sdk.ZeroDec(), sdk.ZeroDec(), sdk.ZeroDec()),
		)
		txBldr := authtx.NewTxBuilderFromCLI().WithCodec(cdc).WithMemo(memo).WithChainID(chainID)
		signedTx, err = txBldr.SignStdTx(
			moniker, keyPass, auth.NewStdTx([]sdk.Msg{msg}, auth.StdFee{}, []auth.StdSignature{}, memo), false,
		)
		if err != nil {
			return
		}
		jsonRawTx, err = cdc.MarshalJSON(signedTx)
		if err != nil {
			return
		}
		genTxs = []json.RawMessage{jsonRawTx}
	}

	cfg.WriteConfigFile(filepath.Join(config.RootDir, "config", "config.toml"), config)
	appState, err = app.GaiaAppGenStateJSON(cdc, genTxs)
	if err != nil {
		return
	}
	err = writeGenesisFile(cdc, genFile, chainID, nil, appState)

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
func readOrCreatePrivValidator(privValFile string) crypto.PubKey {
	// private validator
	var privValidator *privval.FilePV
	if common.FileExists(privValFile) {
		privValidator = privval.LoadFilePV(privValFile)
	} else {
		privValidator = privval.GenFilePV(privValFile)
		privValidator.Save()
	}
	return privValidator.GetPubKey()
}
