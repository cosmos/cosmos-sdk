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
	stakingQueryCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Querying commands for the staking module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}
	stakingQueryCmd.AddCommand(client.GetCommands(
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
		GetCmdQueryParams(queryRoute, cdc),
		GetCmdQueryPool(queryRoute, cdc))...)

	return stakingQueryCmd

}

// GetCmdQueryValidator implements the validator query command.
func GetCmdQueryValidator(storeName string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "validator [validator-addr]",
		Short: "Query a validator",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query details about an individual validator.

Example:
$ %s query staking validator cosmosvaloper1gghjut3ccd8ay0zduzj64hwre2fxs9ldmqhffj
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
				return fmt.Errorf("No validator found with address %s", addr)
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
$ %s query staking validators
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

// GetCmdQueryValidatorUnbondingDelegations implements the query all unbonding delegatations from a validator command.
func GetCmdQueryValidatorUnbondingDelegations(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "unbonding-delegations-from [validator-addr]",
		Short: "Query all unbonding delegatations from a validator",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query delegations that are unbonding _from_ a validator.

Example:
$ %s query staking unbonding-delegations-from cosmosvaloper1gghjut3ccd8ay0zduzj64hwre2fxs9ldmqhffj
`,
				version.ClientName,
			),
		),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			valAddr, err := sdk.ValAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			bz, err := cdc.MarshalJSON(types.NewQueryValidatorParams(valAddr))
			if err != nil {
				return err
			}

			route := fmt.Sprintf("custom/%s/%s", queryRoute, types.QueryValidatorUnbondingDelegations)
			res, _, err := cliCtx.QueryWithData(route, bz)
			if err != nil {
				return err
			}

			var ubds types.UnbondingDelegations
			cdc.MustUnmarshalJSON(res, &ubds)
			return cliCtx.PrintOutput(ubds)
		},
	}
}

// GetCmdQueryValidatorRedelegations implements the query all redelegatations
// from a validator command.
func GetCmdQueryValidatorRedelegations(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "redelegations-from [validator-addr]",
		Short: "Query all outgoing redelegatations from a validator",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query delegations that are redelegating _from_ a validator.

Example:
$ %s query staking redelegations-from cosmosvaloper1gghjut3ccd8ay0zduzj64hwre2fxs9ldmqhffj
`,
				version.ClientName,
			),
		),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			valSrcAddr, err := sdk.ValAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			bz, err := cdc.MarshalJSON(types.QueryRedelegationParams{SrcValidatorAddr: valSrcAddr})
			if err != nil {
				return err
			}

			route := fmt.Sprintf("custom/%s/%s", queryRoute, types.QueryRedelegations)
			res, _, err := cliCtx.QueryWithData(route, bz)
			if err != nil {
				return err
			}

			var resp types.RedelegationResponses
			if err := cdc.UnmarshalJSON(res, &resp); err != nil {
				return err
			}

			return cliCtx.PrintOutput(resp)
		},
	}
}

// GetCmdQueryDelegation the query delegation command.
func GetCmdQueryDelegation(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "delegation [delegator-addr] [validator-addr]",
		Short: "Query a delegation based on address and validator address",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query delegations for an individual delegator on an individual validator.

Example:
$ %s query staking delegation cosmos1gghjut3ccd8ay0zduzj64hwre2fxs9ld75ru9p cosmosvaloper1gghjut3ccd8ay0zduzj64hwre2fxs9ldmqhffj
`,
				version.ClientName,
			),
		),
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			delAddr, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			valAddr, err := sdk.ValAddressFromBech32(args[1])
			if err != nil {
				return err
			}

			bz, err := cdc.MarshalJSON(types.NewQueryBondsParams(delAddr, valAddr))
			if err != nil {
				return err
			}

			route := fmt.Sprintf("custom/%s/%s", queryRoute, types.QueryDelegation)
			res, _, err := cliCtx.QueryWithData(route, bz)
			if err != nil {
				return err
			}

			var resp types.DelegationResponse
			if err := cdc.UnmarshalJSON(res, &resp); err != nil {
				return err
			}

			return cliCtx.PrintOutput(resp)
		},
	}
}

// GetCmdQueryDelegations implements the command to query all the delegations
// made from one delegator.
func GetCmdQueryDelegations(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "delegations [delegator-addr]",
		Short: "Query all delegations made by one delegator",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query delegations for an individual delegator on all validators.

Example:
$ %s query staking delegations cosmos1gghjut3ccd8ay0zduzj64hwre2fxs9ld75ru9p
`,
				version.ClientName,
			),
		),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			delAddr, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			bz, err := cdc.MarshalJSON(types.NewQueryDelegatorParams(delAddr))
			if err != nil {
				return err
			}

			route := fmt.Sprintf("custom/%s/%s", queryRoute, types.QueryDelegatorDelegations)
			res, _, err := cliCtx.QueryWithData(route, bz)
			if err != nil {
				return err
			}

			var resp types.DelegationResponses
			if err := cdc.UnmarshalJSON(res, &resp); err != nil {
				return err
			}

			return cliCtx.PrintOutput(resp)
		},
	}
}

// GetCmdQueryValidatorDelegations implements the command to query all the
// delegations to a specific validator.
func GetCmdQueryValidatorDelegations(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "delegations-to [validator-addr]",
		Short: "Query all delegations made to one validator",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query delegations on an individual validator.

Example:
$ %s query staking delegations-to cosmosvaloper1gghjut3ccd8ay0zduzj64hwre2fxs9ldmqhffj
`,
				version.ClientName,
			),
		),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			valAddr, err := sdk.ValAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			bz, err := cdc.MarshalJSON(types.NewQueryValidatorParams(valAddr))
			if err != nil {
				return err
			}

			route := fmt.Sprintf("custom/%s/%s", queryRoute, types.QueryValidatorDelegations)
			res, _, err := cliCtx.QueryWithData(route, bz)
			if err != nil {
				return err
			}

			var resp types.DelegationResponses
			if err := cdc.UnmarshalJSON(res, &resp); err != nil {
				return err
			}

			return cliCtx.PrintOutput(resp)
		},
	}
}

// GetCmdQueryUnbondingDelegation implements the command to query a single
// unbonding-delegation record.
func GetCmdQueryUnbondingDelegation(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "unbonding-delegation [delegator-addr] [validator-addr]",
		Short: "Query an unbonding-delegation record based on delegator and validator address",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query unbonding delegations for an individual delegator on an individual validator.

Example:
$ %s query staking unbonding-delegation cosmos1gghjut3ccd8ay0zduzj64hwre2fxs9ld75ru9p cosmosvaloper1gghjut3ccd8ay0zduzj64hwre2fxs9ldmqhffj
`,
				version.ClientName,
			),
		),
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			valAddr, err := sdk.ValAddressFromBech32(args[1])
			if err != nil {
				return err
			}

			delAddr, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			bz, err := cdc.MarshalJSON(types.NewQueryBondsParams(delAddr, valAddr))
			if err != nil {
				return err
			}

			route := fmt.Sprintf("custom/%s/%s", queryRoute, types.QueryUnbondingDelegation)
			res, _, err := cliCtx.QueryWithData(route, bz)
			if err != nil {
				return err
			}

			return cliCtx.PrintOutput(types.MustUnmarshalUBD(cdc, res))
		},
	}
}

// GetCmdQueryUnbondingDelegations implements the command to query all the
// unbonding-delegation records for a delegator.
func GetCmdQueryUnbondingDelegations(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "unbonding-delegations [delegator-addr]",
		Short: "Query all unbonding-delegations records for one delegator",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query unbonding delegations for an individual delegator.

Example:
$ %s query staking unbonding-delegation cosmos1gghjut3ccd8ay0zduzj64hwre2fxs9ld75ru9p
`,
				version.ClientName,
			),
		),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			delegatorAddr, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			bz, err := cdc.MarshalJSON(types.NewQueryDelegatorParams(delegatorAddr))
			if err != nil {
				return err
			}

			route := fmt.Sprintf("custom/%s/%s", queryRoute, types.QueryDelegatorUnbondingDelegations)
			res, _, err := cliCtx.QueryWithData(route, bz)
			if err != nil {
				return err
			}

			var ubds types.UnbondingDelegations
			if err = cdc.UnmarshalJSON(res, &ubds); err != nil {
				return err
			}

			return cliCtx.PrintOutput(ubds)
		},
	}
}

// GetCmdQueryRedelegation implements the command to query a single
// redelegation record.
func GetCmdQueryRedelegation(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "redelegation [delegator-addr] [src-validator-addr] [dst-validator-addr]",
		Short: "Query a redelegation record based on delegator and a source and destination validator address",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query a redelegation record for an individual delegator between a source and destination validator.

Example:
$ %s query staking redelegation cosmos1gghjut3ccd8ay0zduzj64hwre2fxs9ld75ru9p cosmosvaloper1l2rsakp388kuv9k8qzq6lrm9taddae7fpx59wm cosmosvaloper1gghjut3ccd8ay0zduzj64hwre2fxs9ldmqhffj
`,
				version.ClientName,
			),
		),
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

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

			bz, err := cdc.MarshalJSON(types.NewQueryRedelegationParams(delAddr, valSrcAddr, valDstAddr))
			if err != nil {
				return err
			}

			route := fmt.Sprintf("custom/%s/%s", queryRoute, types.QueryRedelegations)
			res, _, err := cliCtx.QueryWithData(route, bz)
			if err != nil {
				return err
			}

			var resp types.RedelegationResponses
			if err := cdc.UnmarshalJSON(res, &resp); err != nil {
				return err
			}

			return cliCtx.PrintOutput(resp)
		},
	}
}

// GetCmdQueryRedelegations implements the command to query all the
// redelegation records for a delegator.
func GetCmdQueryRedelegations(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "redelegations [delegator-addr]",
		Args:  cobra.ExactArgs(1),
		Short: "Query all redelegations records for one delegator",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query all redelegation records for an individual delegator.

Example:
$ %s query staking redelegation cosmos1gghjut3ccd8ay0zduzj64hwre2fxs9ld75ru9p
`,
				version.ClientName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			delAddr, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			bz, err := cdc.MarshalJSON(types.QueryRedelegationParams{DelegatorAddr: delAddr})
			if err != nil {
				return err
			}

			route := fmt.Sprintf("custom/%s/%s", queryRoute, types.QueryRedelegations)
			res, _, err := cliCtx.QueryWithData(route, bz)
			if err != nil {
				return err
			}

			var resp types.RedelegationResponses
			if err := cdc.UnmarshalJSON(res, &resp); err != nil {
				return err
			}

			return cliCtx.PrintOutput(resp)
		},
	}
}

// GetCmdQueryPool implements the pool query command.
func GetCmdQueryPool(storeName string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "pool",
		Args:  cobra.NoArgs,
		Short: "Query the current staking pool values",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query values for amounts stored in the staking pool.

Example:
$ %s query staking pool
`,
				version.ClientName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			bz, _, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/pool", storeName), nil)
			if err != nil {
				return err
			}

			var pool types.Pool
			if err := cdc.UnmarshalJSON(bz, &pool); err != nil {
				return err
			}

			return cliCtx.PrintOutput(pool)
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
$ %s query staking params
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
