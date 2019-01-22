package cli

import (
	"fmt"

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
		Use:   "validator [operator-addr]",
		Short: "Query a validator",
		Args:  cobra.ExactArgs(1),
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
		Use:   "unbonding-delegations-from [operator-addr]",
		Short: "Query all unbonding delegatations from a validator",
		Args:  cobra.ExactArgs(1),
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

// GetCmdQueryValidatorRedelegations implements the query all redelegatations from a validator command.
func GetCmdQueryValidatorRedelegations(storeKey string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "redelegations-from [operator-addr]",
		Short: "Query all outgoing redelegatations from a validator",
		Args:  cobra.ExactArgs(1),
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

			route := fmt.Sprintf("custom/%s/%s", storeKey, staking.QueryValidatorRedelegations)
			res, err := cliCtx.QueryWithData(route, bz)
			if err != nil {
				return err
			}

			var reds staking.Redelegations
			cdc.MustUnmarshalJSON(res, &reds)
			return cliCtx.PrintOutput(reds)
		},
	}
}

// GetCmdQueryDelegation the query delegation command.
func GetCmdQueryDelegation(storeName string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "delegation [delegator-addr] [validator-addr]",
		Short: "Query a delegation based on address and validator address",
		Args:  cobra.ExactArgs(2),
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

			res, err := cliCtx.QueryStore(staking.GetDelegationKey(delAddr, valAddr), storeName)
			if err != nil {
				return err
			}

			delegation, err := types.UnmarshalDelegation(cdc, res)
			if err != nil {
				return err
			}

			return cliCtx.PrintOutput(delegation)
		},
	}
}

// GetCmdQueryDelegations implements the command to query all the delegations
// made from one delegator.
func GetCmdQueryDelegations(storeName string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "delegations [delegator-addr]",
		Short: "Query all delegations made from one delegator",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			delegatorAddr, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			resKVs, err := cliCtx.QuerySubspace(staking.GetDelegationsKey(delegatorAddr), storeName)
			if err != nil {
				return err
			}

			var delegations staking.Delegations
			for _, kv := range resKVs {
				delegations = append(delegations, types.MustUnmarshalDelegation(cdc, kv.Value))
			}

			return cliCtx.PrintOutput(delegations)
		},
	}
}

// GetCmdQueryValidatorDelegations implements the command to query all the
// delegations to a specific validator.
func GetCmdQueryValidatorDelegations(storeKey string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "delegations-to [validator-addr]",
		Short: "Query all delegations made to one validator",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			validatorAddr, err := sdk.ValAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			bz, err := cdc.MarshalJSON(staking.NewQueryValidatorParams(validatorAddr))
			if err != nil {
				return err
			}

			route := fmt.Sprintf("custom/%s/%s", storeKey, staking.QueryValidatorDelegations)
			res, err := cliCtx.QueryWithData(route, bz)
			if err != nil {
				return err
			}

			var dels staking.Delegations
			cdc.MustUnmarshalJSON(res, &dels)
			return cliCtx.PrintOutput(dels)
		},
	}
}

// GetCmdQueryUnbondingDelegation implements the command to query a single
// unbonding-delegation record.
func GetCmdQueryUnbondingDelegation(storeName string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "unbonding-delegation [delegator-addr] [validator-addr]",
		Short: "Query an unbonding-delegation record based on delegator and validator address",
		Args:  cobra.ExactArgs(2),
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
		Args:  cobra.ExactArgs(1),
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
func GetCmdQueryRedelegation(storeName string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "redelegation [delegator-addr] [validator-src-addr] [validator-dst-addr]",
		Short: "Query a redelegation record based on delegator and a source and destination validator address",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			valSrcAddr, err := sdk.ValAddressFromBech32(args[1])
			if err != nil {
				return err
			}

			valDstAddr, err := sdk.ValAddressFromBech32(args[2])
			if err != nil {
				return err
			}

			delAddr, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			res, err := cliCtx.QueryStore(staking.GetREDKey(delAddr, valSrcAddr, valDstAddr), storeName)
			if err != nil {
				return err
			}

			return cliCtx.PrintOutput(types.MustUnmarshalRED(cdc, res))
		},
	}
}

// GetCmdQueryRedelegations implements the command to query all the
// redelegation records for a delegator.
func GetCmdQueryRedelegations(storeName string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "redelegations [delegator-addr]",
		Short: "Query all redelegations records for one delegator",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			delegatorAddr, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			resKVs, err := cliCtx.QuerySubspace(staking.GetREDsKey(delegatorAddr), storeName)
			if err != nil {
				return err
			}

			var reds staking.Redelegations
			for _, kv := range resKVs {
				reds = append(reds, types.MustUnmarshalRED(cdc, kv.Value))
			}

			return cliCtx.PrintOutput(reds)
		},
	}
}

// GetCmdQueryPool implements the pool query command.
func GetCmdQueryPool(storeName string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "pool",
		Short: "Query the current staking pool values",
		Args:  cobra.NoArgs,
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
		Short: "Query the current staking parameters information",
		Args:  cobra.NoArgs,
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
