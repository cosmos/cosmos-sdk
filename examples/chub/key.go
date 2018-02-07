package main

import "github.com/spf13/cobra"

const (
	flagPassword    = "password"
	flagNewPassword = "new-password"
	flagType        = "type"
	flagSeed        = "seed"
	flagDryRun      = "dry-run"
)

var (
	listKeysCmd = &cobra.Command{
		Use:   "list",
		Short: "List all locally availably keys",
		RunE:  todoNotImplemented,
	}

	showKeysCmd = &cobra.Command{
		Use:   "show <name>",
		Short: "Show key info for the given name",
		RunE:  todoNotImplemented,
	}
)

// KeyCommands registers a sub-tree of commands to interact with
// local private key storage.
func KeyCommands() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "keys",
		Short: "Add or view local private keys",
	}
	cmd.AddCommand(
		addKeyCommand(),
		listKeysCmd,
		showKeysCmd,
		lineBreak,
		deleteKeyCommand(),
		updateKeyCommand(),
	)
	return cmd
}

func addKeyCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add <name>",
		Short: "Create a new key, or import from seed",
		RunE:  todoNotImplemented,
	}
	cmd.Flags().StringP(flagPassword, "p", "", "Password to encrypt private key")
	cmd.Flags().StringP(flagType, "t", "ed25519", "Type of private key (ed25519|secp256k1|ledger)")
	cmd.Flags().StringP(flagSeed, "s", "", "Provide seed phrase to recover existing key instead of creating")
	cmd.Flags().Bool(flagDryRun, false, "Perform action, but don't add key to local keystore")
	return cmd
}

func updateKeyCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <name>",
		Short: "Change the password used to protect private key",
		RunE:  todoNotImplemented,
	}
	cmd.Flags().StringP(flagPassword, "p", "", "Current password to decrypt key")
	cmd.Flags().String(flagNewPassword, "", "New password to use to protect key")
	return cmd
}

func deleteKeyCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <name>",
		Short: "Delete the given key",
		RunE:  todoNotImplemented,
	}
	cmd.Flags().StringP(flagPassword, "p", "", "Password of existing key to delete")
	return cmd
}
