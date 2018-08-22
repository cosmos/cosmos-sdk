package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tendermint/tendermint/libs/cli"

	"github.com/cosmos/cosmos-sdk/client/context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/cosmos/cosmos-sdk/x/stake"
	"github.com/cosmos/cosmos-sdk/x/stake/types"
)

// GetCmdQueryValidator implements the validator query command.
func GetCmdQueryValidator(storeName string, cdc *wire.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validator [owner-addr]",
		Short: "Query a validator",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			addr, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			key := stake.GetValidatorKey(addr)
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			res, err := cliCtx.QueryStore(key, storeName)
			if err != nil {
				return err
			} else if len(res) == 0 {
				return fmt.Errorf("No validator found with address %s", args[0])
			}

			validator := types.MustUnmarshalValidator(cdc, addr, res)

			switch viper.Get(cli.OutputFlag) {
			case "text":
				human, err := validator.HumanReadableString()
				if err != nil {
					return err
				}
				fmt.Println(human)

			case "json":
				// parse out the validator
				output, err := wire.MarshalJSONIndent(cdc, validator)
				if err != nil {
					return err
				}

				fmt.Println(string(output))
			}

			// TODO: output with proofs / machine parseable etc.
			return nil
		},
	}

	return cmd
}

// GetCmdQueryValidators implements the query all validators command.
func GetCmdQueryValidators(storeName string, cdc *wire.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validators",
		Short: "Query for all validators",
		RunE: func(cmd *cobra.Command, args []string) error {
			key := stake.ValidatorsKey
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			resKVs, err := cliCtx.QuerySubspace(key, storeName)
			if err != nil {
				return err
			}

			// parse out the validators
			var validators []stake.Validator
			for _, kv := range resKVs {
				addr := kv.Key[1:]
				validator := types.MustUnmarshalValidator(cdc, addr, kv.Value)
				validators = append(validators, validator)
			}

			switch viper.Get(cli.OutputFlag) {
			case "text":
				for _, validator := range validators {
					resp, err := validator.HumanReadableString()
					if err != nil {
						return err
					}

					fmt.Println(resp)
				}
			case "json":
				output, err := wire.MarshalJSONIndent(cdc, validators)
				if err != nil {
					return err
				}

				fmt.Println(string(output))
				return nil
			}

			// TODO: output with proofs / machine parseable etc.
			return nil
		},
	}

	return cmd
}

// GetCmdQueryDelegation the query delegation command.
func GetCmdQueryDelegation(storeName string, cdc *wire.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delegation",
		Short: "Query a delegation based on address and validator address",
		RunE: func(cmd *cobra.Command, args []string) error {
			valAddr, err := sdk.AccAddressFromBech32(viper.GetString(FlagAddressValidator))
			if err != nil {
				return err
			}

			delAddr, err := sdk.AccAddressFromBech32(viper.GetString(FlagAddressDelegator))
			if err != nil {
				return err
			}

			key := stake.GetDelegationKey(delAddr, valAddr)
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			res, err := cliCtx.QueryStore(key, storeName)
			if err != nil {
				return err
			}

			// parse out the delegation
			delegation := types.MustUnmarshalDelegation(cdc, key, res)

			switch viper.Get(cli.OutputFlag) {
			case "text":
				resp, err := delegation.HumanReadableString()
				if err != nil {
					return err
				}

				fmt.Println(resp)
			case "json":
				output, err := wire.MarshalJSONIndent(cdc, delegation)
				if err != nil {
					return err
				}

				fmt.Println(string(output))
				return nil
			}

			return nil
		},
	}

	cmd.Flags().AddFlagSet(fsValidator)
	cmd.Flags().AddFlagSet(fsDelegator)

	return cmd
}

// GetCmdQueryDelegations implements the command to query all the delegations
// made from one delegator.
func GetCmdQueryDelegations(storeName string, cdc *wire.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delegations [delegator-addr]",
		Short: "Query all delegations made from one delegator",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			delegatorAddr, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			key := stake.GetDelegationsKey(delegatorAddr)
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			resKVs, err := cliCtx.QuerySubspace(key, storeName)
			if err != nil {
				return err
			}

			// parse out the validators
			var delegations []stake.Delegation
			for _, kv := range resKVs {
				delegation := types.MustUnmarshalDelegation(cdc, kv.Key, kv.Value)
				delegations = append(delegations, delegation)
			}

			output, err := wire.MarshalJSONIndent(cdc, delegations)
			if err != nil {
				return err
			}

			fmt.Println(string(output))

			// TODO: output with proofs / machine parseable etc.
			return nil
		},
	}

	return cmd
}

// GetCmdQueryUnbondingDelegation implements the command to query a single
// unbonding-delegation record.
func GetCmdQueryUnbondingDelegation(storeName string, cdc *wire.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unbonding-delegation",
		Short: "Query an unbonding-delegation record based on delegator and validator address",
		RunE: func(cmd *cobra.Command, args []string) error {
			valAddr, err := sdk.AccAddressFromBech32(viper.GetString(FlagAddressValidator))
			if err != nil {
				return err
			}

			delAddr, err := sdk.AccAddressFromBech32(viper.GetString(FlagAddressDelegator))
			if err != nil {
				return err
			}

			key := stake.GetUBDKey(delAddr, valAddr)
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			res, err := cliCtx.QueryStore(key, storeName)
			if err != nil {
				return err
			}

			// parse out the unbonding delegation
			ubd := types.MustUnmarshalUBD(cdc, key, res)

			switch viper.Get(cli.OutputFlag) {
			case "text":
				resp, err := ubd.HumanReadableString()
				if err != nil {
					return err
				}

				fmt.Println(resp)
			case "json":
				output, err := wire.MarshalJSONIndent(cdc, ubd)
				if err != nil {
					return err
				}

				fmt.Println(string(output))
				return nil
			}

			return nil
		},
	}

	cmd.Flags().AddFlagSet(fsValidator)
	cmd.Flags().AddFlagSet(fsDelegator)

	return cmd
}

// GetCmdQueryUnbondingDelegations implements the command to query all the
// unbonding-delegation records for a delegator.
func GetCmdQueryUnbondingDelegations(storeName string, cdc *wire.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unbonding-delegations [delegator-addr]",
		Short: "Query all unbonding-delegations records for one delegator",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			delegatorAddr, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			key := stake.GetUBDsKey(delegatorAddr)
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			resKVs, err := cliCtx.QuerySubspace(key, storeName)
			if err != nil {
				return err
			}

			// parse out the validators
			var ubds []stake.UnbondingDelegation
			for _, kv := range resKVs {
				ubd := types.MustUnmarshalUBD(cdc, kv.Key, kv.Value)
				ubds = append(ubds, ubd)
			}

			output, err := wire.MarshalJSONIndent(cdc, ubds)
			if err != nil {
				return err
			}

			fmt.Println(string(output))

			// TODO: output with proofs / machine parseable etc.
			return nil
		},
	}

	return cmd
}

// GetCmdQueryRedelegation implements the command to query a single
// redelegation record.
func GetCmdQueryRedelegation(storeName string, cdc *wire.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "redelegation",
		Short: "Query a redelegation record based on delegator and a source and destination validator address",
		RunE: func(cmd *cobra.Command, args []string) error {
			valSrcAddr, err := sdk.AccAddressFromBech32(viper.GetString(FlagAddressValidatorSrc))
			if err != nil {
				return err
			}

			valDstAddr, err := sdk.AccAddressFromBech32(viper.GetString(FlagAddressValidatorDst))
			if err != nil {
				return err
			}

			delAddr, err := sdk.AccAddressFromBech32(viper.GetString(FlagAddressDelegator))
			if err != nil {
				return err
			}

			key := stake.GetREDKey(delAddr, valSrcAddr, valDstAddr)
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			res, err := cliCtx.QueryStore(key, storeName)
			if err != nil {
				return err
			}

			// parse out the unbonding delegation
			red := types.MustUnmarshalRED(cdc, key, res)

			switch viper.Get(cli.OutputFlag) {
			case "text":
				resp, err := red.HumanReadableString()
				if err != nil {
					return err
				}

				fmt.Println(resp)
			case "json":
				output, err := wire.MarshalJSONIndent(cdc, red)
				if err != nil {
					return err
				}

				fmt.Println(string(output))
				return nil
			}

			return nil
		},
	}

	cmd.Flags().AddFlagSet(fsRedelegation)
	cmd.Flags().AddFlagSet(fsDelegator)

	return cmd
}

// GetCmdQueryRedelegations implements the command to query all the
// redelegation records for a delegator.
func GetCmdQueryRedelegations(storeName string, cdc *wire.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "redelegations [delegator-addr]",
		Short: "Query all redelegations records for one delegator",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			delegatorAddr, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			key := stake.GetREDsKey(delegatorAddr)
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			resKVs, err := cliCtx.QuerySubspace(key, storeName)
			if err != nil {
				return err
			}

			// parse out the validators
			var reds []stake.Redelegation
			for _, kv := range resKVs {
				red := types.MustUnmarshalRED(cdc, kv.Key, kv.Value)
				reds = append(reds, red)
			}

			output, err := wire.MarshalJSONIndent(cdc, reds)
			if err != nil {
				return err
			}

			fmt.Println(string(output))

			// TODO: output with proofs / machine parseable etc.
			return nil
		},
	}

	return cmd
}
