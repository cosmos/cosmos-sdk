package cli

import (
	"encoding/hex"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	crypto "github.com/tendermint/go-crypto"

	"github.com/cosmos/cosmos-sdk/client/context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire" // XXX fix
	"github.com/cosmos/cosmos-sdk/x/stake"
)

//// create command to query for all validators
//func GetCmdQueryValidators(storeName string, cdc *wire.Codec) *cobra.Command {
//cmd := &cobra.Command{
//Use:   "validators",
//Short: "Query for the set of validator-validators pubkeys",
//RunE: func(cmd *cobra.Command, args []string) error {

//key := stake.ValidatorsKey

//ctx := context.NewCoreContextFromViper()
//res, err := ctx.Query(key, storeName)
//if err != nil {
//return err
//}

//// parse out the validators
//validators := new(stake.Validators)
//err = cdc.UnmarshalBinary(res, validators)
//if err != nil {
//return err
//}
//output, err := wire.MarshalJSONIndent(cdc, validators)
//if err != nil {
//return err
//}
//fmt.Println(string(output))
//return nil

//// TODO output with proofs / machine parseable etc.
//},
//}

//cmd.Flags().AddFlagSet(fsDelegator)
//return cmd
//}

// get the command to query a validator
func GetCmdQueryValidator(storeName string, cdc *wire.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validator [validator-addr]",
		Short: "Query a validator-validator account",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {

			addr, err := sdk.GetAddress(args[0])
			if err != nil {
				return err
			}
			key := stake.GetValidatorKey(addr)
			ctx := context.NewCoreContextFromViper()

			res, err := ctx.Query(key, storeName)
			if err != nil {
				return err
			}

			// parse out the validator
			validator := new(stake.Validator)
			err = cdc.UnmarshalBinary(res, validator)
			if err != nil {
				return err
			}
			output, err := wire.MarshalJSONIndent(cdc, validator)
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

// get the command to query a single delegator bond
func GetCmdQueryDelegation(storeName string, cdc *wire.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delegator-bond",
		Short: "Query a delegators bond based on address and validator pubkey",
		RunE: func(cmd *cobra.Command, args []string) error {

			addr, err := sdk.GetAddress(viper.GetString(FlagAddressValidator))
			if err != nil {
				return err
			}

			bz, err := hex.DecodeString(viper.GetString(FlagAddressDelegator))
			if err != nil {
				return err
			}
			delegator := crypto.Address(bz)

			key := stake.GetDelegationKey(delegator, addr, cdc)

			ctx := context.NewCoreContextFromViper()

			res, err := ctx.Query(key, storeName)
			if err != nil {
				return err
			}

			// parse out the bond
			bond := new(stake.Delegation)
			err = cdc.UnmarshalBinary(res, bond)
			if err != nil {
				return err
			}
			output, err := wire.MarshalJSONIndent(cdc, bond)
			if err != nil {
				return err
			}
			fmt.Println(string(output))
			return nil

			// TODO output with proofs / machine parseable etc.
		},
	}

	cmd.Flags().AddFlagSet(fsValidator)
	cmd.Flags().AddFlagSet(fsDelegator)
	return cmd
}

//// get the command to query all the validators bonded to a delegator
//func GetCmdQueryDelegations(storeName string, cdc *wire.Codec) *cobra.Command {
//cmd := &cobra.Command{
//Use:   "delegator-validators",
//Short: "Query all delegators bond's validator-addresses based on delegator-address",
//RunE: func(cmd *cobra.Command, args []string) error {

//bz, err := hex.DecodeString(viper.GetString(FlagAddressDelegator))
//if err != nil {
//return err
//}
//delegator := crypto.Address(bz)

//key := stake.GetDelegationsKey(delegator, cdc)

//ctx := context.NewCoreContextFromViper()

//res, err := ctx.Query(key, storeName)
//if err != nil {
//return err
//}

//// parse out the validators list
//var validators []crypto.PubKey
//err = cdc.UnmarshalBinary(res, validators)
//if err != nil {
//return err
//}
//output, err := wire.MarshalJSONIndent(cdc, validators)
//if err != nil {
//return err
//}
//fmt.Println(string(output))
//return nil

//// TODO output with proofs / machine parseable etc.
//},
//}
//cmd.Flags().AddFlagSet(fsDelegator)
//return cmd
//}
