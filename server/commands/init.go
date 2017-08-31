package commands

import (
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	tcmd "github.com/tendermint/tendermint/cmd/tendermint/commands"
	"github.com/tendermint/tendermint/config"
	cmn "github.com/tendermint/tmlibs/common"
)

// InitCmd - node initialization command
var InitCmd = &cobra.Command{
	Use:   "init [address]",
	Short: "Initialize genesis files for a blockchain",
	RunE:  initCmd,
}

//nolint - flags
var (
	FlagChainID = "chain-id" //TODO group with other flags or remove? is this already a flag here?
	FlagOption  = "option"
)

func init() {
	InitCmd.Flags().String(FlagChainID, "test_chain_id", "Chain ID")
	InitCmd.Flags().StringSlice(FlagOption, []string{}, "Genesis option in the format <app>/<option>/<value>")
}

// returns 1 iff it set a file, otherwise 0 (so we can add them)
func setupFile(path, data string, perm os.FileMode) (int, error) {
	_, err := os.Stat(path)
	if !os.IsNotExist(err) { //note, os.IsExist(err) != !os.IsNotExist(err)
		return 0, nil
	}
	err = ioutil.WriteFile(path, []byte(data), perm)
	if err != nil {
		return 0, err
	}
	return 1, nil
}

func initCmd(cmd *cobra.Command, args []string) error {
	// this will ensure that config.toml is there if not yet created, and create dir
	cfg, err := tcmd.ParseConfig()
	if err != nil {
		return err
	}

	if len(args) != 1 {
		return fmt.Errorf("`init` takes one argument, a basecoin account address. Generate one using `basecli keys new mykey`")
	}
	userAddr := args[0]
	// verify this account is correct
	data, err := hex.DecodeString(cmn.StripHex(userAddr))
	if err != nil {
		return errors.Wrap(err, "Invalid address")
	}
	if len(data) != 20 {
		return errors.New("Address must be 20-bytes in hex")
	}

	var options []string
	var optionsStr string
	sep := ",\n      "
	optionsRaw := viper.GetStringSlice(FlagOption)
	if len(optionsRaw) > 0 {
		optionsStr = sep
		for i := 0; i < len(optionsRaw); i++ {
			s := strings.SplitN(optionsRaw[i], "/", 3)
			if len(s) != 3 {
				return errors.New("Genesis option must be in the format <app>/<option>/<value>")
			}
			option := `"` + s[0] + `/` + s[1] + `", "` + s[2] + `"`
			options = append(options, option)
		}
	}
	optionsStr += strings.Join(options[:], sep)

	genesis := GetGenesisJSON(viper.GetString(FlagChainID), userAddr, optionsStr)
	return CreateGenesisValidatorFiles(cfg, genesis, cmd.Root().Name())
}

// CreateGenesisValidatorFiles creates a genesis file with these
// contents and a private validator file
func CreateGenesisValidatorFiles(cfg *config.Config, genesis, appName string) error {
	genesisFile := cfg.GenesisFile()
	privValFile := cfg.PrivValidatorFile()

	mod1, err := setupFile(genesisFile, genesis, 0644)
	if err != nil {
		return err
	}
	mod2, err := setupFile(privValFile, PrivValJSON, 0400)
	if err != nil {
		return err
	}

	if (mod1 + mod2) > 0 {
		msg := fmt.Sprintf("Initialized %s", appName)
		logger.Info(msg, "genesis", genesisFile, "priv_validator", privValFile)
	} else {
		logger.Info("Already initialized", "priv_validator", privValFile)
	}

	return nil
}

// PrivValJSON - validator private key file contents in json
var PrivValJSON = `{
  "address": "7A956FADD20D3A5B2375042B2959F8AB172A058F",
  "last_height": 0,
  "last_round": 0,
  "last_signature": null,
  "last_signbytes": "",
  "last_step": 0,
  "priv_key": {
    "type": "ed25519",
    "data": "D07ABE82A8B15559A983B2DB5D4842B2B6E4D6AF58B080005662F424F17D68C17B90EA87E7DC0C7145C8C48C08992BE271C7234134343E8A8E8008E617DE7B30"
  },
  "pub_key": {
    "type": "ed25519",
    "data": "7B90EA87E7DC0C7145C8C48C08992BE271C7234134343E8A8E8008E617DE7B30"
  }
}`

// GetGenesisJSON returns a new tendermint genesis with Basecoin app_options
// that grant a large amount of "mycoin" to a single address
// TODO: A better UX for generating genesis files
func GetGenesisJSON(chainID, addr string, options string) string {
	return fmt.Sprintf(`{
  "app_hash": "",
  "chain_id": "%s",
  "genesis_time": "0001-01-01T00:00:00.000Z",
  "validators": [
    {
      "amount": 10,
      "name": "",
      "pub_key": {
        "type": "ed25519",
        "data": "7B90EA87E7DC0C7145C8C48C08992BE271C7234134343E8A8E8008E617DE7B30"
      }
    }
  ],
  "app_options": {
    "accounts": [{
      "address": "%s",
      "coins": [
        {
          "denom": "mycoin",
          "amount": 9007199254740992
        }
      ]
    }],
    "plugin_options": [
      "coin/issuer", {"app": "sigs", "addr": "%s"}%s
    ]
  }
}`, chainID, addr, addr, options)
}
