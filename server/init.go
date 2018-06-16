package server

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"sort"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	crypto "github.com/tendermint/go-crypto"
	"github.com/tendermint/go-crypto/keys"
	"github.com/tendermint/go-crypto/keys/words"
	cfg "github.com/tendermint/tendermint/config"
	"github.com/tendermint/tendermint/p2p"
	tmtypes "github.com/tendermint/tendermint/types"
	pvm "github.com/tendermint/tendermint/privval"
	tmcli "github.com/tendermint/tmlibs/cli"
	cmn "github.com/tendermint/tmlibs/common"
	dbm "github.com/tendermint/tmlibs/db"

	clkeys "github.com/cosmos/cosmos-sdk/client/keys"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
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
	flagChainID   = "chain-id"
)

// get cmd to initialize all files for tendermint and application
func GenTxCmd(ctx *Context, cdc *wire.Codec, appInit AppInit) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gen-tx",
		Short: "Create genesis transaction file (under [--home]/config/gentx/gentx-[nodeID].json)",
		Args:  cobra.NoArgs,
		RunE: func(_ *cobra.Command, args []string) error {

			config := ctx.Config
			nodeKey, err := p2p.LoadOrGenNodeKey(config.NodeKeyFile())
			if err != nil {
				return err
			}
			nodeID := string(nodeKey.ID())
			pubKey := readOrCreatePrivValidator(config)

			appGenTx, cliPrint, validator, err := appInit.AppGenTx(cdc, pubKey)
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
			genTxFile := json.RawMessage(bz)
			name := fmt.Sprintf("gentx-%v.json", nodeID)
			writePath := filepath.Join(viper.GetString(tmcli.HomeFlag), "config", "gentx")
			file := filepath.Join(writePath, name)
			err = cmn.EnsureDir(writePath, 0700)
			if err != nil {
				return err
			}
			err = cmn.WriteFile(file, bz, 0644)
			if err != nil {
				return err
			}

			// print out some key information
			toPrint := struct {
				AppMessage json.RawMessage `json:"app_message"`
				GenTxFile  json.RawMessage `json:"gen_tx_file"`
			}{
				cliPrint,
				genTxFile,
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
	cmd.Flags().AddFlagSet(appInit.FlagsAppGenTx)
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
			pubKey := readOrCreatePrivValidator(config)

			chainID := viper.GetString(flagChainID)
			if chainID == "" {
				chainID = cmn.Fmt("test-chain-%v", cmn.RandStr(6))
			}

			genFile := config.GenesisFile()
			if !viper.GetBool(flagOverwrite) && cmn.FileExists(genFile) {
				return fmt.Errorf("genesis.json file already exists: %v", genFile)
			}

			// process genesis transactions, or otherwise create one for defaults
			var appMessage json.RawMessage
			var appGenTxs []json.RawMessage
			var validators []tmtypes.GenesisValidator
			var persistentPeers string

			if viper.GetBool(flagGenTxs) {
				genTxsDir := filepath.Join(viper.GetString(tmcli.HomeFlag), "config", "gentx")
				validators, appGenTxs, persistentPeers, err = processGenTxs(genTxsDir, cdc, appInit)
				if err != nil {
					return err
				}
				config.P2P.PersistentPeers = persistentPeers
				configFilePath := filepath.Join(viper.GetString(tmcli.HomeFlag), "config", "config.toml")
				cfg.WriteConfigFile(configFilePath, config)
			} else {
				appGenTx, am, validator, err := appInit.AppGenTx(cdc, pubKey)
				appMessage = am
				if err != nil {
					return err
				}
				validators = []tmtypes.GenesisValidator{validator}
				appGenTxs = []json.RawMessage{appGenTx}
			}

			appState, err := appInit.AppGenState(cdc, appGenTxs)
			if err != nil {
				return err
			}

			err = writeGenesisFile(cdc, genFile, chainID, validators, appState)
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
			out, err := wire.MarshalJSONIndent(cdc, toPrint)
			if err != nil {
				return err
			}
			fmt.Println(string(out))

			return nil
		},
	}
	cmd.Flags().BoolP(flagOverwrite, "o", false, "overwrite the genesis.json file")
	cmd.Flags().String(flagChainID, "", "genesis file chain-id, if left blank will be randomly created")
	cmd.Flags().Bool(flagGenTxs, false, "apply genesis transactions from [--home]/config/gentx/")
	cmd.Flags().AddFlagSet(appInit.FlagsAppGenState)
	cmd.Flags().AddFlagSet(appInit.FlagsAppGenTx) // need to add this flagset for when no GenTx's provided
	cmd.AddCommand(GenTxCmd(ctx, cdc, appInit))
	return cmd
}

// append a genesis-piece
func processGenTxs(genTxsDir string, cdc *wire.Codec, appInit AppInit) (
	validators []tmtypes.GenesisValidator, appGenTxs []json.RawMessage, persistentPeers string, err error) {

	var fos []os.FileInfo
	fos, err = ioutil.ReadDir(genTxsDir)
	if err != nil {
		return
	}

	genTxs := make(map[string]GenesisTx)
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
		var genTx GenesisTx
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

//________________________________________________________________________________________

// read of create the private key file for this config
func readOrCreatePrivValidator(tmConfig *cfg.Config) crypto.PubKey {
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
func writeGenesisFile(cdc *wire.Codec, genesisFile, chainID string, validators []tmtypes.GenesisValidator, appState json.RawMessage) error {
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
	FlagsAppGenState *pflag.FlagSet
	FlagsAppGenTx    *pflag.FlagSet

	// create the application genesis tx
	AppGenTx func(cdc *wire.Codec, pk crypto.PubKey) (
		appGenTx, cliPrint json.RawMessage, validator tmtypes.GenesisValidator, err error)

	// AppGenState creates the core parameters initialization. It takes in a
	// pubkey meant to represent the pubkey of the validator of this machine.
	AppGenState func(cdc *wire.Codec, appGenTxs []json.RawMessage) (appState json.RawMessage, err error)
}

//_____________________________________________________________________

// simple default application init
var DefaultAppInit = AppInit{
	AppGenTx:    SimpleAppGenTx,
	AppGenState: SimpleAppGenState,
}

// simple genesis tx
type SimpleGenTx struct {
	Addr sdk.Address `json:"addr"`
}

// Generate a genesis transaction
func SimpleAppGenTx(cdc *wire.Codec, pk crypto.PubKey) (
	appGenTx, cliPrint json.RawMessage, validator tmtypes.GenesisValidator, err error) {

	var addr sdk.Address
	var secret string
	addr, secret, err = GenerateCoinKey()
	if err != nil {
		return
	}

	var bz []byte
	simpleGenTx := SimpleGenTx{addr}
	bz, err = cdc.MarshalJSON(simpleGenTx)
	if err != nil {
		return
	}
	appGenTx = json.RawMessage(bz)

	mm := map[string]string{"secret": secret}
	bz, err = cdc.MarshalJSON(mm)
	if err != nil {
		return
	}
	cliPrint = json.RawMessage(bz)

	validator = tmtypes.GenesisValidator{
		PubKey: pk,
		Power:  10,
	}
	return
}

// create the genesis app state
func SimpleAppGenState(cdc *wire.Codec, appGenTxs []json.RawMessage) (appState json.RawMessage, err error) {

	if len(appGenTxs) != 1 {
		err = errors.New("must provide a single genesis transaction")
		return
	}

	var genTx SimpleGenTx
	err = cdc.UnmarshalJSON(appGenTxs[0], &genTx)
	if err != nil {
		return
	}

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
}`, genTx.Addr.String()))
	return
}

//___________________________________________________________________________________________

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

// GenerateSaveCoinKey returns the address of a public key, along with the secret
// phrase to recover the private key.
func GenerateSaveCoinKey(clientRoot, keyName, keyPass string, overwrite bool) (sdk.Address, string, error) {

	// get the keystore from the client
	keybase, err := clkeys.GetKeyBaseFromDir(clientRoot)
	if err != nil {
		return nil, "", err
	}

	// ensure no overwrite
	if !overwrite {
		_, err := keybase.Get(keyName)
		if err == nil {
			return nil, "", errors.New("key already exists, overwrite is disabled")
		}
	}

	// generate a private key, with recovery phrase
	info, secret, err := keybase.Create(keyName, keyPass, keys.AlgoEd25519)
	if err != nil {
		return nil, "", err
	}
	addr := info.PubKey.Address()
	return addr, secret, nil
}
