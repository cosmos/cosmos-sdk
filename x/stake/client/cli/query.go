package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tendermint/tmlibs/cli"

	"github.com/cosmos/cosmos-sdk/client/context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/cosmos/cosmos-sdk/x/stake"
)

// get the command to query a validator
func GetCmdQueryValidator(storeName string, cdc *wire.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validator [owner-addr]",
		Short: "Query a validator",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {

			addr, err := sdk.GetAccAddressBech32(args[0])
			if err != nil {
				return err
			}
			key := stake.GetValidatorKey(addr)
			ctx := context.NewCoreContextFromViper()
			res, err := ctx.QueryStore(key, storeName)
			if err != nil {
				return err
			}
			validator := new(stake.Validator)
			cdc.MustUnmarshalBinary(res, validator)

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
			// TODO output with proofs / machine parseable etc.
			return nil
		},
	}

	return cmd
}

// get the command to query a validator
func GetCmdQueryValidators(storeName string, cdc *wire.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validators",
		Short: "Query for all validators",
		RunE: func(cmd *cobra.Command, args []string) error {

			key := stake.ValidatorsKey
			ctx := context.NewCoreContextFromViper()
			resKVs, err := ctx.QuerySubspace(cdc, key, storeName)
			if err != nil {
				return err
			}

			// parse out the validators
			var validators []stake.Validator
			for _, KV := range resKVs {
				var validator stake.Validator
				cdc.MustUnmarshalBinary(KV.Value, &validator)
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
			return nil

			// TODO output with proofs / machine parseable etc.
		},
	}
	return cmd
}

// get the command to query a single delegation
func GetCmdQueryDelegation(storeName string, cdc *wire.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delegation",
		Short: "Query a delegation based on address and validator address",
		RunE: func(cmd *cobra.Command, args []string) error {

			valAddr, err := sdk.GetAccAddressBech32(viper.GetString(FlagAddressValidator))
			if err != nil {
				return err
			}

			delAddr, err := sdk.GetValAddressHex(viper.GetString(FlagAddressDelegator))
			if err != nil {
				return err
			}

			key := stake.GetDelegationKey(delAddr, valAddr, cdc)
			ctx := context.NewCoreContextFromViper()
			res, err := ctx.QueryStore(key, storeName)
			if err != nil {
				return err
			}

			// parse out the delegation
			delegation := new(stake.Delegation)

			switch viper.Get(cli.OutputFlag) {
			case "text":
				resp, err := delegation.HumanReadableString()
				if err != nil {
					return err
				}
				fmt.Println(resp)
			case "json":
				cdc.MustUnmarshalBinary(res, delegation)
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

// get the command to query all the delegations made from one delegator
func GetCmdQueryDelegations(storeName string, cdc *wire.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delegations [delegator-addr]",
		Short: "Query all delegations made from one delegator",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {

			delegatorAddr, err := sdk.GetAccAddressBech32(args[0])
			if err != nil {
				return err
			}
			key := stake.GetDelegationsKey(delegatorAddr, cdc)
			ctx := context.NewCoreContextFromViper()
			resKVs, err := ctx.QuerySubspace(cdc, key, storeName)
			if err != nil {
				return err
			}

			// parse out the validators
			var delegations []stake.Delegation
			for _, KV := range resKVs {
				var delegation stake.Delegation
				cdc.MustUnmarshalBinary(KV.Value, &delegation)
				delegations = append(delegations, delegation)
			}

			output, err := wire.MarshalJSONIndent(cdc, delegations)
			if err != nil {
				return err
			}
			fmt.Println(string(output))
			return nil

			// TODO output with proofs / machine parseable etc.
		},
	}
	return cmd
}

// get the command to query a single unbonding-delegation record
func GetCmdQueryUnbondingDelegation(storeName string, cdc *wire.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unbonding-delegation",
		Short: "Query an unbonding-delegation record based on delegator and validator address",
		RunE: func(cmd *cobra.Command, args []string) error {

			valAddr, err := sdk.GetAccAddressBech32(viper.GetString(FlagAddressValidator))
			if err != nil {
				return err
			}

			delAddr, err := sdk.GetValAddressHex(viper.GetString(FlagAddressDelegator))
			if err != nil {
				return err
			}

			key := stake.GetUBDKey(delAddr, valAddr, cdc)
			ctx := context.NewCoreContextFromViper()
			res, err := ctx.QueryStore(key, storeName)
			if err != nil {
				return err
			}

			// parse out the unbonding delegation
			ubd := new(stake.UnbondingDelegation)

			switch viper.Get(cli.OutputFlag) {
			case "text":
				resp, err := ubd.HumanReadableString()
				if err != nil {
					return err
				}
				fmt.Println(resp)
			case "json":
				cdc.MustUnmarshalBinary(res, ubd)
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

// get the command to query all the unbonding-delegation records for a delegator
func GetCmdQueryUnbondingDelegations(storeName string, cdc *wire.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unbonding-delegations [delegator-addr]",
		Short: "Query all unbonding-delegations records for one delegator",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {

			delegatorAddr, err := sdk.GetAccAddressBech32(args[0])
			if err != nil {
				return err
			}
			key := stake.GetUBDsKey(delegatorAddr, cdc)
			ctx := context.NewCoreContextFromViper()
			resKVs, err := ctx.QuerySubspace(cdc, key, storeName)
			if err != nil {
				return err
			}

			// parse out the validators
			var ubds []stake.UnbondingDelegation
			for _, KV := range resKVs {
				var ubd stake.UnbondingDelegation
				cdc.MustUnmarshalBinary(KV.Value, &ubd)
				ubds = append(ubds, ubd)
			}

			output, err := wire.MarshalJSONIndent(cdc, ubds)
			if err != nil {
				return err
			}
			fmt.Println(string(output))
			return nil

			// TODO output with proofs / machine parseable etc.
		},
	}
	return cmd
}

// get the command to query a single unbonding-delegation record
func GetCmdQueryRedelegation(storeName string, cdc *wire.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unbonding-delegation",
		Short: "Query an unbonding-delegation record based on delegator and validator address",
		RunE: func(cmd *cobra.Command, args []string) error {

			valSrcAddr, err := sdk.GetAccAddressBech32(viper.GetString(FlagAddressValidatorSrc))
			if err != nil {
				return err
			}
			valDstAddr, err := sdk.GetAccAddressBech32(viper.GetString(FlagAddressValidatorDst))
			if err != nil {
				return err
			}
			delAddr, err := sdk.GetValAddressHex(viper.GetString(FlagAddressDelegator))
			if err != nil {
				return err
			}

			key := stake.GetREDKey(delAddr, valSrcAddr, valDstAddr, cdc)
			ctx := context.NewCoreContextFromViper()
			res, err := ctx.QueryStore(key, storeName)
			if err != nil {
				return err
			}

			// parse out the unbonding delegation
			red := new(stake.Redelegation)

			switch viper.Get(cli.OutputFlag) {
			case "text":
				resp, err := red.HumanReadableString()
				if err != nil {
					return err
				}
				fmt.Println(resp)
			case "json":
				cdc.MustUnmarshalBinary(res, red)
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

// get the command to query all the unbonding-delegation records for a delegator
func GetCmdQueryRedelegations(storeName string, cdc *wire.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unbonding-delegations [delegator-addr]",
		Short: "Query all unbonding-delegations records for one delegator",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {

			delegatorAddr, err := sdk.GetAccAddressBech32(args[0])
			if err != nil {
				return err
			}
			key := stake.GetREDsKey(delegatorAddr, cdc)
			ctx := context.NewCoreContextFromViper()
			resKVs, err := ctx.QuerySubspace(cdc, key, storeName)
			if err != nil {
				return err
			}

			// parse out the validators
			var reds []stake.Redelegation
			for _, KV := range resKVs {
				var red stake.Redelegation
				cdc.MustUnmarshalBinary(KV.Value, &red)
				reds = append(reds, red)
			}

			output, err := wire.MarshalJSONIndent(cdc, reds)
			if err != nil {
				return err
			}
			fmt.Println(string(output))
			return nil

			// TODO output with proofs / machine parseable etc.
		},
	}
	return cmd
}
