package keys

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/input"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// migratePassphrase is used as a no-op migration key passphrase as a passphrase
// is not needed for importing into the Keyring keystore.
const migratePassphrase = "NOOP_PASSPHRASE"

// MigrateCommand migrates key information from legacy keybase to OS secret store.
func MigrateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migrate <old_home_dir>",
		Short: "Migrate keys from the legacy (db-based) Keybase",
		Long: `Migrate key information from the legacy (db-based) Keybase to the new keyring-based Keyring.
The legacy Keybase used to persist keys in a LevelDB database stored in a 'keys' sub-directory of
the old client application's home directory, e.g. $HOME/.gaiacli/keys/.
For each key material entry, the command will prompt if the key should be skipped or not. If the key
is not to be skipped, the passphrase must be entered. The key will only be migrated if the passphrase
is correct. Otherwise, the command will exit and migration must be repeated.

It is recommended to run in 'dry-run' mode first to verify all key migration material.
`,
		Args: cobra.ExactArgs(1),
		RunE: runMigrateCmd,
	}

	cmd.Flags().Bool(flags.FlagDryRun, false, "Run migration without actually persisting any changes to the new Keybase")
	return cmd
}

func runMigrateCmd(cmd *cobra.Command, args []string) error {
	rootDir, _ := cmd.Flags().GetString(flags.FlagHome)

	// instantiate legacy keybase
	var legacyKb keyring.LegacyKeybase
	legacyKb, err := NewLegacyKeyBaseFromDir(args[0])
	if err != nil {
		return err
	}

	defer func() { _ = legacyKb.Close() }()

	// fetch list of keys from legacy keybase
	oldKeys, err := legacyKb.List()
	if err != nil {
		return err
	}

	buf := bufio.NewReader(cmd.InOrStdin())
	keyringServiceName := sdk.KeyringServiceName()

	var (
		tmpDir   string
		migrator keyring.Importer
	)

	if dryRun, _ := cmd.Flags().GetBool(flags.FlagDryRun); dryRun {
		tmpDir, err = ioutil.TempDir("", "migrator-migrate-dryrun")
		if err != nil {
			return errors.Wrap(err, "failed to create temporary directory for dryrun migration")
		}

		defer func() { _ = os.RemoveAll(tmpDir) }()

		migrator, err = keyring.New(keyringServiceName, keyring.BackendTest, tmpDir, buf)
	} else {
		backend, _ := cmd.Flags().GetString(flags.FlagKeyringBackend)
		migrator, err = keyring.New(keyringServiceName, backend, rootDir, buf)
	}

	if err != nil {
		return errors.Wrap(err, fmt.Sprintf(
			"failed to initialize keybase for service %s at directory %s",
			keyringServiceName, rootDir,
		))
	}

	if len(oldKeys) == 0 {
		cmd.Print("Migration Aborted: no keys to migrate")
		return nil
	}

	for _, oldInfo := range oldKeys {
		keyName := oldInfo.GetName()
		keyType := oldInfo.GetType()

		cmd.PrintErrf("Migrating key: '%s (%s)' ...\n", keyName, keyType)

		// allow user to skip migrating specific keys
		ok, err := input.GetConfirmation("Skip key migration?", buf, cmd.ErrOrStderr())
		if err != nil {
			return err
		}
		if ok {
			continue
		}

		// TypeLocal needs an additional step to ask password.
		// The other keyring types are handled by ImportInfo.
		if keyType != keyring.TypeLocal {
			infoImporter, ok := migrator.(keyring.LegacyInfoImporter)
			if !ok {
				return fmt.Errorf("the Keyring implementation does not support import operations of Info types")
			}

			if err = infoImporter.ImportInfo(oldInfo); err != nil {
				return err
			}

			continue
		}

		password, err := input.GetPassword("Enter passphrase to decrypt key:", buf)
		if err != nil {
			return err
		}

		// NOTE: A passphrase is not actually needed here as when the key information
		// is imported into the Keyring-based Keybase it only needs the password
		// (see: writeLocalKey).
		armoredPriv, err := legacyKb.ExportPrivKey(keyName, password, migratePassphrase)
		if err != nil {
			return err
		}

		if err := migrator.ImportPrivKey(keyName, armoredPriv, migratePassphrase); err != nil {
			return err
		}

	}
	cmd.PrintErrln("Migration complete.")

	return err
}
