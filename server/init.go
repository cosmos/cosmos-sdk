package server

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/spf13/cobra"

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
func InitCmd(gen GenAppParams, ctx *Context) *cobra.Command {
	cobraCmd := cobra.Command{
		Use:   "init",
		Short: "Initialize genesis files",
		RunE: func(cmd *cobra.Command, args []string) error {

			config := ctx.Config
			pubkey := ReadOrCreatePrivValidator(config)

			chainID, validators, appState, err := gen(pubkey)
			if err != nil {
				return err
			}

			err = CreateGenesisFile(config, chainID, validators, appState)
			if err != nil {
				return err
			}

			nodeKey, err := p2p.LoadOrGenNodeKey(config.NodeKeyFile())
			if err != nil {
				return err
			}

			// print out some key information
			toPrint := struct {
				ChainID string `json:"chain_id"`
				NodeID  string `json:"node_id"`
			}{
				chainID,
				string(nodeKey.ID()),
			}
			out, err := wire.MarshalJSONIndent(cdc, toPrint)
			if err != nil {
				return err
			}
			fmt.Println(string(out))
			return nil
		},
	}
	return &cobraCmd
}

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
func CreateGenesisFile(tmConfig *cfg.Config, chainID string, validators []tmtypes.GenesisValidator, appState json.RawMessage) error {
	genFile := tmConfig.GenesisFile()
	if cmn.FileExists(genFile) {
		return fmt.Errorf("genesis config file already exists: %v", genFile)
	}
	genDoc := tmtypes.GenesisDoc{
		ChainID:    chainID,
		Validators: validators,
	}
	if err := genDoc.ValidateAndComplete(); err != nil {
		return err
	}
	if err := genDoc.SaveAs(genFile); err != nil {
		return err
	}
	return addAppStateToGenesis(genFile, appState)
}

// Add one line to the genesis file
func addAppStateToGenesis(genesisConfigPath string, appState json.RawMessage) error {
	bz, err := ioutil.ReadFile(genesisConfigPath)
	if err != nil {
		return err
	}

	var doc map[string]json.RawMessage
	err = cdc.UnmarshalJSON(bz, &doc)
	if err != nil {
		return err
	}

	doc["app_state"] = appState
	out, err := wire.MarshalJSONIndent(cdc, doc)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(genesisConfigPath, out, 0600)
}

//_____________________________________________________________________

// GenAppParams creates the core parameters initialization. It takes in a
// pubkey meant to represent the pubkey of the validator of this machine.
type GenAppParams func(crypto.PubKey) (chainID string, validators []tmtypes.GenesisValidator, appState json.RawMessage, err error)

// Create one account with a whole bunch of mycoin in it
func SimpleGenAppState(pubKey crypto.PubKey) (chainID string, validators []tmtypes.GenesisValidator, appState json.RawMessage, err error) {

	var addr sdk.Address
	var secret string
	addr, secret, err = GenerateCoinKey()
	if err != nil {
		return
	}
	fmt.Printf("secret recovery key:\n%s\n", secret)

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
