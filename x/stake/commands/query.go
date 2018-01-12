package commands

import (
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"

	crypto "github.com/tendermint/go-crypto"

	"github.com/cosmos/cosmos-sdk/client/commands"
	"github.com/cosmos/cosmos-sdk/client/commands/query"
	"github.com/cosmos/cosmos-sdk/modules/coin"
	"github.com/cosmos/cosmos-sdk/stack"
	"github.com/cosmos/gaia/modules/stake"
)

//nolint
var (
	CmdQueryCandidates = &cobra.Command{
		Use:   "candidates",
		Short: "Query for the set of validator-candidates pubkeys",
		RunE:  cmdQueryCandidates,
	}

	CmdQueryCandidate = &cobra.Command{
		Use:   "candidate",
		Short: "Query a validator-candidate account",
		RunE:  cmdQueryCandidate,
	}

	CmdQueryDelegatorBond = &cobra.Command{
		Use:   "delegator-bond",
		Short: "Query a delegators bond based on address and candidate pubkey",
		RunE:  cmdQueryDelegatorBond,
	}

	CmdQueryDelegatorCandidates = &cobra.Command{
		Use:   "delegator-candidates",
		RunE:  cmdQueryDelegatorCandidates,
		Short: "Query all delegators candidates' pubkeys based on address",
	}

	FlagDelegatorAddress = "delegator-address"
)

func init() {
	//Add Flags
	fsPk := flag.NewFlagSet("", flag.ContinueOnError)
	fsPk.String(FlagPubKey, "", "PubKey of the validator-candidate")
	fsAddr := flag.NewFlagSet("", flag.ContinueOnError)
	fsAddr.String(FlagDelegatorAddress, "", "Delegator Hex Address")

	CmdQueryCandidate.Flags().AddFlagSet(fsPk)
	CmdQueryDelegatorBond.Flags().AddFlagSet(fsPk)
	CmdQueryDelegatorBond.Flags().AddFlagSet(fsAddr)
	CmdQueryDelegatorCandidates.Flags().AddFlagSet(fsAddr)
}

func cmdQueryCandidates(cmd *cobra.Command, args []string) error {

	var pks []crypto.PubKey

	prove := !viper.GetBool(commands.FlagTrustNode)
	key := stack.PrefixedKey(stake.Name(), stake.CandidatesPubKeysKey)
	height, err := query.GetParsed(key, &pks, query.GetHeight(), prove)
	if err != nil {
		return err
	}

	return query.OutputProof(pks, height)
}

func cmdQueryCandidate(cmd *cobra.Command, args []string) error {

	var candidate stake.Candidate

	pk, err := GetPubKey(viper.GetString(FlagPubKey))
	if err != nil {
		return err
	}

	prove := !viper.GetBool(commands.FlagTrustNode)
	key := stack.PrefixedKey(stake.Name(), stake.GetCandidateKey(pk))
	height, err := query.GetParsed(key, &candidate, query.GetHeight(), prove)
	if err != nil {
		return err
	}

	return query.OutputProof(candidate, height)
}

func cmdQueryDelegatorBond(cmd *cobra.Command, args []string) error {

	var bond stake.DelegatorBond

	pk, err := GetPubKey(viper.GetString(FlagPubKey))
	if err != nil {
		return err
	}

	delegatorAddr := viper.GetString(FlagDelegatorAddress)
	delegator, err := commands.ParseActor(delegatorAddr)
	if err != nil {
		return err
	}
	delegator = coin.ChainAddr(delegator)

	prove := !viper.GetBool(commands.FlagTrustNode)
	key := stack.PrefixedKey(stake.Name(), stake.GetDelegatorBondKey(delegator, pk))
	height, err := query.GetParsed(key, &bond, query.GetHeight(), prove)
	if err != nil {
		return err
	}

	return query.OutputProof(bond, height)
}

func cmdQueryDelegatorCandidates(cmd *cobra.Command, args []string) error {

	delegatorAddr := viper.GetString(FlagDelegatorAddress)
	delegator, err := commands.ParseActor(delegatorAddr)
	if err != nil {
		return err
	}
	delegator = coin.ChainAddr(delegator)

	prove := !viper.GetBool(commands.FlagTrustNode)
	key := stack.PrefixedKey(stake.Name(), stake.GetDelegatorBondsKey(delegator))
	var candidates []crypto.PubKey
	height, err := query.GetParsed(key, &candidates, query.GetHeight(), prove)
	if err != nil {
		return err
	}

	return query.OutputProof(candidates, height)
}
