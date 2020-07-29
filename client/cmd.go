package client

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/tendermint/tendermint/libs/cli"

	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ClientContextKey defines the context key used to retrieve a client.Context from
// a command's Context.
const ClientContextKey = sdk.ContextKey("client.context")

// SetCmdClientContextHandler is to be used in a command pre-hook execution to
// read flags that populate a Context and sets that to the command's Context.
func SetCmdClientContextHandler(clientCtx Context, cmd *cobra.Command) (err error) {
	clientCtx, err = ReadPersistentCommandFlags(clientCtx, cmd.Flags())
	if err != nil {
		return err
	}

	return SetCmdClientContext(cmd, clientCtx)
}

// ValidateCmd returns unknown command error or Help display if help flag set
func ValidateCmd(cmd *cobra.Command, args []string) error {
	var unknownCmd string
	var skipNext bool

	for _, arg := range args {
		// search for help flag
		if arg == "--help" || arg == "-h" {
			return cmd.Help()
		}

		// check if the current arg is a flag
		switch {
		case len(arg) > 0 && (arg[0] == '-'):
			// the next arg should be skipped if the current arg is a
			// flag and does not use "=" to assign the flag's value
			if !strings.Contains(arg, "=") {
				skipNext = true
			} else {
				skipNext = false
			}
		case skipNext:
			// skip current arg
			skipNext = false
		case unknownCmd == "":
			// unknown command found
			// continue searching for help flag
			unknownCmd = arg
		}
	}

	// return the help screen if no unknown command is found
	if unknownCmd != "" {
		err := fmt.Sprintf("unknown command \"%s\" for \"%s\"", unknownCmd, cmd.CalledAs())

		// build suggestions for unknown argument
		if suggestions := cmd.SuggestionsFor(unknownCmd); len(suggestions) > 0 {
			err += "\n\nDid you mean this?\n"
			for _, s := range suggestions {
				err += fmt.Sprintf("\t%v\n", s)
			}
		}
		return errors.New(err)
	}

	return cmd.Help()
}

// ReadPersistentCommandFlags returns a Context with fields set for "persistent"
// flags that do not necessarily change with context. These must be checked if
// the caller explicitly changed the values.
func ReadPersistentCommandFlags(clientCtx Context, flagSet *pflag.FlagSet) (Context, error) {
	if clientCtx.OutputFormat == "" {
		output, _ := flagSet.GetString(cli.OutputFlag)
		clientCtx = clientCtx.WithOutputFormat(output)
	}

	if clientCtx.HomeDir == "" {
		homeDir, _ := flagSet.GetString(flags.FlagHome)
		clientCtx = clientCtx.WithHomeDir(homeDir)
	}

	if clientCtx.ChainID == "" {
		chainID, _ := flagSet.GetString(flags.FlagChainID)
		clientCtx = clientCtx.WithChainID(chainID)
	}

	if clientCtx.Keyring == nil {
		keyringBackend, _ := flagSet.GetString(flags.FlagKeyringBackend)

		if keyringBackend != "" {
			kr, err := newKeyringFromFlags(clientCtx, keyringBackend)
			if err != nil {
				return clientCtx, err
			}

			clientCtx = clientCtx.WithKeyring(kr)
		}
	}

	if clientCtx.Client == nil {
		rpcURI, _ := flagSet.GetString(flags.FlagNode)
		if rpcURI != "" {
			clientCtx = clientCtx.WithNodeURI(rpcURI)
		}
	}

	return clientCtx, nil
}

// ReadQueryCommandFlags returns an updated Context with fields set based on flags
// defined in GetCommands. An error is returned if any flag query fails.
//
// Certain flags are naturally command and context dependent, so for these flags
// we do not check if they've been explicitly set by the caller. Other flags can
// be considered "persistent" (e.g. KeyBase or Client) and these should be checked
// if the caller explicitly set those.
func ReadQueryCommandFlags(clientCtx Context, flagSet *pflag.FlagSet) (Context, error) {
	if clientCtx.Height == 0 {
		height, _ := flagSet.GetInt64(flags.FlagHeight)
		clientCtx = clientCtx.WithHeight(height)
	}

	if !clientCtx.UseLedger {
		useLedger, _ := flagSet.GetBool(flags.FlagUseLedger)
		clientCtx = clientCtx.WithUseLedger(useLedger)
	}

	return ReadPersistentCommandFlags(clientCtx, flagSet)
}

// ReadTxCommandFlags returns an updated Context with fields set based on flags
// defined in PostCommands. An error is returned if any flag query fails.
//
// Certain flags are naturally command and context dependent, so for these flags
// we do not check if they've been explicitly set by the caller. Other flags can
// be considered "persistent" (e.g. KeyBase or Client) and these should be checked
// if the caller explicitly set those.
func ReadTxCommandFlags(clientCtx Context, flagSet *pflag.FlagSet) (Context, error) {
	clientCtx, err := ReadPersistentCommandFlags(clientCtx, flagSet)
	if err != nil {
		return clientCtx, err
	}

	if !clientCtx.GenerateOnly {
		genOnly, _ := flagSet.GetBool(flags.FlagGenerateOnly)
		clientCtx = clientCtx.WithGenerateOnly(genOnly)

	}

	if !clientCtx.Simulate {
		dryRun, _ := flagSet.GetBool(flags.FlagDryRun)
		clientCtx = clientCtx.WithSimulation(dryRun)
	}

	if !clientCtx.Offline {
		offline, _ := flagSet.GetBool(flags.FlagOffline)
		clientCtx = clientCtx.WithOffline(offline)
	}

	if !clientCtx.UseLedger {
		useLedger, _ := flagSet.GetBool(flags.FlagUseLedger)
		clientCtx = clientCtx.WithUseLedger(useLedger)
	}

	if clientCtx.BroadcastMode == "" {
		bMode, _ := flagSet.GetString(flags.FlagBroadcastMode)
		clientCtx = clientCtx.WithBroadcastMode(bMode)
	}

	if !clientCtx.SkipConfirm {
		skipConfirm, _ := flagSet.GetBool(flags.FlagSkipConfirmation)
		clientCtx = clientCtx.WithSkipConfirmation(skipConfirm)
	}

	if clientCtx.From == "" {
		from, _ := flagSet.GetString(flags.FlagFrom)
		fromAddr, fromName, err := GetFromFields(clientCtx.Keyring, from, clientCtx.GenerateOnly)
		if err != nil {
			return clientCtx, err
		}

		clientCtx = clientCtx.WithFrom(from).WithFromAddress(fromAddr).WithFromName(fromName)
	}

	return clientCtx, nil
}

// GetClientContextFromCmd returns a Context from a command or an empty Context
// if it has not been set.
func GetClientContextFromCmd(cmd *cobra.Command) Context {
	if v := cmd.Context().Value(ClientContextKey); v != nil {
		clientCtxPtr := v.(*Context)
		return *clientCtxPtr
	}

	return Context{}
}

// SetCmdClientContext sets a command's Context value to the provided argument.
func SetCmdClientContext(cmd *cobra.Command, clientCtx Context) error {
	v := cmd.Context().Value(ClientContextKey)
	if v == nil {
		return errors.New("client context not set")
	}

	clientCtxPtr := v.(*Context)
	*clientCtxPtr = clientCtx

	return nil
}
