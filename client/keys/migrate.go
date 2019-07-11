package keys

import (
	"bufio"
	"fmt"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/input"
	"github.com/cosmos/cosmos-sdk/crypto/keys"
	"github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func migrateKeyCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Migrate key information from the lagacy key database to the OS secret store, or encrypted file store as a fall-back and save it",
		Long: `This command migrates key information from the legacy secret store to the OS secret store. The command asks for every passphrase. 
		If passphrase is incorrect, it skips the key. 

		Previous versions of Gaia used a custom secret store. On version xxx, Gaia CLI was updated to use a library Keyring (https://github.com/99designs/keyring) to 
		preferentially store secrets in the secret manager of many Operating Systems. This is intended to provide stronger security guarantees than the 
		custom secret store is provided. 
`,
		Args: cobra.ExactArgs(1),
		RunE: runMigrateCmd,
	}

	return cmd
}

func runMigrateCmd(cmd *cobra.Command, args []string) error {

	var legacykb keys.Keybase
	var keyringkb keys.Keybase
	var err error

	//instantiating variables
	legacykb, err = NewKeyBaseFromHomeFlag()
	if err != nil {
		return err
	}

	rootDir := viper.GetString(flags.FlagHome)

	keyringkb = keys.NewKeybaseKeyring(types.GetConfig().GetKeyringServiceName(), rootDir)

	legacyKeyList, err := legacykb.List()
	for _, key := range legacyKeyList {

		legKeyInfo, err := legacykb.Export(key.GetName())
		if err != nil {
			return err
		}

		keyringkb.Import(key.GetName(), legKeyInfo)

		switch key.GetType() {
		case keys.TypeLocal:
			buf := bufio.NewReader(cmd.InOrStdin())
			fmt.Printf(" Migrating %s \n", key.GetName())
			decryptPassword, err := input.GetPassword("Enter passphrase to decrypt your key:", buf)

			if err != nil {
				return err
			}

			privData, err := legacykb.ExportPrivKey(key.GetName(), decryptPassword, "abc")

			keyringkb.ImportPrivKey(key.GetName(), privData, "abc")

		case keys.TypeOffline, keys.TypeMulti, keys.TypeLedger:
			continue

		}
	}

	//no private key, write data into keyring
	//has private key, prompt for passphrase, decrypt, add keyingo plus private data

	return err
}
