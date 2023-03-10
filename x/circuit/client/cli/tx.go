package cli

import (
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/version"

	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/circuit/types"
	"github.com/spf13/cobra"
)

// NewTxCmd returns a root CLI command handler for all x/circuit transaction commands.
func NewTxCmd() *cobra.Command {
	txCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Circuit transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	txCmd.AddCommand(
		AuthorizeCircuitBreakerCmd(),
		TripCircuitBreakerCmd(),
	)

	return txCmd
}

// AuthorizeCircuitBreakerCmd returns a CLI command handler for creating a MsgAuthorizeCircuitBreaker transaction.
func AuthorizeCircuitBreakerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "authorize [grantee] [permission_level] [limit_type_urls] --from [granter]",
		Short: "Authorize an account to trip the circuit breaker.",
		Long: `Authorize an account to trip the circuit breaker.
		"SOME_MSGS" =     1,
		"ALL_MSGS" =      2,
		"SUPER_ADMIN" =   3,`,
		Example: fmt.Sprintf(`%s circuit authorize [address] 0 "cosmos.bank.v1beta1.MsgSend,cosmos.bank.v1beta1.MsgMultiSend"`, version.AppName),
		Args:    cobra.RangeArgs(3, 4),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			grantee, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			lvl, err := math.ParseUint(args[1])
			if err != nil {
				return err
			}

			var typeUrls []string
			if len(args) == 4 {
				typeUrls = strings.Split(args[2], ",")
			}

			permission := types.Permissions{Level: types.Permissions_Level(lvl.Uint64()), LimitTypeUrls: typeUrls}

			msg := types.NewMsgAuthorizeCircuitBreaker(clientCtx.GetFromAddress().String(), grantee.String(), &permission)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// TripCircuitBreakerCmd returns a CLI command handler for creating a MsgTripCircuitBreaker transaction.
func TripCircuitBreakerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "Disable [type_url]",
		Short: "Disable a message from being executed",
		Long: `Disable a message  from entering the mempool and/or being executed
		
		Example: 

		<app> circuit authorize trip "cosmos.bank.v1beta1.MsgSend,cosmos.bank.v1beta1.MsgMultiSend"
		`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmd.Flags().Set(flags.FlagFrom, args[0]); err != nil {
				return err
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			var msgTypeUrls []string
			if len(args) == 4 {
				msgTypeUrls = strings.Split(args[0], ",")
			}

			msg := types.NewMsgTripCircuitBreaker(clientCtx.GetFromAddress().String(), msgTypeUrls)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// ResetCircuitBreakerCmd returns a CLI command handler for creating a MsgRestCircuitBreaker transaction.
func ResetCircuitBreakerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reset [type_url]",
		Short: "Enable a message to be executed",
		Long: `Enable a message  that was disabled from entering the mempool and/or being executed
		
		Example: 

		<app> circuit authorize reset "cosmos.bank.v1beta1.MsgSend,cosmos.bank.v1beta1.MsgMultiSend"
		`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cmd.Flags().Set(flags.FlagFrom, args[0]); err != nil {
				return err
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			var msgTypeUrls []string
			if len(args) == 4 {
				msgTypeUrls = strings.Split(args[0], ",")
			}

			msg := types.NewMsgResetCircuitBreaker(clientCtx.GetFromAddress().String(), msgTypeUrls)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
