package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

// GetCmdQueryValidator implements the validator query command.
func GetCmdQueryValidator(storeName string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "validator [validator-addr]",
		Short: "Query a validator",
		Long: strings.TrimSpace(`Query details about an individual validator:

$ <appcli> query staking validator cosmosvaloper1gghjut3ccd8ay0zduzj64hwre2fxs9ldmqhffj
`),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			addr, err := sdk.ValAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			res, err := cliCtx.QueryStore(staking.GetValidatorKey(addr), storeName)
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
		Long: strings.TrimSpace(`Query details about all validators on a network:

$ <appcli> query staking validators
`),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			resKVs, err := cliCtx.QuerySubspace(staking.ValidatorsKey, storeName)
			if err != nil {
				return err
			}

			var validators staking.Validators
			for _, kv := range resKVs {
				validators = append(validators, types.MustUnmarshalValidator(cdc, kv.Value))
			}

			return cliCtx.PrintOutput(validators)
		},
	}
}

// GetCmdQueryValidatorUnbondingDelegations implements the query all unbonding delegatations from a validator command.
func GetCmdQueryValidatorUnbondingDelegations(storeKey string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "unbonding-delegations-from [validator-addr]",
		Short: "Query all unbonding delegatations from a validator",
		Long: strings.TrimSpace(`Query delegations that are unbonding _from_ a validator:

$ <appcli> query staking unbonding-delegations-from cosmosvaloper1gghjut3ccd8ay0zduzj64hwre2fxs9ldmqhffj
`),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			valAddr, err := sdk.ValAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			bz, err := cdc.MarshalJSON(staking.NewQueryValidatorParams(valAddr))
			if err != nil {
				return err
			}

			route := fmt.Sprintf("custom/%s/%s", storeKey, staking.QueryValidatorUnbondingDelegations)
			res, err := cliCtx.QueryWithData(route, bz)
			if err != nil {
				return err
			}

			var ubds staking.UnbondingDelegations
			cdc.MustUnmarshalJSON(res, &ubds)
			return cliCtx.PrintOutput(ubds)
		},
	}
}

// GetCmdQueryValidatorRedelegations implements the query all redelegatations
// from a validator command.
func GetCmdQueryValidatorRedelegations(storeKey string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "redelegations-from [validator-addr]",
		Short: "Query all outgoing redelegatations from a validator",
		Long: strings.TrimSpace(`Query delegations that are redelegating _from_ a validator:

$ <appcli> query staking redelegations-from cosmosvaloper1gghjut3ccd8ay0zduzj64hwre2fxs9ldmqhffj
`),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			valSrcAddr, err := sdk.ValAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			bz, err := cdc.MarshalJSON(staking.QueryRedelegationParams{SrcValidatorAddr: valSrcAddr})
			if err != nil {
				return err
			}

			route := fmt.Sprintf("custom/%s/%s", storeKey, staking.QueryRedelegations)
			res, err := cliCtx.QueryWithData(route, bz)
			if err != nil {
				return err
			}

			var resp staking.RedelegationResponses
			if err := cdc.UnmarshalJSON(res, &resp); err != nil {
				return err
			}

			return cliCtx.PrintOutput(resp)
		},
	}
}

// GetCmdQueryDelegation the query delegation command.
func GetCmdQueryDelegation(storeKey string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "delegation [delegator-addr] [validator-addr]",
		Short: "Query a delegation based on address and validator address",
		Long: strings.TrimSpace(`Query delegations for an individual delegator on an individual validator:

$ <appcli> query staking delegation cosmos1gghjut3ccd8ay0zduzj64hwre2fxs9ld75ru9p cosmosvaloper1gghjut3ccd8ay0zduzj64hwre2fxs9ldmqhffj
`),
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

			bz, err := cdc.MarshalJSON(staking.NewQueryBondsParams(delAddr, valAddr))
			if err != nil {
				return err
			}

			route := fmt.Sprintf("custom/%s/%s", storeKey, staking.QueryDelegation)
			res, err := cliCtx.QueryWithData(route, bz)
			if err != nil {
				return err
			}

			var resp staking.DelegationResp
			if err := cdc.UnmarshalJSON(res, &resp); err != nil {
				return err
			}

			return cliCtx.PrintOutput(resp)
		},
	}
}

// GetCmdQueryDelegations implements the command to query all the delegations
// made from one delegator.
func GetCmdQueryDelegations(storeKey string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "delegations [delegator-addr]",
		Short: "Query all delegations made by one delegator",
		Long: strings.TrimSpace(`Query delegations for an individual delegator on all validators:

$ <appcli> query staking delegations cosmos1gghjut3ccd8ay0zduzj64hwre2fxs9ld75ru9p
`),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			delAddr, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			bz, err := cdc.MarshalJSON(staking.NewQueryDelegatorParams(delAddr))
			if err != nil {
				return err
			}

			route := fmt.Sprintf("custom/%s/%s", storeKey, staking.QueryDelegatorDelegations)
			res, err := cliCtx.QueryWithData(route, bz)
			if err != nil {
				return err
			}

			var resp staking.DelegationResponses
			if err := cdc.UnmarshalJSON(res, &resp); err != nil {
				return err
			}

			return cliCtx.PrintOutput(resp)
		},
	}
}

// GetCmdQueryValidatorDelegations implements the command to query all the
// delegations to a specific validator.
func GetCmdQueryValidatorDelegations(storeKey string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "delegations-to [validator-addr]",
		Short: "Query all delegations made to one validator",
		Long: strings.TrimSpace(`Query delegations on an individual validator:

$ <appcli> query staking delegations-to cosmosvaloper1gghjut3ccd8ay0zduzj64hwre2fxs9ldmqhffj
`),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			valAddr, err := sdk.ValAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			bz, err := cdc.MarshalJSON(staking.NewQueryValidatorParams(valAddr))
			if err != nil {
				return err
			}

			route := fmt.Sprintf("custom/%s/%s", storeKey, staking.QueryValidatorDelegations)
			res, err := cliCtx.QueryWithData(route, bz)
			if err != nil {
				return err
			}

			var resp staking.DelegationResponses
			if err := cdc.UnmarshalJSON(res, &resp); err != nil {
				return err
			}

			return cliCtx.PrintOutput(resp)
		},
	}
}

// GetCmdQueryUnbondingDelegation implements the command to query a single
// unbonding-delegation record.
func GetCmdQueryUnbondingDelegation(storeName string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "unbonding-delegation [delegator-addr] [validator-addr]",
		Short: "Query an unbonding-delegation record based on delegator and validator address",
		Long: strings.TrimSpace(`Query unbonding delegations for an individual delegator on an individual validator:

$ <appcli> query staking unbonding-delegation cosmos1gghjut3ccd8ay0zduzj64hwre2fxs9ld75ru9p cosmosvaloper1gghjut3ccd8ay0zduzj64hwre2fxs9ldmqhffj
`),
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

			res, err := cliCtx.QueryStore(staking.GetUBDKey(delAddr, valAddr), storeName)
			if err != nil {
				return err
			}

			return cliCtx.PrintOutput(types.MustUnmarshalUBD(cdc, res))

		},
	}
}

// GetCmdQueryUnbondingDelegations implements the command to query all the
// unbonding-delegation records for a delegator.
func GetCmdQueryUnbondingDelegations(storeName string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "unbonding-delegations [delegator-addr]",
		Short: "Query all unbonding-delegations records for one delegator",
		Long: strings.TrimSpace(`Query unbonding delegations for an individual delegator:

$ <appcli> query staking unbonding-delegation cosmos1gghjut3ccd8ay0zduzj64hwre2fxs9ld75ru9p
`),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			delegatorAddr, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			resKVs, err := cliCtx.QuerySubspace(staking.GetUBDsKey(delegatorAddr), storeName)
			if err != nil {
				return err
			}

			var ubds staking.UnbondingDelegations
			for _, kv := range resKVs {
				ubds = append(ubds, types.MustUnmarshalUBD(cdc, kv.Value))
			}

			return cliCtx.PrintOutput(ubds)
		},
	}
}

// GetCmdQueryRedelegation implements the command to query a single
// redelegation record.
func GetCmdQueryRedelegation(storeKey string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "redelegation [delegator-addr] [src-validator-addr] [dst-validator-addr]",
		Short: "Query a redelegation record based on delegator and a source and destination validator address",
		Long: strings.TrimSpace(`Query a redelegation record  for an individual delegator between a source and destination validator:

$ <appcli> query staking redelegation cosmos1gghjut3ccd8ay0zduzj64hwre2fxs9ld75ru9p cosmosvaloper1l2rsakp388kuv9k8qzq6lrm9taddae7fpx59wm cosmosvaloper1gghjut3ccd8ay0zduzj64hwre2fxs9ldmqhffj
`),
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

			bz, err := cdc.MarshalJSON(staking.NewQueryRedelegationParams(delAddr, valSrcAddr, valDstAddr))
			if err != nil {
				return err
			}

			route := fmt.Sprintf("custom/%s/%s", storeKey, staking.QueryRedelegations)
			res, err := cliCtx.QueryWithData(route, bz)
			if err != nil {
				return err
			}

			var resp staking.RedelegationResponses
			if err := cdc.UnmarshalJSON(res, &resp); err != nil {
				return err
			}

			return cliCtx.PrintOutput(resp)
		},
	}
}

// GetCmdQueryRedelegations implements the command to query all the
// redelegation records for a delegator.
func GetCmdQueryRedelegations(storeKey string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "redelegations [delegator-addr]",
		Args:  cobra.ExactArgs(1),
		Short: "Query all redelegations records for one delegator",
		Long: strings.TrimSpace(`Query all redelegation records for an individual delegator:

$ <appcli> query staking redelegation cosmos1gghjut3ccd8ay0zduzj64hwre2fxs9ld75ru9p
`),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			delAddr, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			bz, err := cdc.MarshalJSON(staking.QueryRedelegationParams{DelegatorAddr: delAddr})
			if err != nil {
				return err
			}

			route := fmt.Sprintf("custom/%s/%s", storeKey, staking.QueryRedelegations)
			res, err := cliCtx.QueryWithData(route, bz)
			if err != nil {
				return err
			}

			var resp staking.RedelegationResponses
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
		Long: strings.TrimSpace(`Query values for amounts stored in the staking pool:

$ <appcli> query staking pool
`),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			res, err := cliCtx.QueryStore(staking.PoolKey, storeName)
			if err != nil {
				return err
			}

			return cliCtx.PrintOutput(types.MustUnmarshalPool(cdc, res))
		},
	}
}

// GetCmdQueryPool implements the params query command.
func GetCmdQueryParams(storeName string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "params",
		Args:  cobra.NoArgs,
		Short: "Query the current staking parameters information",
		Long: strings.TrimSpace(`Query values set as staking parameters:

$ <appcli> query staking params
`),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			route := fmt.Sprintf("custom/%s/%s", storeName, staking.QueryParameters)
			bz, err := cliCtx.QueryWithData(route, nil)
			if err != nil {
				return err
			}

			var params staking.Params
			cdc.MustUnmarshalJSON(bz, &params)
			return cliCtx.PrintOutput(params)
		},
	}
}
