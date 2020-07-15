package cli

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd(queryRoute string, cdc *codec.Codec) *cobra.Command {
	stakingQueryCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Querying commands for the staking module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	stakingQueryCmd.AddCommand(
		GetCmdQueryDelegation(queryRoute, cdc),
		GetCmdQueryDelegations(queryRoute, cdc),
		GetCmdQueryUnbondingDelegation(queryRoute, cdc),
		GetCmdQueryUnbondingDelegations(queryRoute, cdc),
		GetCmdQueryRedelegation(queryRoute, cdc),
		GetCmdQueryRedelegations(queryRoute, cdc),
		GetCmdQueryValidator(queryRoute, cdc),
		GetCmdQueryValidators(queryRoute, cdc),
		GetCmdQueryValidatorDelegations(queryRoute, cdc),
		GetCmdQueryValidatorUnbondingDelegations(queryRoute, cdc),
		GetCmdQueryValidatorRedelegations(queryRoute, cdc),
		GetCmdQueryHistoricalInfo(queryRoute, cdc),
		GetCmdQueryParams(queryRoute, cdc),
		GetCmdQueryPool(queryRoute, cdc),
	)

	return stakingQueryCmd
}

// GetCmdQueryValidator implements the validator query command.
func GetCmdQueryValidator(storeName string, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validator [validator-addr]",
		Short: "Query a validator",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query details about an individual validator.

Example:
$ %s query staking validator cosmosvaloper1gghjut3ccd8ay0zduzj64hwre2fxs9ldmqhffj
`,
				version.AppName,
			),
		),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			queryClient := types.NewQueryClient(clientCtx)

			addr, err := sdk.ValAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			params := &types.QueryValidatorRequest{ValidatorAddr: addr}
			res, err := queryClient.Validator(cmd.Context(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintOutput(res.Validator)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdQueryValidators implements the query all validators command.
func GetCmdQueryValidators(storeName string, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validators",
		Short: "Query for all validators",
		Args:  cobra.NoArgs,
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query details about all validators on a network.

Example:
$ %s query staking validators
`,
				version.AppName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.NewContext().WithCodec(cdc).WithJSONMarshaler(cdc)

			resKVs, _, err := clientCtx.QuerySubspace(types.ValidatorsKey, storeName)
			if err != nil {
				return err
			}

			var validators types.Validators
			for _, kv := range resKVs {
				validator, err := types.UnmarshalValidator(types.ModuleCdc, kv.Value)
				if err != nil {
					return err
				}

				validators = append(validators, validator)
			}

			return clientCtx.PrintOutput(validators)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdQueryValidatorUnbondingDelegations implements the query all unbonding delegatations from a validator command.
func GetCmdQueryValidatorUnbondingDelegations(queryRoute string, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unbonding-delegations-from [validator-addr]",
		Short: "Query all unbonding delegatations from a validator",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query delegations that are unbonding _from_ a validator.

Example:
$ %s query staking unbonding-delegations-from cosmosvaloper1gghjut3ccd8ay0zduzj64hwre2fxs9ldmqhffj
`,
				version.AppName,
			),
		),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			queryClient := types.NewQueryClient(clientCtx)

			valAddr, err := sdk.ValAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			params := &types.QueryValidatorUnbondingDelegationsRequest{
				ValidatorAddr: valAddr,
				Req:           &query.PageRequest{},
			}

			res, err := queryClient.ValidatorUnbondingDelegations(context.Background(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintOutput(res.UnbondingResponses)
		},
	}

	cmd.Flags().Int(flags.FlagPage, 1, "pagination page of unbonding delegations to query for")
	cmd.Flags().Int(flags.FlagLimit, 100, "pagination limit of unbonding delegations to query for")
	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdQueryValidatorRedelegations implements the query all redelegatations
// from a validator command.
func GetCmdQueryValidatorRedelegations(queryRoute string, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "redelegations-from [validator-addr]",
		Short: "Query all outgoing redelegatations from a validator",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query delegations that are redelegating _from_ a validator.

Example:
$ %s query staking redelegations-from cosmosvaloper1gghjut3ccd8ay0zduzj64hwre2fxs9ldmqhffj
`,
				version.AppName,
			),
		),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			queryClient := types.NewQueryClient(clientCtx)

			valSrcAddr, err := sdk.ValAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			params := &types.QueryRedelegationsRequest{
				SrcValidatorAddr: valSrcAddr,
				Req:              &query.PageRequest{},
			}

			res, err := queryClient.Redelegations(context.Background(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintOutput(res.RedelegationResponses)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdQueryDelegation the query delegation command.
func GetCmdQueryDelegation(queryRoute string, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delegation [delegator-addr] [validator-addr]",
		Short: "Query a delegation based on address and validator address",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query delegations for an individual delegator on an individual validator.

Example:
$ %s query staking delegation cosmos1gghjut3ccd8ay0zduzj64hwre2fxs9ld75ru9p cosmosvaloper1gghjut3ccd8ay0zduzj64hwre2fxs9ldmqhffj
`,
				version.AppName,
			),
		),
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			queryClient := types.NewQueryClient(clientCtx)

			delAddr, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			valAddr, err := sdk.ValAddressFromBech32(args[1])
			if err != nil {
				return err
			}

			params := &types.QueryDelegationRequest{
				DelegatorAddr: delAddr,
				ValidatorAddr: valAddr,
			}

			res, err := queryClient.Delegation(context.Background(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintOutput(res.DelegationResponse)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdQueryDelegations implements the command to query all the delegations
// made from one delegator.
func GetCmdQueryDelegations(queryRoute string, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delegations [delegator-addr]",
		Short: "Query all delegations made by one delegator",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query delegations for an individual delegator on all validators.

Example:
$ %s query staking delegations cosmos1gghjut3ccd8ay0zduzj64hwre2fxs9ld75ru9p
`,
				version.AppName,
			),
		),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			queryClient := types.NewQueryClient(clientCtx)

			delAddr, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			params := &types.QueryDelegatorDelegationsRequest{
				DelegatorAddr: delAddr,
				Req:           &query.PageRequest{},
			}

			res, err := queryClient.DelegatorDelegations(context.Background(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintOutput(res.DelegationResponses)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdQueryValidatorDelegations implements the command to query all the
// delegations to a specific validator.
func GetCmdQueryValidatorDelegations(queryRoute string, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delegations-to [validator-addr]",
		Short: "Query all delegations made to one validator",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query delegations on an individual validator.

Example:
$ %s query staking delegations-to cosmosvaloper1gghjut3ccd8ay0zduzj64hwre2fxs9ldmqhffj
`,
				version.AppName,
			),
		),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			queryClient := types.NewQueryClient(clientCtx)

			valAddr, err := sdk.ValAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			params := &types.QueryValidatorDelegationsRequest{
				ValidatorAddr: valAddr,
				Req:           &query.PageRequest{},
			}

			res, err := queryClient.ValidatorDelegations(context.Background(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintOutput(res.DelegationResponses)
		},
	}

	cmd.Flags().Int(flags.FlagPage, 1, "pagination page of delegations to query for")
	cmd.Flags().Int(flags.FlagLimit, 100, "pagination limit of delegations to query for")
	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdQueryUnbondingDelegation implements the command to query a single
// unbonding-delegation record.
func GetCmdQueryUnbondingDelegation(queryRoute string, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unbonding-delegation [delegator-addr] [validator-addr]",
		Short: "Query an unbonding-delegation record based on delegator and validator address",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query unbonding delegations for an individual delegator on an individual validator.

Example:
$ %s query staking unbonding-delegation cosmos1gghjut3ccd8ay0zduzj64hwre2fxs9ld75ru9p cosmosvaloper1gghjut3ccd8ay0zduzj64hwre2fxs9ldmqhffj
`,
				version.AppName,
			),
		),
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			queryClient := types.NewQueryClient(clientCtx)

			valAddr, err := sdk.ValAddressFromBech32(args[1])
			if err != nil {
				return err
			}

			delAddr, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			params := &types.QueryUnbondingDelegationRequest{
				DelegatorAddr: delAddr,
				ValidatorAddr: valAddr,
			}

			res, err := queryClient.UnbondingDelegation(context.Background(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintOutput(res.Unbond)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdQueryUnbondingDelegations implements the command to query all the
// unbonding-delegation records for a delegator.
func GetCmdQueryUnbondingDelegations(queryRoute string, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unbonding-delegations [delegator-addr]",
		Short: "Query all unbonding-delegations records for one delegator",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query unbonding delegations for an individual delegator.

Example:
$ %s query staking unbonding-delegations cosmos1gghjut3ccd8ay0zduzj64hwre2fxs9ld75ru9p
`,
				version.AppName,
			),
		),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			queryClient := types.NewQueryClient(clientCtx)

			delegatorAddr, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			params := &types.QueryDelegatorUnbondingDelegationsRequest{
				DelegatorAddr: delegatorAddr,
				Req:           &query.PageRequest{},
			}

			res, err := queryClient.DelegatorUnbondingDelegations(context.Background(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintOutput(res.UnbondingResponses)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdQueryRedelegation implements the command to query a single
// redelegation record.
func GetCmdQueryRedelegation(queryRoute string, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "redelegation [delegator-addr] [src-validator-addr] [dst-validator-addr]",
		Short: "Query a redelegation record based on delegator and a source and destination validator address",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query a redelegation record for an individual delegator between a source and destination validator.

Example:
$ %s query staking redelegation cosmos1gghjut3ccd8ay0zduzj64hwre2fxs9ld75ru9p cosmosvaloper1l2rsakp388kuv9k8qzq6lrm9taddae7fpx59wm cosmosvaloper1gghjut3ccd8ay0zduzj64hwre2fxs9ldmqhffj
`,
				version.AppName,
			),
		),
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			queryClient := types.NewQueryClient(clientCtx)

			delAddr, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			valSrcAddr, err := sdk.ValAddressFromBech32(args[1])
			if err != nil {
				return err
			}

			valDstAddr, err := sdk.ValAddressFromBech32(args[2])
			if err != nil {
				return err
			}

			params := &types.QueryRedelegationsRequest{
				DelegatorAddr:    delAddr,
				DstValidatorAddr: valDstAddr,
				SrcValidatorAddr: valSrcAddr,
				Req:              &query.PageRequest{},
			}

			res, err := queryClient.Redelegations(context.Background(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintOutput(res.RedelegationResponses)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdQueryRedelegations implements the command to query all the
// redelegation records for a delegator.
func GetCmdQueryRedelegations(queryRoute string, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "redelegations [delegator-addr]",
		Args:  cobra.ExactArgs(1),
		Short: "Query all redelegations records for one delegator",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query all redelegation records for an individual delegator.

Example:
$ %s query staking redelegation cosmos1gghjut3ccd8ay0zduzj64hwre2fxs9ld75ru9p
`,
				version.AppName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			queryClient := types.NewQueryClient(clientCtx)

			delAddr, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			params := &types.QueryRedelegationsRequest{
				DelegatorAddr: delAddr,
				Req:           &query.PageRequest{},
			}

			res, err := queryClient.Redelegations(context.Background(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintOutput(res.RedelegationResponses)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdQueryHistoricalInfo implements the historical info query command
func GetCmdQueryHistoricalInfo(queryRoute string, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "historical-info [height]",
		Args:  cobra.ExactArgs(1),
		Short: "Query historical info at given height",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query historical info at given height.

Example:
$ %s query staking historical-info 5
`,
				version.AppName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			queryClient := types.NewQueryClient(clientCtx)

			height, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil || height < 0 {
				return fmt.Errorf("height argument provided must be a non-negative-integer: %v", err)
			}

			params := &types.QueryHistoricalInfoRequest{Height: height}
			res, err := queryClient.HistoricalInfo(context.Background(), params)

			if err != nil {
				return err
			}

			return clientCtx.PrintOutput(res.Hist)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdQueryPool implements the pool query command.
func GetCmdQueryPool(storeName string, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pool",
		Args:  cobra.NoArgs,
		Short: "Query the current staking pool values",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query values for amounts stored in the staking pool.

Example:
$ %s query staking pool
`,
				version.AppName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.Pool(context.Background(), &types.QueryPoolRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintOutput(res.Pool)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdQueryParams implements the params query command.
func GetCmdQueryParams(storeName string, cdc *codec.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "params",
		Args:  cobra.NoArgs,
		Short: "Query the current staking parameters information",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query values set as staking parameters.

Example:
$ %s query staking params
`,
				version.AppName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.Params(context.Background(), &types.QueryParamsRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintOutput(res.Params)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
