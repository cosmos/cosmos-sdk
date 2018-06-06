package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tendermint/tmlibs/cli"

	"github.com/cosmos/cosmos-sdk/client/context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire" // XXX fix
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
			res, err := ctx.Query(key, storeName)
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

// get the command to query a single delegation bond
func GetCmdQueryDelegation(storeName string, cdc *wire.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delegation",
		Short: "Query a delegations bond based on address and validator address",
		RunE: func(cmd *cobra.Command, args []string) error {

			addr, err := sdk.GetAccAddressBech32(viper.GetString(FlagAddressValidator))
			if err != nil {
				return err
			}

			delAddr, err := sdk.GetValAddressHex(viper.GetString(FlagAddressDelegator))
			if err != nil {
				return err
			}

			key := stake.GetDelegationKey(delAddr, addr, cdc)
			ctx := context.NewCoreContextFromViper()
			res, err := ctx.Query(key, storeName)
			if err != nil {
				return err
			}

			// parse out the bond
			bond := new(stake.Delegation)

			switch viper.Get(cli.OutputFlag) {
			case "text":
				resp, err := bond.HumanReadableString()
				if err != nil {
					return err
				}
				fmt.Println(resp)
			case "json":
				cdc.MustUnmarshalBinary(res, bond)
				output, err := wire.MarshalJSONIndent(cdc, bond)
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

// get the command to query all the validators bonded to a delegation
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
