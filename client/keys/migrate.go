package keys

import (
	"bufio"
	"fmt"
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
)

// MigrateCommand migrates key information from legacy keybase to OS secret store.
func MigrateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Migrate keys from amino to proto serialization format",
		Long: `Migrate keys from Amino to Protocol Buffers records.
For each key material entry, the command will check if the key can be deserialized using proto.
If this is the case, the key is already migrated. Therefore, we skip it and continue with a next one. 
Otherwise, we try to deserialize it using Amino into LegacyInfo. If this attempt is successful, we serialize 
LegacyInfo to Protobuf serialization format and overwrite the keyring entry. If any error occurred, it will be 
outputted in CLI and migration will be continued until all keys in the keyring DB are exhausted.
See https://github.com/cosmos/cosmos-sdk/pull/9695 for more details.

It is recommended to run in 'dry-run' mode first to verify all key migration material.
`,
		Args: cobra.NoArgs,
		RunE: runMigrateCmd,
	}

	return cmd
}

func runMigrateCmd(cmd *cobra.Command, _ []string) error {
	clientCtx, err := client.GetClientQueryContext(cmd)
	if err != nil {
		return err
	}

	if _, err = clientCtx.Keyring.MigrateAll(); err != nil {
		return err
	}

	buf := bufio.NewReader(cmd.InOrStdin())
	keyringServiceName := sdk.KeyringServiceName()

	var (
		tmpDir   string
		migrator keyring.Importer
	)

	if dryRun, _ := cmd.Flags().GetBool(flags.FlagDryRun); dryRun {
		tmpDir, err = os.MkdirTemp("", "migrator-migrate-dryrun")
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
		cmd.PrintErrln("Migration Aborted: no keys to migrate")
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
