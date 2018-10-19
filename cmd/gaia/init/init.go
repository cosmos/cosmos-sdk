package init

import (
	"encoding/json"
	"errors"
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
	flagWithTxs      = "with-txs"
	flagOverwrite    = "overwrite"
	flagClientHome   = "home-client"
	flagOverwriteKey = "overwrite-key"
	flagSkipGenesis  = "skip-genesis"
	flagMoniker      = "moniker"
)

type initConfig struct {
	ChainID      string
	GenTxsDir    string
	Name         string
	NodeID       string
	ClientHome   string
	WithTxs      bool
	Overwrite    bool
	OverwriteKey bool
	ValPubKey    crypto.PubKey
}

type printInfo struct {
	Moniker    string          `json:"moniker"`
	ChainID    string          `json:"chain_id"`
	NodeID     string          `json:"node_id"`
	AppMessage json.RawMessage `json:"app_message"`
}

// nolint: errcheck
func displayInfo(cdc *codec.Codec, info printInfo) error {
	out, err := codec.MarshalJSONIndent(cdc, info)
	if err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "%s\n", string(out))
	return nil
}

// get cmd to initialize all files for tendermint and application
// nolint
func InitCmd(ctx *server.Context, cdc *codec.Codec, appInit server.AppInit) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize private validator, p2p, genesis, and application configuration files",
		Long: `Initialize validators's and node's configuration files.

Note that only node's configuration files will be written if the flag --skip-genesis is
enabled, and the genesis file will not be generated.
`,
		Args: cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			config := ctx.Config
			config.SetRoot(viper.GetString(cli.HomeFlag))

			name := viper.GetString(client.FlagName)
			chainID := viper.GetString(client.FlagChainID)
			if chainID == "" {
				chainID = fmt.Sprintf("test-chain-%v", common.RandStr(6))
			}
			nodeID, valPubKey, err := InitializeNodeValidatorFiles(config)
			if err != nil {
				return err
			}

			if viper.GetString(flagMoniker) != "" {
				config.Moniker = viper.GetString(flagMoniker)
			}
			if config.Moniker == "" && name != "" {
				config.Moniker = name
			}
			toPrint := printInfo{
				ChainID: chainID,
				Moniker: config.Moniker,
				NodeID:  nodeID,
			}
			if viper.GetBool(flagSkipGenesis) {
				cfg.WriteConfigFile(filepath.Join(config.RootDir, "config", "config.toml"), config)
				return displayInfo(cdc, toPrint)
			}

			initCfg := initConfig{
				ChainID:      chainID,
				GenTxsDir:    filepath.Join(config.RootDir, "config", "gentx"),
				Name:         name,
				NodeID:       nodeID,
				ClientHome:   viper.GetString(flagClientHome),
				WithTxs:      viper.GetBool(flagWithTxs),
				Overwrite:    viper.GetBool(flagOverwrite),
				OverwriteKey: viper.GetBool(flagOverwriteKey),
				ValPubKey:    valPubKey,
			}
			appMessage, err := initWithConfig(cdc, config, initCfg)
			// print out some key information
			if err != nil {
				return err
			}

			toPrint.AppMessage = appMessage
			return displayInfo(cdc, toPrint)
		},
	}

	cmd.Flags().String(cli.HomeFlag, app.DefaultNodeHome, "node's home directory")
	cmd.Flags().BoolP(flagOverwrite, "o", false, "overwrite the genesis.json file")
	cmd.Flags().String(client.FlagChainID, "", "genesis file chain-id, if left blank will be randomly created")
	cmd.Flags().Bool(flagWithTxs, false, "apply existing genesis transactions from [--home]/config/gentx/")
	cmd.Flags().String(client.FlagName, "", "name of private key with which to sign the gentx")
	cmd.Flags().String(flagMoniker, "", "overrides --name flag and set the validator's moniker to a different value; ignored if it runs without the --with-txs flag")
	cmd.Flags().String(flagClientHome, app.DefaultCLIHome, "client's home directory")
	cmd.Flags().Bool(flagOverwriteKey, false, "overwrite client's key")
	cmd.Flags().Bool(flagSkipGenesis, false, "do not create genesis.json")
	return cmd
}

// InitializeNodeValidatorFiles creates private validator and p2p configuration files.
func InitializeNodeValidatorFiles(config *cfg.Config) (nodeID string, valPubKey crypto.PubKey, err error) {
	nodeKey, err := p2p.LoadOrGenNodeKey(config.NodeKeyFile())
	if err != nil {
		return
	}
	nodeID = string(nodeKey.ID())
	valPubKey = ReadOrCreatePrivValidator(config.PrivValidatorFile())
	return
}

func initWithConfig(cdc *codec.Codec, config *cfg.Config, initCfg initConfig) (
	appMessage json.RawMessage, err error) {
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
	chainID := initCfg.ChainID

	if initCfg.WithTxs {
		_, appGenTxs, persistentPeers, err = app.CollectStdTxs(config.Moniker, initCfg.GenTxsDir, cdc)
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

		if initCfg.Name == "" {
			err = errors.New("must specify validator's moniker (--name)")
			return
		}

		config.Moniker = initCfg.Name
		ip, err = server.ExternalIP()
		if err != nil {
			return
		}
		memo := fmt.Sprintf("%s@%s:26656", initCfg.NodeID, ip)
		buf := client.BufferStdin()
		prompt := fmt.Sprintf("Password for account %q (default: %q):", initCfg.Name, app.DefaultKeyPass)
		keyPass, err = client.GetPassword(prompt, buf)
		if err != nil && keyPass != "" {
			// An error was returned that either failed to read the password from
			// STDIN or the given password is not empty but failed to meet minimum
			// length requirements.
			return
		}
		if keyPass == "" {
			keyPass = app.DefaultKeyPass
		}

		addr, secret, err = server.GenerateSaveCoinKey(initCfg.ClientHome, initCfg.Name, keyPass, initCfg.OverwriteKey)
		if err != nil {
			return
		}
		appMessage, err = json.Marshal(map[string]string{"secret": secret})
		if err != nil {
			return
		}

		msg := stake.NewMsgCreateValidator(
			sdk.ValAddress(addr),
			initCfg.ValPubKey,
			sdk.NewInt64Coin("steak", 100),
			stake.NewDescription(config.Moniker, "", "", ""),
			stake.NewCommissionMsg(sdk.ZeroDec(), sdk.ZeroDec(), sdk.ZeroDec()),
		)
		txBldr := authtx.NewTxBuilderFromCLI().WithCodec(cdc).WithMemo(memo).WithChainID(chainID)
		signedTx, err = txBldr.SignStdTx(
			initCfg.Name, keyPass, auth.NewStdTx([]sdk.Msg{msg}, auth.StdFee{}, []auth.StdSignature{}, memo), false,
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
	err = WriteGenesisFile(genFile, chainID, nil, appState)

	return
}

// WriteGenesisFile creates and writes the genesis configuration to disk. An
// error is returned if building or writing the configuration to file fails.
// nolint: unparam
func WriteGenesisFile(genesisFile, chainID string, validators []types.GenesisValidator, appState json.RawMessage) error {
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
func ReadOrCreatePrivValidator(privValFile string) crypto.PubKey {
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
