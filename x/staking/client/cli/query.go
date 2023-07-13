package cli

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd() *cobra.Command {
	stakingQueryCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Querying commands for the staking module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	stakingQueryCmd.AddCommand(
		GetCmdQueryDelegation(),
		GetCmdQueryDelegations(),
		GetCmdQueryUnbondingDelegation(),
		GetCmdQueryUnbondingDelegations(),
		GetCmdQueryRedelegation(),
		GetCmdQueryRedelegations(),
		GetCmdQueryValidator(),
		GetCmdQueryValidators(),
		GetCmdQueryValidatorDelegations(),
		GetCmdQueryValidatorUnbondingDelegations(),
		GetCmdQueryValidatorRedelegations(),
		GetCmdQueryHistoricalInfo(),
		GetCmdQueryParams(),
		GetCmdQueryPool(),
		GetCmdQueryTokenizeShareRecordByID(),
		GetCmdQueryTokenizeShareRecordByDenom(),
		GetCmdQueryTokenizeShareRecordsOwned(),
		GetCmdQueryAllTokenizeShareRecords(),
		GetCmdQueryLastTokenizeShareRecordID(),
		GetCmdQueryTotalTokenizeSharedAssets(),
		GetCmdQueryTokenizeShareLockInfo(),
		GetCmdQueryTotalLiquidStaked(),
	)

	return stakingQueryCmd
}

// GetCmdQueryValidator implements the validator query command.
func GetCmdQueryValidator() *cobra.Command {
	bech32PrefixValAddr := sdk.GetConfig().GetBech32ValidatorAddrPrefix()

	cmd := &cobra.Command{
		Use:   "validator [validator-addr]",
		Short: "Query a validator",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query details about an individual validator.

Example:
$ %s query staking validator %s1gghjut3ccd8ay0zduzj64hwre2fxs9ldmqhffj
`,
				version.AppName, bech32PrefixValAddr,
			),
		),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			addr, err := sdk.ValAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			params := &types.QueryValidatorRequest{ValidatorAddr: addr.String()}
			res, err := queryClient.Validator(cmd.Context(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(&res.Validator)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdQueryValidators implements the query all validators command.
func GetCmdQueryValidators() *cobra.Command {
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
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)
			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			result, err := queryClient.Validators(cmd.Context(), &types.QueryValidatorsRequest{
				// Leaving status empty on purpose to query all validators.
				Pagination: pageReq,
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(result)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "validators")

	return cmd
}

// GetCmdQueryValidatorUnbondingDelegations implements the query all unbonding delegatations from a validator command.
func GetCmdQueryValidatorUnbondingDelegations() *cobra.Command {
	bech32PrefixValAddr := sdk.GetConfig().GetBech32ValidatorAddrPrefix()

	cmd := &cobra.Command{
		Use:   "unbonding-delegations-from [validator-addr]",
		Short: "Query all unbonding delegatations from a validator",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query delegations that are unbonding _from_ a validator.

Example:
$ %s query staking unbonding-delegations-from %s1gghjut3ccd8ay0zduzj64hwre2fxs9ldmqhffj
`,
				version.AppName, bech32PrefixValAddr,
			),
		),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			valAddr, err := sdk.ValAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			params := &types.QueryValidatorUnbondingDelegationsRequest{
				ValidatorAddr: valAddr.String(),
				Pagination:    pageReq,
			}

			res, err := queryClient.ValidatorUnbondingDelegations(cmd.Context(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "unbonding delegations")

	return cmd
}

// GetCmdQueryValidatorRedelegations implements the query all redelegatations
// from a validator command.
func GetCmdQueryValidatorRedelegations() *cobra.Command {
	bech32PrefixValAddr := sdk.GetConfig().GetBech32ValidatorAddrPrefix()

	cmd := &cobra.Command{
		Use:   "redelegations-from [validator-addr]",
		Short: "Query all outgoing redelegatations from a validator",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query delegations that are redelegating _from_ a validator.

Example:
$ %s query staking redelegations-from %s1gghjut3ccd8ay0zduzj64hwre2fxs9ldmqhffj
`,
				version.AppName, bech32PrefixValAddr,
			),
		),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			valSrcAddr, err := sdk.ValAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			params := &types.QueryRedelegationsRequest{
				SrcValidatorAddr: valSrcAddr.String(),
				Pagination:       pageReq,
			}

			res, err := queryClient.Redelegations(cmd.Context(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "validator redelegations")

	return cmd
}

// GetCmdQueryDelegation the query delegation command.
func GetCmdQueryDelegation() *cobra.Command {
	bech32PrefixAccAddr := sdk.GetConfig().GetBech32AccountAddrPrefix()
	bech32PrefixValAddr := sdk.GetConfig().GetBech32ValidatorAddrPrefix()

	cmd := &cobra.Command{
		Use:   "delegation [delegator-addr] [validator-addr]",
		Short: "Query a delegation based on address and validator address",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query delegations for an individual delegator on an individual validator.

Example:
$ %s query staking delegation %s1gghjut3ccd8ay0zduzj64hwre2fxs9ld75ru9p %s1gghjut3ccd8ay0zduzj64hwre2fxs9ldmqhffj
`,
				version.AppName, bech32PrefixAccAddr, bech32PrefixValAddr,
			),
		),
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
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
				DelegatorAddr: delAddr.String(),
				ValidatorAddr: valAddr.String(),
			}

			res, err := queryClient.Delegation(cmd.Context(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res.DelegationResponse)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdQueryDelegations implements the command to query all the delegations
// made from one delegator.
func GetCmdQueryDelegations() *cobra.Command {
	bech32PrefixAccAddr := sdk.GetConfig().GetBech32AccountAddrPrefix()

	cmd := &cobra.Command{
		Use:   "delegations [delegator-addr]",
		Short: "Query all delegations made by one delegator",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query delegations for an individual delegator on all validators.

Example:
$ %s query staking delegations %s1gghjut3ccd8ay0zduzj64hwre2fxs9ld75ru9p
`,
				version.AppName, bech32PrefixAccAddr,
			),
		),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			delAddr, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			params := &types.QueryDelegatorDelegationsRequest{
				DelegatorAddr: delAddr.String(),
				Pagination:    pageReq,
			}

			res, err := queryClient.DelegatorDelegations(cmd.Context(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "delegations")

	return cmd
}

// GetCmdQueryValidatorDelegations implements the command to query all the
// delegations to a specific validator.
func GetCmdQueryValidatorDelegations() *cobra.Command {
	bech32PrefixValAddr := sdk.GetConfig().GetBech32ValidatorAddrPrefix()

	cmd := &cobra.Command{
		Use:   "delegations-to [validator-addr]",
		Short: "Query all delegations made to one validator",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query delegations on an individual validator.

Example:
$ %s query staking delegations-to %s1gghjut3ccd8ay0zduzj64hwre2fxs9ldmqhffj
`,
				version.AppName, bech32PrefixValAddr,
			),
		),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			valAddr, err := sdk.ValAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			params := &types.QueryValidatorDelegationsRequest{
				ValidatorAddr: valAddr.String(),
				Pagination:    pageReq,
			}

			res, err := queryClient.ValidatorDelegations(cmd.Context(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "validator delegations")

	return cmd
}

// GetCmdQueryUnbondingDelegation implements the command to query a single
// unbonding-delegation record.
func GetCmdQueryUnbondingDelegation() *cobra.Command {
	bech32PrefixAccAddr := sdk.GetConfig().GetBech32AccountAddrPrefix()
	bech32PrefixValAddr := sdk.GetConfig().GetBech32ValidatorAddrPrefix()

	cmd := &cobra.Command{
		Use:   "unbonding-delegation [delegator-addr] [validator-addr]",
		Short: "Query an unbonding-delegation record based on delegator and validator address",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query unbonding delegations for an individual delegator on an individual validator.

Example:
$ %s query staking unbonding-delegation %s1gghjut3ccd8ay0zduzj64hwre2fxs9ld75ru9p %s1gghjut3ccd8ay0zduzj64hwre2fxs9ldmqhffj
`,
				version.AppName, bech32PrefixAccAddr, bech32PrefixValAddr,
			),
		),
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
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
				DelegatorAddr: delAddr.String(),
				ValidatorAddr: valAddr.String(),
			}

			res, err := queryClient.UnbondingDelegation(cmd.Context(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(&res.Unbond)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdQueryUnbondingDelegations implements the command to query all the
// unbonding-delegation records for a delegator.
func GetCmdQueryUnbondingDelegations() *cobra.Command {
	bech32PrefixAccAddr := sdk.GetConfig().GetBech32AccountAddrPrefix()

	cmd := &cobra.Command{
		Use:   "unbonding-delegations [delegator-addr]",
		Short: "Query all unbonding-delegations records for one delegator",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query unbonding delegations for an individual delegator.

Example:
$ %s query staking unbonding-delegations %s1gghjut3ccd8ay0zduzj64hwre2fxs9ld75ru9p
`,
				version.AppName, bech32PrefixAccAddr,
			),
		),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			delegatorAddr, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			params := &types.QueryDelegatorUnbondingDelegationsRequest{
				DelegatorAddr: delegatorAddr.String(),
				Pagination:    pageReq,
			}

			res, err := queryClient.DelegatorUnbondingDelegations(cmd.Context(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "unbonding delegations")

	return cmd
}

// GetCmdQueryRedelegation implements the command to query a single
// redelegation record.
func GetCmdQueryRedelegation() *cobra.Command {
	bech32PrefixAccAddr := sdk.GetConfig().GetBech32AccountAddrPrefix()
	bech32PrefixValAddr := sdk.GetConfig().GetBech32ValidatorAddrPrefix()

	cmd := &cobra.Command{
		Use:   "redelegation [delegator-addr] [src-validator-addr] [dst-validator-addr]",
		Short: "Query a redelegation record based on delegator and a source and destination validator address",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query a redelegation record for an individual delegator between a source and destination validator.

Example:
$ %s query staking redelegation %s1gghjut3ccd8ay0zduzj64hwre2fxs9ld75ru9p %s1l2rsakp388kuv9k8qzq6lrm9taddae7fpx59wm %s1gghjut3ccd8ay0zduzj64hwre2fxs9ldmqhffj
`,
				version.AppName, bech32PrefixAccAddr, bech32PrefixValAddr, bech32PrefixValAddr,
			),
		),
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
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
				DelegatorAddr:    delAddr.String(),
				DstValidatorAddr: valDstAddr.String(),
				SrcValidatorAddr: valSrcAddr.String(),
			}

			res, err := queryClient.Redelegations(cmd.Context(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdQueryRedelegations implements the command to query all the
// redelegation records for a delegator.
func GetCmdQueryRedelegations() *cobra.Command {
	bech32PrefixAccAddr := sdk.GetConfig().GetBech32AccountAddrPrefix()

	cmd := &cobra.Command{
		Use:   "redelegations [delegator-addr]",
		Args:  cobra.ExactArgs(1),
		Short: "Query all redelegations records for one delegator",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query all redelegation records for an individual delegator.

Example:
$ %s query staking redelegation %s1gghjut3ccd8ay0zduzj64hwre2fxs9ld75ru9p
`,
				version.AppName, bech32PrefixAccAddr,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			delAddr, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			params := &types.QueryRedelegationsRequest{
				DelegatorAddr: delAddr.String(),
				Pagination:    pageReq,
			}

			res, err := queryClient.Redelegations(cmd.Context(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "delegator redelegations")

	return cmd
}

// GetCmdQueryHistoricalInfo implements the historical info query command
func GetCmdQueryHistoricalInfo() *cobra.Command {
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
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			height, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil || height < 0 {
				return fmt.Errorf("height argument provided must be a non-negative-integer: %v", err)
			}

			params := &types.QueryHistoricalInfoRequest{Height: height}
			res, err := queryClient.HistoricalInfo(cmd.Context(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res.Hist)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdQueryPool implements the pool query command.
func GetCmdQueryPool() *cobra.Command {
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
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.Pool(cmd.Context(), &types.QueryPoolRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(&res.Pool)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdQueryParams implements the params query command.
func GetCmdQueryParams() *cobra.Command {
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
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.Params(cmd.Context(), &types.QueryParamsRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(&res.Params)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdQueryTokenizeShareRecordById implements the query for individual tokenize share record information by share by id
func GetCmdQueryTokenizeShareRecordByID() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tokenize-share-record-by-id [id]",
		Args:  cobra.ExactArgs(1),
		Short: "Query individual tokenize share record information by share by id",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query individual tokenize share record information by share by id.

Example:
$ %s query staking tokenize-share-record-by-id [id]
`,
				version.AppName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			id, err := strconv.Atoi(args[0])
			if err != nil {
				return err
			}

			res, err := queryClient.TokenizeShareRecordById(cmd.Context(), &types.QueryTokenizeShareRecordByIdRequest{
				Id: uint64(id),
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdQueryTokenizeShareRecordByDenom implements the query for individual tokenize share record information by share denom
func GetCmdQueryTokenizeShareRecordByDenom() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tokenize-share-record-by-denom",
		Args:  cobra.ExactArgs(1),
		Short: "Query individual tokenize share record information by share denom",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query individual tokenize share record information by share denom.

Example:
$ %s query staking tokenize-share-record-by-denom
`,
				version.AppName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.TokenizeShareRecordByDenom(cmd.Context(), &types.QueryTokenizeShareRecordByDenomRequest{
				Denom: args[0],
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdQueryTokenizeShareRecordsOwned implements the query tokenize share records by address
func GetCmdQueryTokenizeShareRecordsOwned() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tokenize-share-records-owned",
		Args:  cobra.ExactArgs(1),
		Short: "Query tokenize share records by address",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query tokenize share records by address.

Example:
$ %s query staking tokenize-share-records-owned [owner]
`,
				version.AppName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			owner, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			res, err := queryClient.TokenizeShareRecordsOwned(cmd.Context(), &types.QueryTokenizeShareRecordsOwnedRequest{
				Owner: owner.String(),
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdQueryAllTokenizeShareRecords implements the query for all tokenize share records
func GetCmdQueryAllTokenizeShareRecords() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "all-tokenize-share-records",
		Args:  cobra.NoArgs,
		Short: "Query for all tokenize share records",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query for all tokenize share records.

Example:
$ %s query staking all-tokenize-share-records
`,
				version.AppName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			params := &types.QueryAllTokenizeShareRecordsRequest{
				Pagination: pageReq,
			}

			res, err := queryClient.AllTokenizeShareRecords(cmd.Context(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "tokenize share records")

	return cmd
}

// GetCmdQueryLastTokenizeShareRecordId implements the query for last tokenize share record id
func GetCmdQueryLastTokenizeShareRecordID() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "last-tokenize-share-record-id",
		Args:  cobra.NoArgs,
		Short: "Query for last tokenize share record id",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query for last tokenize share record id.

Example:
$ %s query staking last-tokenize-share-record-id
`,
				version.AppName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.LastTokenizeShareRecordId(cmd.Context(), &types.QueryLastTokenizeShareRecordIdRequest{})
			if err != nil {
				return err
			}

			// log a dummy print statement to check we're propagating to dockernet
			fmt.Println("lalala")

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdQueryTotalTokenizeSharedAssets implements the query for total tokenized staked assets
func GetCmdQueryTotalTokenizeSharedAssets() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "total-tokenize-share-assets",
		Args:  cobra.NoArgs,
		Short: "Query for total tokenized staked assets",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query for total tokenized staked assets.

Example:
$ %s query staking total-tokenize-share-assets
`,
				version.AppName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.TotalTokenizeSharedAssets(cmd.Context(), &types.QueryTotalTokenizeSharedAssetsRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdQueryTotalLiquidStaked implements the query for total liquid staked tokens
func GetCmdQueryTotalLiquidStaked() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "total-liquid-staked",
		Args:  cobra.NoArgs,
		Short: "Query for total liquid staked tokens",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query for total number of liquid staked tokens.
Liquid staked tokens are identified as either a tokenized delegation, 
or tokens owned by an interchain account.
Example:
$ %s query staking total-liquid-staked
`,
				version.AppName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.TotalLiquidStaked(cmd.Context(), &types.QueryTotalLiquidStaked{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// GetCmdQueryTokenizeShareLockInfo returns the tokenize share lock status for a user
func GetCmdQueryTokenizeShareLockInfo() *cobra.Command {
	bech32PrefixAccAddr := sdk.GetConfig().GetBech32AccountAddrPrefix()

	cmd := &cobra.Command{
		Use:   "tokenize-share-lock-info [address]",
		Args:  cobra.ExactArgs(1),
		Short: "Query tokenize share lock information",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query the status of a tokenize share lock for a given account
Example:
$ %s query staking tokenize-share-lock-info %s1gghjut3ccd8ay0zduzj64hwre2fxs9ldmqhffj
`,
				version.AppName, bech32PrefixAccAddr,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			address := args[0]
			if _, err := sdk.AccAddressFromBech32(address); err != nil {
				return err
			}

			res, err := queryClient.TokenizeShareLockInfo(
				cmd.Context(),
				&types.QueryTokenizeShareLockInfo{Address: address},
			)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}
