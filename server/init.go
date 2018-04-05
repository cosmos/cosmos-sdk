package server

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cobra"

	"github.com/tendermint/go-crypto/keys"
	"github.com/tendermint/go-crypto/keys/words"
	cfg "github.com/tendermint/tendermint/config"
	"github.com/tendermint/tendermint/p2p"
	tmtypes "github.com/tendermint/tendermint/types"
	cmn "github.com/tendermint/tmlibs/common"
	dbm "github.com/tendermint/tmlibs/db"
)

// testnetInformation contains the info necessary
// to setup a testnet including this account and validator.
type testnetInformation struct {
	Secret string `json:"secret"`

	ChainID   string                   `json:"chain_id"`
	Account   string                   `json:"account"`
	Validator tmtypes.GenesisValidator `json:"validator"`
	NodeID    p2p.ID                   `json:"node_id"`
}

type initCmd struct {
	genAppState GenAppState
	context     *Context
}

// InitCmd will initialize all files for tendermint,
// along with proper app_state.
// The application can pass in a function to generate
// proper state. And may want to use GenerateCoinKey
// to create default account(s).
func InitCmd(gen GenAppState, ctx *Context) *cobra.Command {
	cmd := initCmd{
		genAppState: gen,
		context:     ctx,
	}
	cobraCmd := cobra.Command{
		Use:   "init",
		Short: "Initialize genesis files",
		RunE:  cmd.run,
	}
	return &cobraCmd
}

func (c initCmd) run(cmd *cobra.Command, args []string) error {
	// Store testnet information as we go
	var testnetInfo testnetInformation

	// Run the basic tendermint initialization,
	// set up a default genesis with no app_options
	config := c.context.Config
	err := c.initTendermintFiles(config, &testnetInfo)
	if err != nil {
		return err
	}

	// no app_options, leave like tendermint
	if c.genAppState == nil {
		return nil
	}

	// generate secrete and address
	addr, secret, err := GenerateCoinKey()
	if err != nil {
		return err
	}

	var DEFAULT_DENOM = "mycoin"

	// Now, we want to add the custom app_state
	appState, err := c.genAppState(args, addr, DEFAULT_DENOM)
	if err != nil {
		return err
	}

	testnetInfo.Secret = secret
	testnetInfo.Account = addr.String()

	// And add them to the genesis file
	genFile := config.GenesisFile()
	if err := addGenesisState(genFile, appState); err != nil {
		return err
	}

	nodeKey, err := p2p.LoadOrGenNodeKey(config.NodeKeyFile())
	if err != nil {
		return err
	}
	testnetInfo.NodeID = nodeKey.ID()
	out, err := json.MarshalIndent(testnetInfo, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(out))
	return nil
}

// This was copied from tendermint/cmd/tendermint/commands/init.go
// so we could pass in the config and the logger.
func (c initCmd) initTendermintFiles(config *cfg.Config, info *testnetInformation) error {
	// private validator
	privValFile := config.PrivValidatorFile()
	var privValidator *tmtypes.PrivValidatorFS
	if cmn.FileExists(privValFile) {
		privValidator = tmtypes.LoadPrivValidatorFS(privValFile)
		c.context.Logger.Info("Found private validator", "path", privValFile)
	} else {
		privValidator = tmtypes.GenPrivValidatorFS(privValFile)
		privValidator.Save()
		c.context.Logger.Info("Generated private validator", "path", privValFile)
	}

	// genesis file
	genFile := config.GenesisFile()
	if cmn.FileExists(genFile) {
		c.context.Logger.Info("Found genesis file", "path", genFile)
	} else {
		genDoc := tmtypes.GenesisDoc{
			ChainID: cmn.Fmt("test-chain-%v", cmn.RandStr(6)),
		}
		genDoc.Validators = []tmtypes.GenesisValidator{{
			PubKey: privValidator.GetPubKey(),
			Power:  10,
		}}

		if err := genDoc.SaveAs(genFile); err != nil {
			return err
		}
		c.context.Logger.Info("Generated genesis file", "path", genFile)
	}

	// reload the config file and find our validator info
	loadedDoc, err := tmtypes.GenesisDocFromFile(genFile)
	if err != nil {
		return err
	}
	for _, validator := range loadedDoc.Validators {
		if validator.PubKey == privValidator.GetPubKey() {
			info.Validator = validator
		}
	}
	info.ChainID = loadedDoc.ChainID

	return nil
}

//-------------------------------------------------------------------

// GenAppState takes the command line args, as well
// as an address and coin denomination.
// It returns a default app_state to be included in
// in the genesis file.
// This is application-specific
type GenAppState func(args []string, addr sdk.Address, coinDenom string) (json.RawMessage, error)

// DefaultGenAppState expects two args: an account address
// and a coin denomination, and gives lots of coins to that address.
func DefaultGenAppState(args []string, addr sdk.Address, coinDenom string) (json.RawMessage, error) {
	opts := fmt.Sprintf(`{
      "accounts": [{
        "address": "%s",
        "coins": [
          {
            "denom": "%s",
            "amount": 9007199254740992
          }
        ]
      }]
    }`, addr.String(), coinDenom)
	return json.RawMessage(opts), nil
}

//-------------------------------------------------------------------

// GenesisDoc involves some tendermint-specific structures we don't
// want to parse, so we just grab it into a raw object format,
// so we can add one line.
type GenesisDoc map[string]json.RawMessage

func addGenesisState(filename string, appState json.RawMessage) error {
	bz, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	var doc GenesisDoc
	err = json.Unmarshal(bz, &doc)
	if err != nil {
		return err
	}

	doc["app_state"] = appState
	out, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return err
	}

	return ioutil.WriteFile(filename, out, 0600)
}

//-------------------------------------------------------------------

// GenerateCoinKey returns the address of a public key,
// along with the secret phrase to recover the private key.
// You can give coins to this address and return the recovery
// phrase to the user to access them.
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
