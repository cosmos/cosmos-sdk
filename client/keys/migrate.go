package keys

import (
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
)

// MigrateCommand migrates key information from legacy keybase to OS secret store.
func MigrateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Migrate keys from amino to proto serialization format",
		Long: `Migrate key information from legacyInfo(amino) to record(proto).
		LegacyInfo is an interface that is used to persist keys in keyring DB using Amino serialization format.
Record is a struct that is used to persist keys in keyring DB using Proto for serialization and deserialization.
For each key material entry, the command will check if the key can be deserialized using proto.
If this is the case, the key is already migrated. Therefore, we skip it and continue with a next one. 
Otherwise, we try to deserialize it using Amino to legacyInfo. If this attempt is successful, we serialize legacyInfo to proto serialization format. 
Finally, we overwrite keyring entry with new keyring.Item. 
If any error occured, it will be outputed in CLI. In this case, migration will be continued until all keys in keyring DB are exhausted.

It is recommended to run in 'dry-run' mode first to verify all key migration material.
`,
		Args: cobra.NoArgs,
		RunE: runMigrateCmd,
	}

	cmd.Flags().Bool(flags.FlagDryRun, false, "Run migration without actually persisting any changes to the new Keybase")
	return cmd
}

func runMigrateCmd(cmd *cobra.Command, args []string) error {

	clientCtx, err := client.GetClientQueryContext(cmd)
	if err != nil {
		return err
	}

	if _, err := clientCtx.Keyring.MigrateAll(); err != nil {
		return err
	}

	return nil

}
