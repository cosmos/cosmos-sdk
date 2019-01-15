package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
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
			addr, err := sdk.ValAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			key := staking.GetValidatorKey(addr)
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			res, err := cliCtx.QueryStore(key, storeName)
			if err != nil {
				return err
			} else if len(res) == 0 {
				return fmt.Errorf("No validator found with address %s", args[0])
			}

			client.PrintOutput(cdc, types.MustUnmarshalValidator(cdc, res))
			return nil
		},
	}
}

// GetCmdQueryValidators implements the query all validators command.
func GetCmdQueryValidators(storeName string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "validators",
		Short: "Query for all validators",
		RunE: func(cmd *cobra.Command, args []string) error {
			key := staking.ValidatorsKey
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			resKVs, err := cliCtx.QuerySubspace(key, storeName)
			if err != nil {
				return err
			}

			// parse out the validators
			var validators staking.Validators
			for _, kv := range resKVs {
				validators = append(validators, types.MustUnmarshalValidator(cdc, kv.Value))
			}

			client.PrintOutput(cdc, validators)
			return nil
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
			valAddr, err := sdk.ValAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			cliCtx := context.NewCLIContext().WithCodec(cdc)
			params := staking.NewQueryValidatorParams(valAddr)

			bz, err := cdc.MarshalJSON(params)
			if err != nil {
				return err
			}

			route := fmt.Sprintf("custom/%s/validatorUnbondingDelegations", storeKey)
			res, err := cliCtx.QueryWithData(route, bz)
			if err != nil {
				return err
			}

			var ubds staking.UnbondingDelegations
			cdc.MustUnmarshalJSON(res, &ubds)
			client.PrintOutput(cdc, ubds)
			return nil
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
			valAddr, err := sdk.ValAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			cliCtx := context.NewCLIContext().WithCodec(cdc)
			params := staking.NewQueryValidatorParams(valAddr)

			bz, err := cdc.MarshalJSON(params)
			if err != nil {
				return err
			}

			route := fmt.Sprintf("custom/%s/validatorRedelegations", storeKey)
			res, err := cliCtx.QueryWithData(route, bz)
			if err != nil {
				return err
			}

			var rds staking.Redelegations
			cdc.MustUnmarshalJSON(res, &rds)
			client.PrintOutput(cdc, rds)
			return nil
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
			valAddr, err := sdk.ValAddressFromBech32(args[1])
			if err != nil {
				return err
			}

			delAddr, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			key := staking.GetDelegationKey(delAddr, valAddr)
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			res, err := cliCtx.QueryStore(key, storeName)
			if err != nil {
				return err
			}

			// parse out the delegation
			delegation, err := types.UnmarshalDelegation(cdc, res)
			if err != nil {
				return err
			}

			client.PrintOutput(cdc, delegation)
			return nil
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
			delegatorAddr, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			key := staking.GetDelegationsKey(delegatorAddr)
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			resKVs, err := cliCtx.QuerySubspace(key, storeName)
			if err != nil {
				return err
			}

			// parse out the validators
			var delegations staking.Delegations
			for _, kv := range resKVs {
				delegations = append(delegations, types.MustUnmarshalDelegation(cdc, kv.Value))
			}

			client.PrintOutput(cdc, delegations)
			return nil
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
			validatorAddr, err := sdk.ValAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			params := staking.NewQueryValidatorParams(validatorAddr)

			bz, err := cdc.MarshalJSON(params)
			if err != nil {
				return err
			}

			cliCtx := context.NewCLIContext().WithCodec(cdc)

			route := fmt.Sprintf("custom/%s/validatorDelegations", storeKey)
			res, err := cliCtx.QueryWithData(route, bz)
			if err != nil {
				return err
			}

			var dels staking.Delegations
			cdc.MustUnmarshalJSON(res, &dels)
			client.PrintOutput(cdc, dels)
			return nil
		},
	}
}

// GetCmdQueryUnbondingDelegation implements the command to query a single
// unbonding-delegation record.
func GetCmdQueryUnbondingDelegation(storeName string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "unbonding-delegation [delegator-addr] [validator-addr]",
		Args:  cobra.ExactArgs(2),
		Short: "Query an unbonding-delegation record based on delegator and validator address",
		RunE: func(cmd *cobra.Command, args []string) error {
			valAddr, err := sdk.ValAddressFromBech32(args[1])
			if err != nil {
				return err
			}

			delAddr, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			key := staking.GetUBDKey(delAddr, valAddr)
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			res, err := cliCtx.QueryStore(key, storeName)
			if err != nil {
				return err
			}

			client.PrintOutput(cdc, types.MustUnmarshalUBD(cdc, res))
			return nil
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
			delegatorAddr, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			key := staking.GetUBDsKey(delegatorAddr)
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			resKVs, err := cliCtx.QuerySubspace(key, storeName)
			if err != nil {
				return err
			}

			// parse out the unbonding delegations
			var ubds staking.UnbondingDelegations
			for _, kv := range resKVs {
				ubds = append(ubds, types.MustUnmarshalUBD(cdc, kv.Value))
			}

			client.PrintOutput(cdc, ubds)
			return nil
		},
	}
}

// GetCmdQueryRedelegation implements the command to query a single
// redelegation record.
func GetCmdQueryRedelegation(storeName string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "redelegation [delegator-addr] [validator-src-addr] [validator-dst-addr]",
		Args:  cobra.ExactArgs(3),
		Short: "Query a redelegation record based on delegator and src/dst validator addresses",
		RunE: func(cmd *cobra.Command, args []string) error {
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

			key := staking.GetREDKey(delAddr, valSrcAddr, valDstAddr)
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			res, err := cliCtx.QueryStore(key, storeName)
			if err != nil {
				return err
			}

			client.PrintOutput(cdc, types.MustUnmarshalRED(cdc, res))
			return nil
		},
	}
}

// GetCmdQueryRedelegations implements the command to query all the
// redelegation records for a delegator.
func GetCmdQueryRedelegations(storeName string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "redelegations [delegator-addr]",
		Args:  cobra.ExactArgs(1),
		Short: "Query all redelegations records for one delegator",
		RunE: func(cmd *cobra.Command, args []string) error {
			delegatorAddr, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			key := staking.GetREDsKey(delegatorAddr)
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			resKVs, err := cliCtx.QuerySubspace(key, storeName)
			if err != nil {
				return err
			}

			// parse out the validators
			var reds staking.Redelegations
			for _, kv := range resKVs {
				reds = append(reds, types.MustUnmarshalRED(cdc, kv.Value))
			}

			client.PrintOutput(cdc, reds)
			return nil
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
			key := staking.PoolKey
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			res, err := cliCtx.QueryStore(key, storeName)
			if err != nil {
				return err
			}

			client.PrintOutput(cdc, types.MustUnmarshalPool(cdc, res))
			return nil
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
			bz, err := cliCtx.QueryWithData("custom/staking/"+staking.QueryParameters, nil)
			if err != nil {
				return err
			}

			var params staking.Params
			cdc.MustUnmarshalJSON(bz, &params)
			client.PrintOutput(cdc, params)
			return nil
		},
	}
}
