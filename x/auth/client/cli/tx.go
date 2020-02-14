package cli

import (
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
)

// Codec defines the CLI codec to be used for use with the AccountRetriever.
// Client must be sure to set this to their respective codec that implements the
// AuthCodec interface and must be the same codec that passed to the x/auth
// module.
//
// TODO:/XXX: Using a package-level global isn't ideal and we should consider
// refactoring the module manager to allow passing in the correct module codec.
var Codec types.AuthCodec

// GetTxCmd returns the transaction commands for this module
func GetTxCmd(cdc *codec.Codec) *cobra.Command {
	txCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Auth transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}
	txCmd.AddCommand(
		GetMultiSignCommand(cdc),
		GetSignCommand(cdc),
	)
	return txCmd
}
