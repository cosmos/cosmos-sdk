package server

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path"
	"path/filepath"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	crypto "github.com/tendermint/go-crypto"
	"github.com/tendermint/go-crypto/keys"
	"github.com/tendermint/go-crypto/keys/words"
	cfg "github.com/tendermint/tendermint/config"
	"github.com/tendermint/tendermint/p2p"
	tmtypes "github.com/tendermint/tendermint/types"
	pvm "github.com/tendermint/tendermint/types/priv_validator"
	cmn "github.com/tendermint/tmlibs/common"
	dbm "github.com/tendermint/tmlibs/db"
)

// genesis piece structure for creating combined genesis
type GenesisTx struct {
	NodeID    string                   `json:"node_id"`
	IP        string                   `json:"ip"`
	Validator tmtypes.GenesisValidator `json:"validator"`
	AppGenTx  json.RawMessage          `json:"app_gen_tx"`
}

var (
	flagOverwrite = "overwrite"
	flagGenTxs    = "gen-txs"
	flagIP        = "ip"
	flagChainID   = "ip"
)

// get cmd to initialize all files for tendermint and application
func GenTxCmd(ctx *Context, cdc *wire.Codec, appInit AppInit) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gen-tx",
		Short: "Create genesis transaction file (under [--home]/gentx-[nodeID].json)",
		Args:  cobra.NoArgs,
		RunE: func(_ *cobra.Command, args []string) error {

			config := ctx.Config
			nodeKey, err := p2p.LoadOrGenNodeKey(config.NodeKeyFile())
			if err != nil {
				return err
			}
			nodeID := string(nodeKey.ID())
			pubKey := ReadOrCreatePrivValidator(config)

			appGenTx, toPrint, validator, err := appInit.GenAppTx(cdc, pubKey)
			if err != nil {
				return err
			}

			ip := viper.GetString(flagIP)
			if len(ip) == 0 {
				ip, err = externalIP()
				if err != nil {
					return err
				}
			}

			tx := GenesisTx{
				NodeID:    nodeID,
				IP:        ip,
				Validator: validator,
				AppGenTx:  appGenTx,
			}
			bz, err := wire.MarshalJSONIndent(cdc, tx)
			if err != nil {
				return err
			}
			name := fmt.Sprintf("gentx-%v.json", nodeID)
			file := filepath.Join(viper.GetString("home"), name)
			if err != nil {
				return err
				err = cmn.WriteFile(file, bz, 0644)
			}

			out, err := wire.MarshalJSONIndent(cdc, toPrint)
			if err != nil {
				return err
			}
			fmt.Println(string(out))
			return nil
		},
	}
	cmd.Flags().String(flagIP, "", "external facing IP to use if left blank IP will be retrieved from this machine")
	cmd.Flags().AddFlagSet(appInit.FlagsAppTx)
	return cmd
}

// get cmd to initialize all files for tendermint and application
func InitCmd(ctx *Context, cdc *wire.Codec, appInit AppInit) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize genesis config, priv-validator file, and p2p-node file",
		Args:  cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {

			config := ctx.Config
			nodeKey, err := p2p.LoadOrGenNodeKey(config.NodeKeyFile())
			if err != nil {
				return err
			}
			nodeID := string(nodeKey.ID())
			pubKey := ReadOrCreatePrivValidator(config)

			chainID := viper.GetString(flagChainID)
			if chainID == "" {
				chainID = cmn.Fmt("test-chain-%v", cmn.RandStr(6))
			}

			genFile := config.GenesisFile()
			if !viper.GetBool(flagOverwrite) && cmn.FileExists(genFile) {
				return fmt.Errorf("genesis.json file already exists: %v", genFile)
			}

			// process genesis transactions, or otherwise create one for defaults
			var appGenTxs, cliPrints []json.RawMessage
			var validators []tmtypes.GenesisValidator
			var persistentPeers string
			genTxsDir := viper.GetString(flagGenTxs)
			if genTxsDir != "" {
				validators, persistentPeers, appGenTxs, cliPrints = processGenTxs(genTxsDir, cdc, appInit)
				config.P2P.PersistentPeers = persistentPeers
				configFilePath := filepath.Join(viper.GetString("home"), "config", "config.toml") //TODO this is annoying should be easier to get
				cfg.WriteConfigFile(configFilePath, config)
			} else {
				appTx, cliPrint, validator, err := appInit.GenAppTx(cdc, pubKey)
				if err != nil {
					return err
				}
				validators = []tmtypes.GenesisValidator{validator}
				appGenTxs = []json.RawMessage{appTx}
				cliPrints = []json.RawMessage{cliPrint}
			}

			appState, err := appInit.GenAppParams(cdc, appGenTxs)
			if err != nil {
				return err
			}

			err = WriteGenesisFile(cdc, genFile, chainID, validators, appState)
			if err != nil {
				return err
			}

			// print out some key information
			toPrint := struct {
				ChainID    string            `json:"chain_id"`
				NodeID     string            `json:"node_id"`
				AppMessage []json.RawMessage `json:"app_messages"`
			}{
				chainID,
				nodeID,
				cliPrints,
			}
			out, err := wire.MarshalJSONIndent(cdc, toPrint)
			if err != nil {
				return err
			}
			fmt.Println(string(out))

			return nil
		},
	}
	cmd.Flags().BoolP(flagOverwrite, "o", false, "overwrite the genesis.json file")
	cmd.Flags().String(flagChainID, "", "designated chain-id for the genesis")
	cmd.Flags().AddFlagSet(appInit.FlagsAppParams)
	return cmd
}

// append a genesis-piece
func processGenTxs(genTxsDir string, cdc *wire.Codec, appInit AppInit) (
	validators []tmtypes.GenesisValidator, appGenTxs, cliPrints []json.RawMessage, persistentPeers string, err error) {

	fos, err := ioutil.ReadDir(genTxsDir)
	for _, fo := range fos {
		filename := fo.Name()
		if !fo.IsDir() && (path.Ext(filename) != ".json") {
			return nil
		}

		// get the genTx
		bz, err := ioutil.ReadFile(filename)
		if err != nil {
			return err
		}
		var genTx GenesisTx
		err = cdc.UnmarshalJSON(bz, &genTx)
		if err != nil {
			return err
		}

		// combine some stuff
		validators = append(validators, genTx.Validator)
		appGenTxs = append(appGenTxs, genTx.AppGenTxs)
		cliPrints = append(cliPrints, genTx.CliPrints)

		// Add a persistent peer
		comma := ","
		if len(persistentPeers) == 0 {
			comma = ""
		}
		persistentPeers += fmt.Sprintf("%s%s@%s:46656", comma, piece.NodeID, piece.IP)
	}

	return nil
}

//________________________________________________________________________________________

// read of create the private key file for this config
func ReadOrCreatePrivValidator(tmConfig *cfg.Config) crypto.PubKey {
	// private validator
	privValFile := tmConfig.PrivValidatorFile()
	var privValidator *pvm.FilePV
	if cmn.FileExists(privValFile) {
		privValidator = pvm.LoadFilePV(privValFile)
	} else {
		privValidator = pvm.GenFilePV(privValFile)
		privValidator.Save()
	}
	return privValidator.GetPubKey()
}

// create the genesis file
func WriteGenesisFile(cdc *wire.Codec, genesisFile, chainID string, validators []tmtypes.GenesisValidator, appState json.RawMessage) error {
	genDoc := tmtypes.GenesisDoc{
		ChainID:    chainID,
		Validators: validators,
	}
	if err := genDoc.ValidateAndComplete(); err != nil {
		return err
	}
	if err := genDoc.SaveAs(genesisFile); err != nil {
		return err
	}
	return addAppStateToGenesis(cdc, genesisFile, appState)
}

// Add one line to the genesis file
func addAppStateToGenesis(cdc *wire.Codec, genesisConfigPath string, appState json.RawMessage) error {
	bz, err := ioutil.ReadFile(genesisConfigPath)
	if err != nil {
		return err
	}
	out, err := AppendJSON(cdc, bz, "app_state", appState)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(genesisConfigPath, out, 0600)
}

//_____________________________________________________________________

// Core functionality passed from the application to the server init command
type AppInit struct {

	// flags required for application init functions
	FlagsAppParams *pflag.FlagSet
	FlagsAppTx     *pflag.FlagSet

	// GenAppParams creates the core parameters initialization. It takes in a
	// pubkey meant to represent the pubkey of the validator of this machine.
	GenAppParams func(cdc *wire.Codec, appGenTxs []json.RawMessage) (appState json.RawMessage, err error)

	// create the application genesis tx
	GenAppTx func(cdc *wire.Codec, pk crypto.PubKey) (
		appTx, cliPrint json.RawMessage, validator tmtypes.GenesisValidator, err error)
}

// simple default application init
var DefaultAppInit = AppInit{
	GenAppParams: SimpleGenAppParams,
}

// Create one account with a whole bunch of mycoin in it
func SimpleGenAppParams(cdc *wire.Codec, pubKey crypto.PubKey, _ json.RawMessage) (
	validators []tmtypes.GenesisValidator, appState, cliPrint json.RawMessage, err error) {

	var addr sdk.Address

	var secret string
	addr, secret, err = GenerateCoinKey()
	if err != nil {
		return
	}

	mm := map[string]string{"secret": secret}
	bz, err := cdc.MarshalJSON(mm)
	cliPrint = json.RawMessage(bz)

	validators = []tmtypes.GenesisValidator{{
		PubKey: pubKey,
		Power:  10,
	}}

	appState = json.RawMessage(fmt.Sprintf(`{
  "accounts": [{
    "address": "%s",
    "coins": [
      {
        "denom": "mycoin",
        "amount": 9007199254740992
      }
    ]
  }]
}`, addr.String()))
	return
}

// GenerateCoinKey returns the address of a public key, along with the secret
// phrase to recover the private key.
func GenerateCoinKey() (sdk.Address, string, error) {

	// construct an in-memory key store
	codec, err := words.LoadCodec("english")
	if err != nil {
		return nil, "", err
	}
	keybase := keys.New(
		dbm.NewMemDB(),
		codec,
	)

	// generate a private key, with recovery phrase
	info, secret, err := keybase.Create("name", "pass", keys.AlgoEd25519)
	if err != nil {
		return nil, "", err
	}
	addr := info.PubKey.Address()
	return addr, secret, nil
}
