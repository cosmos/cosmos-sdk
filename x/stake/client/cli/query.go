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

// get the command to query a candidate
func GetCmdQueryCandidate(storeName string, cdc *wire.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "candidate",
		Short: "Query a validator-candidate account",
		RunE: func(cmd *cobra.Command, args []string) error {

			addr, err := sdk.GetAddress(viper.GetString(FlagAddressCandidate))
			if err != nil {
				return err
			}

			key := stake.GetCandidateKey(addr)
			ctx := context.NewCoreContextFromViper()
			res, err := ctx.Query(key, storeName)
			if err != nil {
				return err
			}

			// parse out the candidate
			candidate := new(stake.Candidate)
			cdc.MustUnmarshalBinary(res, candidate)
			output, err := wire.MarshalJSONIndent(cdc, candidate)
			if err != nil {
				return err
			}
			fmt.Println(string(output))
			return nil

			// TODO output with proofs / machine parseable etc.
		},
	}

	cmd.Flags().AddFlagSet(fsCandidate)
	return cmd
}

// get the command to query a candidate
func GetCmdQueryCandidates(storeName string, cdc *wire.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "candidates",
		Short: "Query for all validator-candidate accounts",
		RunE: func(cmd *cobra.Command, args []string) error {

			key := stake.CandidatesKey
			ctx := context.NewCoreContextFromViper()
			resKVs, err := ctx.QuerySubspace(cdc, key, storeName)
			if err != nil {
				return err
			}

			// parse out the candidates
			var candidates []stake.Candidate
			for _, KV := range resKVs {
				var candidate stake.Candidate
				cdc.MustUnmarshalBinary(KV.Value, &candidate)
				candidates = append(candidates, candidate)
			}

			output, err := wire.MarshalJSONIndent(cdc, candidates)
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
func GetCmdQueryDelegatorBond(storeName string, cdc *wire.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delegator-bond",
		Short: "Query a delegators bond based on address and candidate pubkey",
		RunE: func(cmd *cobra.Command, args []string) error {

			addr, err := sdk.GetAddress(viper.GetString(FlagAddressCandidate))
			if err != nil {
				return err
			}

			bz, err := hex.DecodeString(viper.GetString(FlagAddressDelegator))
			if err != nil {
				return err
			}
			delegator := crypto.Address(bz)

			key := stake.GetDelegatorBondKey(delegator, addr, cdc)

			ctx := context.NewCoreContextFromViper()

			res, err := ctx.Query(key, storeName)
			if err != nil {
				return err
			}

			// parse out the bond
			bond := new(stake.DelegatorBond)
			cdc.MustUnmarshalBinary(res, bond)
			output, err := wire.MarshalJSONIndent(cdc, bond)
			if err != nil {
				return err
			}
			fmt.Println(string(output))
			return nil

			// TODO output with proofs / machine parseable etc.
		},
	}

	cmd.Flags().AddFlagSet(fsCandidate)
	cmd.Flags().AddFlagSet(fsDelegator)
	return cmd
}

//// get the command to query all the candidates bonded to a delegator
//func GetCmdQueryDelegatorBonds(storeName string, cdc *wire.Codec) *cobra.Command {
//cmd := &cobra.Command{
//Use:   "delegator-candidates",
//Short: "Query all delegators bond's candidate-addresses based on delegator-address",
//RunE: func(cmd *cobra.Command, args []string) error {

//bz, err := hex.DecodeString(viper.GetString(FlagAddressDelegator))
//if err != nil {
//return err
//}
//delegator := crypto.Address(bz)

//key := stake.GetDelegatorBondsKey(delegator, cdc)

//ctx := context.NewCoreContextFromViper()

//res, err := ctx.Query(key, storeName)
//if err != nil {
//return err
//}

//// parse out the candidates list
//var candidates []crypto.PubKey
//err = cdc.UnmarshalBinary(res, candidates)
//if err != nil {
//return err
//}
//output, err := wire.MarshalJSONIndent(cdc, candidates)
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
