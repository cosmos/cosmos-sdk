package commands

import (
	"encoding/hex"
	"fmt"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"

	crypto "github.com/tendermint/go-crypto"

	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/stake"
)

// nolint
const (
	FlagPubKey = "pubkey"
	FlagAmount = "amount"
	FlagShares = "shares"

	FlagMoniker  = "moniker"
	FlagIdentity = "keybase-sig"
	FlagWebsite  = "website"
	FlagDetails  = "details"
)

// nolint
var (
	CmdDeclareCandidacy = &cobra.Command{
		Use:   "declare-candidacy",
		Short: "create new validator-candidate account and delegate some coins to it",
		RunE:  cmdDeclareCandidacy,
	}
	CmdEditCandidacy = &cobra.Command{
		Use:   "edit-candidacy",
		Short: "edit and existing validator-candidate account",
		RunE:  cmdEditCandidacy,
	}
	CmdDelegate = &cobra.Command{
		Use:   "delegate",
		Short: "delegate coins to an existing validator/candidate",
		RunE:  cmdDelegate,
	}
	CmdUnbond = &cobra.Command{
		Use:   "unbond",
		Short: "unbond coins from a validator/candidate",
		RunE:  cmdUnbond,
	}
)

func init() {

	// define the flags
	fsPk := flag.NewFlagSet("", flag.ContinueOnError)
	fsPk.String(FlagPubKey, "", "PubKey of the validator-candidate")

	fsAmount := flag.NewFlagSet("", flag.ContinueOnError)
	fsAmount.String(FlagAmount, "1fermion", "Amount of coins to bond")

	fsShares := flag.NewFlagSet("", flag.ContinueOnError)
	fsShares.String(FlagShares, "", "Amount of shares to unbond, either in decimal or keyword MAX (ex. 1.23456789, 99, MAX)")

	fsCandidate := flag.NewFlagSet("", flag.ContinueOnError)
	fsCandidate.String(FlagMoniker, "", "validator-candidate name")
	fsCandidate.String(FlagIdentity, "", "optional keybase signature")
	fsCandidate.String(FlagWebsite, "", "optional website")
	fsCandidate.String(FlagDetails, "", "optional detailed description space")

	// add the flags
	CmdDelegate.Flags().AddFlagSet(fsPk)
	CmdDelegate.Flags().AddFlagSet(fsAmount)

	CmdUnbond.Flags().AddFlagSet(fsPk)
	CmdUnbond.Flags().AddFlagSet(fsShares)

	CmdDeclareCandidacy.Flags().AddFlagSet(fsPk)
	CmdDeclareCandidacy.Flags().AddFlagSet(fsAmount)
	CmdDeclareCandidacy.Flags().AddFlagSet(fsCandidate)

	CmdEditCandidacy.Flags().AddFlagSet(fsPk)
	CmdEditCandidacy.Flags().AddFlagSet(fsCandidate)
}

func cmdDeclareCandidacy(cmd *cobra.Command, args []string) error {
	amount, err := sdk.ParseCoin(viper.GetString(FlagAmount))
	if err != nil {
		return err
	}

	pk, err := GetPubKey(viper.GetString(FlagPubKey))
	if err != nil {
		return err
	}

	if viper.GetString(FlagMoniker) == "" {
		return fmt.Errorf("please enter a moniker for the validator-candidate using --moniker")
	}

	description := stake.Description{
		Moniker:  viper.GetString(FlagMoniker),
		Identity: viper.GetString(FlagIdentity),
		Website:  viper.GetString(FlagWebsite),
		Details:  viper.GetString(FlagDetails),
	}

	tx := stake.NewTxDeclareCandidacy(amount, pk, description)
	return doTx(tx)
}

func cmdEditCandidacy(cmd *cobra.Command, args []string) error {

	pk, err := GetPubKey(viper.GetString(FlagPubKey))
	if err != nil {
		return err
	}

	description := stake.Description{
		Moniker:  viper.GetString(FlagMoniker),
		Identity: viper.GetString(FlagIdentity),
		Website:  viper.GetString(FlagWebsite),
		Details:  viper.GetString(FlagDetails),
	}

	tx := stake.NewTxEditCandidacy(pk, description)
	return doTx(tx)
}

func cmdDelegate(cmd *cobra.Command, args []string) error {
	amount, err := sdk.ParseCoin(viper.GetString(FlagAmount))
	if err != nil {
		return err
	}

	pk, err := GetPubKey(viper.GetString(FlagPubKey))
	if err != nil {
		return err
	}

	tx := stake.NewTxDelegate(amount, pk)
	return doTx(tx)
}

func cmdUnbond(cmd *cobra.Command, args []string) error {

	// TODO once go-wire refactored the shares can be broadcast as a Rat instead of a string

	// check the shares before broadcasting
	sharesStr := viper.GetString(FlagShares)
	var shares sdk.Rat
	if sharesStr != "MAX" {
		var err error
		shares, err = sdk.NewRatFromDecimal(sharesStr)
		if err != nil {
			return err
		}
		if !shares.GT(sdk.ZeroRat) {
			return fmt.Errorf("shares must be positive integer or decimal (ex. 123, 1.23456789)")
		}
	}

	pk, err := GetPubKey(viper.GetString(FlagPubKey))
	if err != nil {
		return err
	}

	tx := stake.NewTxUnbond(sharesStr, pk)
	return doTx(tx)
}

// GetPubKey - create the pubkey from a pubkey string
func GetPubKey(pubKeyStr string) (pk crypto.PubKey, err error) {

	if len(pubKeyStr) == 0 {
		err = fmt.Errorf("must use --pubkey flag")
		return
	}
	if len(pubKeyStr) != 64 { //if len(pkBytes) != 32 {
		err = fmt.Errorf("pubkey must be Ed25519 hex encoded string which is 64 characters long")
		return
	}
	var pkBytes []byte
	pkBytes, err = hex.DecodeString(pubKeyStr)
	if err != nil {
		return
	}
	var pkEd crypto.PubKeyEd25519
	copy(pkEd[:], pkBytes[:])
	pk = pkEd.Wrap()
	return
}

//--------------------------------------------------------------------
// XXX consolidate to client

func doTx(tx []byte) {

	uri := viper.GetString(client.FlagNode)
	if uri == "" {
		return errors.New("Must define which node to query with --node")
	}
	node := client.GetNode(uri)

	result, err := node.BroadcastTxCommit(tx)
	if err != nil {
		return err
	}

	if result.CheckTx.Code != uint32(0) {
		fmt.Printf("CheckTx failed: (%d) %s\n",
			result.CheckTx.Code,
			result.CheckTx.Log)
	}
	if result.DeliverTx.Code != uint32(0) {
		fmt.Printf("DeliverTx failed: (%d) %s\n",
			result.DeliverTx.Code,
			result.DeliverTx.Log)
	}

	fmt.Printf("Committed at block %d. Hash: %s\n", result.Height, result.Hash.String())
	return nil
}
