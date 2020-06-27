package client

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/cosmos/cosmos-sdk/client/flags"
)

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

// ReadGetCommandFlags returns an updated Context with fields set based on flags
// defined in GetCommands. An error is returned if any flag query fails.
func ReadGetCommandFlags(clientCtx Context, flagSet *pflag.FlagSet) (Context, error) {
	if flagSet.Changed(flags.FlagHeight) {
		height, err := flagSet.GetInt64(flags.FlagHeight)
		if err != nil {
			return clientCtx, err
		}

		clientCtx = clientCtx.WithHeight(height)
	}

	if flagSet.Changed(flags.FlagTrustNode) {
		trustNode, err := flagSet.GetBool(flags.FlagTrustNode)
		if err != nil {
			return clientCtx, err
		}

		clientCtx = clientCtx.WithTrustNode(trustNode)
	}

	if flagSet.Changed(flags.FlagUseLedger) {
		useLedger, err := flagSet.GetBool(flags.FlagUseLedger)
		if err != nil {
			return clientCtx, err
		}

		clientCtx = clientCtx.WithUseLedger(useLedger)
	}

	if clientCtx.Keyring == nil && flagSet.Changed(flags.FlagKeyringBackend) {
		keyringBackend, err := flagSet.GetString(flags.FlagKeyringBackend)
		if err != nil {
			return clientCtx, err
		}

		kr, err := newKeyringFromFlags(clientCtx, keyringBackend)
		if err != nil {
			return clientCtx, err
		}

		clientCtx = clientCtx.WithKeyring(kr)
	}

	if clientCtx.Client == nil && flagSet.Changed(flags.FlagNode) {
		rpcURI, err := flagSet.GetString(flags.FlagNode)
		if err != nil {
			return clientCtx, err
		}

		clientCtx = clientCtx.WithNodeURI(rpcURI)
	}

	return clientCtx, nil
}
