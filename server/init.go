package server

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/spf13/cobra"

	crypto "github.com/tendermint/go-crypto"
	"github.com/tendermint/go-crypto/keys"
	"github.com/tendermint/go-crypto/keys/words"
	tcmd "github.com/tendermint/tendermint/cmd/tendermint/commands"
	tmtypes "github.com/tendermint/tendermint/types"
)

// InitCmd will initialize all files for tendermint,
// along with proper app_options.
// The application can pass in a function to generate
// proper options. And may want to use GenerateCoinKey
// to create default account(s).
func InitCmd(gen GenOptions) *cobra.Command {
	cmd := initCmd{
		gen: gen,
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
	gen GenOptions
}

func (c initCmd) run(cmd *cobra.Command, args []string) error {
	// Run the basic tendermint initialization,
	// set up a default genesis with no app_options
	cfg, err := tcmd.ParseConfig()
	if err != nil {
		return err
	}
	tcmd.InitFilesCmd.Run(cmd, args)

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
	genFile := cfg.GenesisFile()
	return addGenesisOptions(genFile, options)
}

func addGenesisOptions(filename string, options json.RawMessage) error {
	bz, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	var doc tmtypes.GenesisDoc
	err = json.Unmarshal(bz, &doc)
	if err != nil {
		return err
	}

	doc.AppOptions = options
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
  "app_hash": "",
  "chain_id": "%s",
  "genesis_time": "0001-01-01T00:00:00.000Z",
  "validators": [
    {
      "power": 10,
      "name": "",
      "pub_key": {
        "type": "ed25519",
        "data": "%s"
      }
    }
  ],
  "app_options": {
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
  }
}`, chainID, pubkey, addr, denom, addr, options)
}
