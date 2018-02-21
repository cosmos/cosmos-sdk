package server

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/spf13/cobra"

	crypto "github.com/tendermint/go-crypto"
	"github.com/tendermint/go-crypto/keys"
	"github.com/tendermint/go-crypto/keys/words"
	cmn "github.com/tendermint/tmlibs/common"
	dbm "github.com/tendermint/tmlibs/db"
	"github.com/tendermint/tmlibs/log"

	tcmd "github.com/tendermint/tendermint/cmd/tendermint/commands"
	cfg "github.com/tendermint/tendermint/config"
	tmtypes "github.com/tendermint/tendermint/types"
)

// InitCmd will initialize all files for tendermint,
// along with proper app_options.
// The application can pass in a function to generate
// proper options. And may want to use GenerateCoinKey
// to create default account(s).
func InitCmd(gen GenOptions, logger log.Logger) *cobra.Command {
	cmd := initCmd{
		gen:    gen,
		logger: logger,
	}
	return &cobra.Command{
		Use:   "init",
		Short: "Initialize genesis files",
		RunE:  cmd.run,
	}
}

// GenOptions can parse command-line and flag to
// generate default app_options for the genesis file.
// This is application-specific
type GenOptions func(args []string) (json.RawMessage, error)

// GenerateCoinKey returns the address of a public key,
// along with the secret phrase to recover the private key.
// You can give coins to this address and return the recovery
// phrase to the user to access them.
func GenerateCoinKey() (crypto.Address, string, error) {
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

type initCmd struct {
	gen    GenOptions
	logger log.Logger
}

func (c initCmd) run(cmd *cobra.Command, args []string) error {
	// Run the basic tendermint initialization,
	// set up a default genesis with no app_options
	config, err := tcmd.ParseConfig()
	if err != nil {
		return err
	}
	err = c.initTendermintFiles(config)
	if err != nil {
		return err
	}

	// no app_options, leave like tendermint
	if c.gen == nil {
		return nil
	}

	// Now, we want to add the custom app_options
	options, err := c.gen(args)
	if err != nil {
		return err
	}

	// And add them to the genesis file
	genFile := config.GenesisFile()
	return addGenesisOptions(genFile, options)
}

// This was copied from tendermint/cmd/tendermint/commands/init.go
// so we could pass in the config and the logger.
func (c initCmd) initTendermintFiles(config *cfg.Config) error {
	// private validator
	privValFile := config.PrivValidatorFile()
	var privValidator *tmtypes.PrivValidatorFS
	if cmn.FileExists(privValFile) {
		privValidator = tmtypes.LoadPrivValidatorFS(privValFile)
		c.logger.Info("Found private validator", "path", privValFile)
	} else {
		privValidator = tmtypes.GenPrivValidatorFS(privValFile)
		privValidator.Save()
		c.logger.Info("Genetated private validator", "path", privValFile)
	}

	// genesis file
	genFile := config.GenesisFile()
	if cmn.FileExists(genFile) {
		c.logger.Info("Found genesis file", "path", genFile)
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
		c.logger.Info("Genetated genesis file", "path", genFile)
	}
	return nil
}

// GenesisDoc involves some tendermint-specific structures we don't
// want to parse, so we just grab it into a raw object format,
// so we can add one line.
type GenesisDoc map[string]json.RawMessage

func addGenesisOptions(filename string, options json.RawMessage) error {
	bz, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	var doc GenesisDoc
	err = json.Unmarshal(bz, &doc)
	if err != nil {
		return err
	}

	doc["app_state"] = options
	out, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return err
	}

	return ioutil.WriteFile(filename, out, 0600)
}

// GetGenesisJSON returns a new tendermint genesis with Basecoin app_options
// that grant a large amount of "mycoin" to a single address
// TODO: A better UX for generating genesis files
func GetGenesisJSON(pubkey, chainID, denom, addr string, options string) string {
	return fmt.Sprintf(`{
    "accounts": [{
      "address": "%s",
      "coins": [
        {
          "denom": "%s",
          "amount": 9007199254740992
        }
      ]
    }],
    "plugin_options": [
      "coin/issuer", {"app": "sigs", "addr": "%s"}%s
    ]
}`, chainID, pubkey, addr, denom, addr, options)
}
