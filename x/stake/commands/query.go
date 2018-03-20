package commands

import (
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"

	crypto "github.com/tendermint/go-crypto"

	"github.com/cosmos/cosmos-sdk/client/builder"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire" // XXX fix
	"github.com/cosmos/cosmos-sdk/x/stake"
)

// XXX remove dependancy
func PrefixedKey(app string, key []byte) []byte {
	prefix := append([]byte(app), byte(0))
	return append(prefix, key...)
}

//nolint
var (
	fsValAddr         = flag.NewFlagSet("", flag.ContinueOnError)
	fsDelAddr         = flag.NewFlagSet("", flag.ContinueOnError)
	FlagValidatorAddr = "address"
	FlagDelegatorAddr = "delegator-address"
)

func init() {
	//Add Flags
	fsValAddr.String(FlagValidatorAddr, "", "Address of the validator/candidate")
	fsDelAddr.String(FlagDelegatorAddr, "", "Delegator hex address")

}

// create command to query for all candidates
func GetCmdQueryCandidates(cdc *wire.Codec, storeName string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "candidates",
		Short: "Query for the set of validator-candidates pubkeys",
		RunE: func(cmd *cobra.Command, args []string) error {

			key := PrefixedKey(stake.MsgType, stake.CandidatesAddrKey)

			res, err := builder.Query(key, storeName)
			if err != nil {
				return err
			}

			// parse out the candidates
			candidates := new(stake.Candidates)
			err = cdc.UnmarshalJSON(res, candidates)
			if err != nil {
				return err
			}
			output, err := json.MarshalIndent(candidates, "", "  ")
			if err != nil {
				return err
			}
			fmt.Println(string(output))
			return nil

			// TODO output with proofs / machine parseable etc.
		},
	}

	cmd.Flags().AddFlagSet(fsDelAddr)
	return cmd
}

// get the command to query a candidate
func GetCmdQueryCandidate(cdc *wire.Codec, storeName string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "candidate",
		Short: "Query a validator-candidate account",
		RunE: func(cmd *cobra.Command, args []string) error {

			addr, err := sdk.GetAddress(viper.GetString(FlagValidatorAddr))
			if err != nil {
				return err
			}

			key := PrefixedKey(stake.MsgType, stake.GetCandidateKey(addr))

			res, err := builder.Query(key, storeName)
			if err != nil {
				return err
			}

			// parse out the candidate
			candidate := new(stake.Candidate)
			err = cdc.UnmarshalBinary(res, candidate)
			if err != nil {
				return err
			}
			output, err := json.MarshalIndent(candidate, "", "  ")
			if err != nil {
				return err
			}
			fmt.Println(string(output))
			return nil

			// TODO output with proofs / machine parseable etc.
		},
	}

	cmd.Flags().AddFlagSet(fsValAddr)
	return cmd
}

// get the command to query a single delegator bond
func GetCmdQueryDelegatorBond(cdc *wire.Codec, storeName string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delegator-bond",
		Short: "Query a delegators bond based on address and candidate pubkey",
		RunE: func(cmd *cobra.Command, args []string) error {

			addr, err := sdk.GetAddress(viper.GetString(FlagValidatorAddr))
			if err != nil {
				return err
			}

			bz, err := hex.DecodeString(viper.GetString(FlagDelegatorAddr))
			if err != nil {
				return err
			}
			delegator := crypto.Address(bz)

			key := PrefixedKey(stake.MsgType, stake.GetDelegatorBondKey(delegator, addr, cdc))

			res, err := builder.Query(key, storeName)
			if err != nil {
				return err
			}

			// parse out the bond
			var bond stake.DelegatorBond
			err = cdc.UnmarshalBinary(res, bond)
			if err != nil {
				return err
			}
			output, err := json.MarshalIndent(bond, "", "  ")
			if err != nil {
				return err
			}
			fmt.Println(string(output))
			return nil

			// TODO output with proofs / machine parseable etc.
		},
	}

	cmd.Flags().AddFlagSet(fsValAddr)
	cmd.Flags().AddFlagSet(fsDelAddr)
	return cmd
}

// get the command to query all the candidates bonded to a delegator
func GetCmdQueryDelegatorBonds(cdc *wire.Codec, storeName string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delegator-candidates",
		Short: "Query all delegators candidates' pubkeys based on address",
		RunE: func(cmd *cobra.Command, args []string) error {

			bz, err := hex.DecodeString(viper.GetString(FlagDelegatorAddr))
			if err != nil {
				return err
			}
			delegator := crypto.Address(bz)

			key := PrefixedKey(stake.MsgType, stake.GetDelegatorBondsKey(delegator, cdc))

			res, err := builder.Query(key, storeName)
			if err != nil {
				return err
			}

			// parse out the candidates list
			var candidates []crypto.PubKey
			err = cdc.UnmarshalBinary(res, candidates)
			if err != nil {
				return err
			}
			output, err := json.MarshalIndent(candidates, "", "  ")
			if err != nil {
				return err
			}
			fmt.Println(string(output))
			return nil

			// TODO output with proofs / machine parseable etc.
		},
	}
	cmd.Flags().AddFlagSet(fsDelAddr)
	return cmd
}
