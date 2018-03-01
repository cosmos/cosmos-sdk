package commands

import (
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"

	crypto "github.com/tendermint/go-crypto"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/examples/basecoin/app"
	coin "github.com/cosmos/cosmos-sdk/x/bank" // XXX fix
	"github.com/cosmos/cosmos-sdk/x/stake"
)

// XXX remove dependancy
func PrefixedKey(app string, key []byte) []byte {
	prefix := append([]byte(app), byte(0))
	return append(prefix, key...)
}

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

// XXX move to common directory in client helpers
func makeQuery(key, storeName string) (res []byte, err error) {

	path := fmt.Sprintf("/%s/key", a.storeName)

	uri := viper.GetString(client.FlagNode)
	if uri == "" {
		return res, errors.New("Must define which node to query with --node")
	}
	node := client.GetNode(uri)

	opts := rpcclient.ABCIQueryOptions{
		Height:  viper.GetInt64(client.FlagHeight),
		Trusted: viper.GetBool(client.FlagTrustNode),
	}
	result, err := node.ABCIQueryWithOptions(path, key, opts)
	if err != nil {
		return res, err
	}
	resp := result.Response
	if resp.Code != uint32(0) {
		return res, errors.Errorf("Query failed: (%d) %s", resp.Code, resp.Log)
	}
	return resp.val, nil
}

func cmdQueryCandidates(cmd *cobra.Command, args []string) error {

	var pks []crypto.PubKey

	prove := !viper.GetBool(client.FlagTrustNode)
	key := PrefixedKey(stake.Name(), stake.CandidatesPubKeysKey)

	res, err := makeQuery(key, "gaia-store-name") // XXX move gaia store name out of here
	if err != nil {
		return err
	}

	// parse out the candidates
	candidates := new(stake.Candidates)
	cdc := app.MakeTxCodec() // XXX create custom Tx for Staking Module
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
}

func cmdQueryCandidate(cmd *cobra.Command, args []string) error {

	var candidate stake.Candidate

	pk, err := GetPubKey(viper.GetString(FlagPubKey))
	if err != nil {
		return err
	}

	prove := !viper.GetBool(client.FlagTrustNode)
	key := PrefixedKey(stake.Name(), stake.GetCandidateKey(pk))

	// parse out the candidate
	candidate := new(stake.Candidate)
	cdc := app.MakeTxCodec() // XXX create custom Tx for Staking Module
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
}

func cmdQueryDelegatorBond(cmd *cobra.Command, args []string) error {

	pk, err := GetPubKey(viper.GetString(FlagPubKey))
	if err != nil {
		return err
	}

	bz, err := hex.DecodeString(viper.GetString(FlagDelegatorAddress))
	if err != nil {
		return err
	}
	delegator := crypto.Address(bz)
	delegator = coin.ChainAddr(delegator)

	prove := !viper.GetBool(client.FlagTrustNode)
	key := PrefixedKey(stake.Name(), stake.GetDelegatorBondKey(delegator, pk))

	// parse out the bond
	var bond stake.DelegatorBond
	cdc := app.MakeTxCodec() // XXX create custom Tx for Staking Module
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
}

func cmdQueryDelegatorCandidates(cmd *cobra.Command, args []string) error {

	bz, err := hex.DecodeString(viper.GetString(FlagDelegatorAddress))
	if err != nil {
		return err
	}
	delegator := crypto.Address(bz)
	delegator = coin.ChainAddr(delegator)

	prove := !viper.GetBool(client.FlagTrustNode)
	key := PrefixedKey(stake.Name(), stake.GetDelegatorBondsKey(delegator))

	// parse out the candidates list
	var candidates []crypto.PubKey
	cdc := app.MakeTxCodec() // XXX create custom Tx for Staking Module
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
}
