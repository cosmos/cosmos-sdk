package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd(queryRoute string, cdc *codec.Codec) *cobra.Command {
	poaQueryCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Querying commands for the POA module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}
	poaQueryCmd.AddCommand(client.GetCommands(
		GetCmdQueryValidator(queryRoute, cdc),
		GetCmdQueryValidators(queryRoute, cdc),
		GetCmdQueryParams(queryRoute, cdc))...)
	// GetCmdQueryPool(queryRoute, cdc))

	return poaQueryCmd

}

// GetCmdQueryValidator implements the validator query command.
func GetCmdQueryValidator(storeName string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "validator [validator-addr]",
		Short: "Query a validator",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query details about an individual validator.

Example:
$ %s query poa validator cosmosvaloper1gghjut3ccd8ay0zduzj64hwre2fxs9ldmqhffj
`,
				version.ClientName,
			),
		),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			addr, err := sdk.ValAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			res, _, err := cliCtx.QueryStore(types.GetValidatorKey(addr), storeName)
			if err != nil {
				return err
			}

			if len(res) == 0 {
				return fmt.Errorf("no validator found with address %s", addr)
			}

			return cliCtx.PrintOutput(types.MustUnmarshalValidator(cdc, res))
		},
	}
}

// GetCmdQueryValidators implements the query all validators command.
func GetCmdQueryValidators(storeName string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "validators",
		Short: "Query for all validators",
		Args:  cobra.NoArgs,
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query details about all validators on a network.

Example:
$ %s query validators
`,
				version.ClientName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			resKVs, _, err := cliCtx.QuerySubspace(types.ValidatorsKey, storeName)
			if err != nil {
				return err
			}

			var validators types.Validators
			for _, kv := range resKVs {
				validators = append(validators, types.MustUnmarshalValidator(cdc, kv.Value))
			}

			return cliCtx.PrintOutput(validators)
		},
	}
}

// GetCmdQueryParams implements the params query command.
func GetCmdQueryParams(storeName string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "params",
		Args:  cobra.NoArgs,
		Short: "Query the current staking parameters information",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query values set as staking parameters.

Example:
$ %s query poa params
`,
				version.ClientName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			route := fmt.Sprintf("custom/%s/%s", storeName, types.QueryParameters)
			bz, _, err := cliCtx.QueryWithData(route, nil)
			if err != nil {
				return err
			}

			var params types.Params
			cdc.MustUnmarshalJSON(bz, &params)
			return cliCtx.PrintOutput(params)
		},
	}
}

// // GetCmdQueryPool implements the pool query command.
// func GetCmdQueryPool(storeName string, cdc *codec.Codec) *cobra.Command {
// 	return &cobra.Command{
// 		Use:   "pool",
// 		Args:  cobra.NoArgs,
// 		Short: "Query the current staking pool values",
// 		Long: strings.TrimSpace(
// 			fmt.Sprintf(`Query values for amounts stored in the staking pool.

// Example:
// $ %s query staking pool
// `,
// 				version.ClientName,
// 			),
// 		),
// 		RunE: func(cmd *cobra.Command, args []string) error {
// 			cliCtx := context.NewCLIContext().WithCodec(cdc)

// 			bz, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/pool", storeName), nil)
// 			if err != nil {
// 				return err
// 			}

// 			var pool types.Pool
// 			if err := cdc.UnmarshalJSON(bz, &pool); err != nil {
// 				return err
// 			}

// 			return cliCtx.PrintOutput(pool)
// 		},
// 	}
// }
