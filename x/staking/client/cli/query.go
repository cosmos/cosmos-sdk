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

// GetCmdQueryPool implements the pool query command.
func GetCmdQueryMsgCreateValidators() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "buffered-create-validators",
		Args:  cobra.NoArgs,
		Short: "Query the queued MsgCreateValidator",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query queued CreateValidators messages.

Example:
$ %s query staking buffered-create-validators
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

			res, err := queryClient.QueuedMsgCreateValidators(cmd.Context(), &types.QueryQueuedMsgCreateValidatorsRequest{
				Pagination: pageReq,
			})

			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "msgCreateValidators")

	return cmd
}

// GetCmdQueryMsgEditValidators implements the msgEditValidators query command.
func GetCmdQueryMsgEditValidators() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "buffered-edit-validators",
		Args:  cobra.NoArgs,
		Short: "Query the queued MsgEditValidator",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query queued editValidator messages.

Example:
$ %s query staking buffered-edit-validators
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

			res, err := queryClient.QueuedMsgEditValidators(cmd.Context(), &types.QueryQueuedMsgEditValidatorsRequest{
				Pagination: pageReq,
			})

			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "msgEditValidators")

	return cmd
}

// GetCmdQueryMsgDelegate implements the delegate msg query command.
func GetCmdQueryMsgDelegate() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "buffered-delegation-message [address]",
		Args:  cobra.ExactArgs(1),
		Short: "Query for specific queued delegate messages",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query queued delegation messages.

Example:
$ %s query staking buffered-delegation-messages [delegator_address]
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

			delAddr, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			res, err := queryClient.QueuedMsgDelegate(cmd.Context(), &types.QueryQueuedMsgDelegateRequest{
				DelegatorAddr: delAddr.String(),
				Pagination:    pageReq,
			})

			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "delegation")

	return cmd
}

// GetCmdQueryMsgDelegates implements the delegate msg query command.
func GetCmdQueryMsgDelegates() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "buffered-delegation-messages",
		Args:  cobra.NoArgs,
		Short: "Query all queued delegation messages",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query queued delegation messages.

Example:
$ %s query staking buffered-delegation-messages
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

			res, err := queryClient.QueuedMsgDelegates(cmd.Context(), &types.QueryQueuedMsgDelegatesRequest{
				Pagination: pageReq,
			})

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

// GetCmdQueryMsgDelegates implements the delegate msg query command.
func GetCmdQueryMsgUndelegates() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "buffered-undelegation-messages",
		Args:  cobra.NoArgs,
		Short: "Query all queued undelegation messages",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query queued undelegation messages.

Example:
$ %s query staking buffered-undelegation-messages
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

			res, err := queryClient.QueuedMsgUndelegates(cmd.Context(), &types.QueryQueuedMsgUndelegatesRequest{
				Pagination: pageReq,
			})

			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "undelegations")

	return cmd
}

// GetCmdQueryMsgDelegate implements the delegate msg query command.
func GetCmdQueryMsgUndelegate() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "buffered-undelegation-messages [delegator_addr]",
		Args:  cobra.ExactArgs(1),
		Short: "Query all queued undelegation messages",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query queued undelegation messages.

Example:
$ %s query staking buffered-undelegation-messages
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

			delAddr, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			res, err := queryClient.QueuedMsgUndelegate(cmd.Context(), &types.QueryQueuedMsgUndelegateRequest{
				DelegatorAddr: delAddr.String(),
				Pagination:    pageReq,
			})

			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "undelegations")

	return cmd
}

// GetCmdQueryMsgDelegates implements the delegate msg query command.
func GetCmdQueryMsgBeginRedelegates() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "buffered-begin-redelegate-messages",
		Args:  cobra.NoArgs,
		Short: "Query all queued being redelegation messages",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query queued begin redelegation messages.

Example:
$ %s query staking buffered-begin-redelegate-messages
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

			res, err := queryClient.QueuedMsgBeginRedelegates(cmd.Context(), &types.QueryQueuedMsgBeginRedelegatesRequest{
				Pagination: pageReq,
			})

			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "begin redelegations")

	return cmd
}

// GetCmdQueryMsgDelegates implements the delegate msg query command.
func GetCmdQueryMsgBeginRedelegate() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "buffered-begin-redelegate-messages [delegator_addr]",
		Args:  cobra.ExactArgs(1),
		Short: "Query specific queued begin redelegation messages",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query queued begin redelegation messages.

Example:
$ %s query staking buffered-undelegation-messages
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

			delAddr, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			res, err := queryClient.QueuedMsgBeginRedelegate(cmd.Context(), &types.QueryQueuedMsgBeginRedelegateRequest{
				DelegatorAddr: delAddr.String(),
				Pagination:    pageReq,
			})

			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "redelegations")

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
