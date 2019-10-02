package keys

import (
	"bufio"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/input"
	"github.com/cosmos/cosmos-sdk/crypto/keys"
	"github.com/cosmos/cosmos-sdk/types"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// migratePassphrase is used as a no-op migration key passphrase as a passphrase
// is not needed for importing into the Keyring keystore.
const migratePassphrase = "NOOP_PASSPHRASE"

func migrateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Migrate key information from the lagacy key database to the OS secret store, or encrypted file store as a fall-back and save it",
		Long: `Migrate keys from the legacy on-disk secret store to the OS keyring.
The command asks for every passphrase. If the passphrase is incorrect, it skips the respective key.
`,
		Args: cobra.ExactArgs(0),
		RunE: runMigrateCmd,
	}

	cmd.Flags().Bool(flags.FlagDryRun, false, "Do everything which is supposed to be done, but don't write any changes to the keyring.")
	return cmd
}

func runMigrateCmd(cmd *cobra.Command, args []string) error {
	// instantiate legacy keybase
	rootDir := viper.GetString(flags.FlagHome)
	legacykb, err := NewKeyBaseFromDir(rootDir)
	if err != nil {
		return err
	}

	// fetch list of keys from legacy keybase
	oldKeys, err := legacykb.List()
	if err != nil {
		return err
	}

	// instantiate keyring
	var keyring keys.Keybase
	buf := bufio.NewReader(cmd.InOrStdin())
	if viper.GetBool(flags.FlagDryRun) {
		keyring = keys.NewTestKeyring(types.GetConfig().GetKeyringServiceName(), rootDir)
	} else {
		keyring = keys.NewKeyring(types.GetConfig().GetKeyringServiceName(), rootDir, buf)
	}

	for _, key := range oldKeys {
		legKeyInfo, err := legacykb.Export(key.GetName())
		if err != nil {
			return err
		}

		keyName := key.GetName()
		keyType := key.GetType()
		cmd.PrintErrf("Migrating %s (%s) ...\n", key.GetName(), keyType)
		if keyType != keys.TypeLocal {
			if err := keyring.Import(keyName, legKeyInfo); err != nil {
				return err
			}
			continue
		}

		password, err := input.GetPassword("Enter passphrase to decrypt your key:", buf)
		if err != nil {
			return err
		}

		// NOTE: A passphrase is not actually needed here as when the key information
		// is imported into the Keyring keystore it only needs the password (see: writeLocalKey).
		armoredPriv, err := legacykb.ExportPrivKey(keyName, password, migratePassphrase)
		if err != nil {
			return err
		}

		if err := keyring.ImportPrivKey(keyName, armoredPriv, migratePassphrase); err != nil {
			return err
		}
	}

	return err
}
