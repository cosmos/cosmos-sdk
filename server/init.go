package server

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
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

// get cmd to initialize all files for tendermint and application
func InitCmd(ctx *Context, cdc *wire.Codec, appInit AppInit) *cobra.Command {
	flagOverwrite, flagPieceFile, flagIP := "overwrite", "piece-file", "ip"
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize genesis config, priv-validator file, and p2p-node file",
		Args:  cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {

			config := ctx.Config
			pubkey := ReadOrCreatePrivValidator(config)

			chainID, validators, appState, cliPrint, err := appInit.GenAppParams(cdc, pubkey)
			if err != nil {
				return err
			}

			genFile := config.GenesisFile()
			if !viper.GetBool(flagOverwrite) && cmn.FileExists(genFile) {
				return fmt.Errorf("genesis config file already exists: %v", genFile)
			}

			err = WriteGenesisFile(cdc, genFile, chainID, validators, appState)
			if err != nil {
				return err
			}

			nodeKey, err := p2p.LoadOrGenNodeKey(config.NodeKeyFile())
			if err != nil {
				return err
			}
			nodeID := string(nodeKey.ID())

			// print out some key information

			toPrint := struct {
				ChainID    string          `json:"chain_id"`
				NodeID     string          `json:"node_id"`
				AppMessage json.RawMessage `json:"app_message"`
			}{
				chainID,
				nodeID,
				cliPrint,
			}
			out, err := wire.MarshalJSONIndent(cdc, toPrint)
			if err != nil {
				return err
			}
			fmt.Println(string(out))

			// write the piece file is path specified
			if viper.GetBool(flagPieceFile) {
				//create the piece
				ip := viper.GetString(flagIP)
				if len(ip) == 0 {
					ip, err = externalIP()
					if err != nil {
						return err
					}
				}
				piece := GenesisPiece{
					ChainID:    chainID,
					NodeID:     nodeID,
					IP:         ip,
					AppState:   appState,
					Validators: validators,
				}
				bz, err := wire.MarshalJSONIndent(cdc, piece)
				if err != nil {
					return err
				}
				name := fmt.Sprintf("piece%v.json", nodeID)
				file := filepath.Join(viper.GetString("home"), name)
				return cmn.WriteFile(file, bz, 0644)
			}

			return nil
		},
	}
	if appInit.AppendAppState != nil {
		cmd.AddCommand(FromPiecesCmd(ctx, cdc, appInit))
		cmd.Flags().BoolP(flagPieceFile, "a", false, "create an append file (under [--home]/[nodeID]piece.json) for others to import")
	}
	cmd.Flags().BoolP(flagOverwrite, "o", false, "overwrite the config file")
	cmd.Flags().String(flagIP, "", "external facing IP to use if left blank IP will be retrieved from this machine")
	cmd.Flags().AddFlagSet(appInit.Flags)
	return cmd
}

// genesis piece structure for creating combined genesis
type GenesisPiece struct {
	ChainID    string                     `json:"chain_id"`
	NodeID     string                     `json:"node_id"`
	IP         string                     `json:"ip"`
	AppState   json.RawMessage            `json:"app_state"`
	Validators []tmtypes.GenesisValidator `json:"validators"`
}

// get cmd to initialize all files for tendermint and application
func FromPiecesCmd(ctx *Context, cdc *wire.Codec, appInit AppInit) *cobra.Command {
	return &cobra.Command{
		Use:   "from-pieces [directory]",
		Short: "Create genesis from directory of genesis pieces",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			pieceDir := args[0]

			// ensure that the privVal and nodeKey file already exist
			config := ctx.Config
			privValFile := config.PrivValidatorFile()
			nodeKeyFile := config.NodeKeyFile()
			if !cmn.FileExists(privValFile) {
				return fmt.Errorf("privVal file must already exist, please initialize with init cmd: %v", privValFile)
			}
			if !cmn.FileExists(nodeKeyFile) {
				return fmt.Errorf("nodeKey file must already exist, please initialize with init cmd: %v", nodeKeyFile)
			}

			// remove genFile for creation
			genFile := config.GenesisFile()
			os.Remove(genFile)

			// deterministically walk the directory for genesis-piece files to import
			err := filepath.Walk(pieceDir, appendPiece(ctx, cdc, appInit, nodeKeyFile, genFile))
			if err != nil {
				return err
			}

			return nil
		},
	}
}

// append a genesis-piece
func appendPiece(ctx *Context, cdc *wire.Codec, appInit AppInit, nodeKeyFile, genFile string) filepath.WalkFunc {
	return func(pieceFile string, _ os.FileInfo, err error) error {
		fmt.Printf("debug pieceFile: %v\n", pieceFile)
		if err != nil {
			return err
		}
		if path.Ext(pieceFile) != ".json" {
			return nil
		}

		// get the piece file bytes
		bz, err := ioutil.ReadFile(pieceFile)
		if err != nil {
			return err
		}

		// get the piece
		var piece GenesisPiece
		err = cdc.UnmarshalJSON(bz, &piece)
		if err != nil {
			return err
		}

		// if the first file, create the genesis from scratch with piece inputs
		if !cmn.FileExists(genFile) {
			return WriteGenesisFile(cdc, genFile, piece.ChainID, piece.Validators, piece.AppState)
		}

		// read in the genFile
		bz, err = ioutil.ReadFile(genFile)
		if err != nil {
			return err
		}
		var genMap map[string]json.RawMessage
		err = cdc.UnmarshalJSON(bz, &genMap)
		if err != nil {
			return err
		}
		appState := genMap["app_state"]

		// verify chain-ids are the same
		var genChainID string
		err = cdc.UnmarshalJSON(genMap["chain_id"], &genChainID)
		if err != nil {
			return err
		}
		if piece.ChainID != genChainID {
			return fmt.Errorf("piece chain id's are mismatched, %s != %s", piece.ChainID, genChainID)
		}

		// combine the validator set
		var validators []tmtypes.GenesisValidator
		err = cdc.UnmarshalJSON(genMap["validators"], &validators)
		if err != nil {
			return err
		}
		validators = append(validators, piece.Validators...)

		// combine the app state
		appState, err = appInit.AppendAppState(cdc, appState, piece.AppState)
		if err != nil {
			return err
		}

		// write the appended genesis file
		err = WriteGenesisFile(cdc, genFile, piece.ChainID, validators, appState)
		if err != nil {
			return err
		}

		// Add a persistent peer if the config (if it's not me)
		myIP, err := externalIP()
		if err != nil {
			return err
		}
		if myIP == piece.IP {
			return nil
		}
		comma := ","
		if len(ctx.Config.P2P.PersistentPeers) == 0 {
			comma = ""
		}
		//newPeer := fmt.Sprintf("%s%s@%s:46656", comma, piece.NodeID, piece.IP)
		ctx.Config.P2P.PersistentPeers += fmt.Sprintf("%s%s@%s:46656", comma, piece.NodeID, piece.IP)
		configFilePath := filepath.Join(viper.GetString("home"), "config", "config.toml") //TODO this is annoying should be easier to get
		cfg.WriteConfigFile(configFilePath, ctx.Config)

		return nil
	}
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

	// flags required for GenAppParams
	Flags *pflag.FlagSet

	// GenAppParams creates the core parameters initialization. It takes in a
	// pubkey meant to represent the pubkey of the validator of this machine.
	GenAppParams func(*wire.Codec, crypto.PubKey) (chainID string, validators []tmtypes.GenesisValidator, appState, cliPrint json.RawMessage, err error)

	// append appState1 with appState2
	AppendAppState func(cdc *wire.Codec, appState1, appState2 json.RawMessage) (appState json.RawMessage, err error)
}

// simple default application init
var DefaultAppInit = AppInit{
	GenAppParams: SimpleGenAppParams,
}

// Create one account with a whole bunch of mycoin in it
func SimpleGenAppParams(cdc *wire.Codec, pubKey crypto.PubKey) (chainID string, validators []tmtypes.GenesisValidator, appState, cliPrint json.RawMessage, err error) {

	var addr sdk.Address

	var secret string
	addr, secret, err = GenerateCoinKey()
	if err != nil {
		return
	}

	mm := map[string]string{"secret": secret}
	bz, err := cdc.MarshalJSON(mm)
	cliPrint = json.RawMessage(bz)

	chainID = cmn.Fmt("test-chain-%v", cmn.RandStr(6))

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
