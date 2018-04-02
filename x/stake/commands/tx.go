package commands

import (
	"encoding/hex"
	"fmt"

	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"

	crypto "github.com/tendermint/go-crypto"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/cosmos/cosmos-sdk/x/stake"
)

// nolint
const (
	FlagAddressDelegator = "addressD"
	FlagAddressCandidate = "addressC"
	FlagPubKey           = "pubkey"
	FlagAmount           = "amount"
	FlagShares           = "shares"

	FlagMoniker  = "moniker"
	FlagIdentity = "keybase-sig"
	FlagWebsite  = "website"
	FlagDetails  = "details"
)

// common flagsets to add to various functions
var (
	fsPk        = flag.NewFlagSet("", flag.ContinueOnError)
	fsAmount    = flag.NewFlagSet("", flag.ContinueOnError)
	fsShares    = flag.NewFlagSet("", flag.ContinueOnError)
	fsCandidate = flag.NewFlagSet("", flag.ContinueOnError)
	fsDelegator = flag.NewFlagSet("", flag.ContinueOnError)
)

func init() {
	fsPk.String(FlagPubKey, "", "PubKey of the validator-candidate")
	fsAmount.String(FlagAmount, "1fermion", "Amount of coins to bond")
	fsShares.String(FlagShares, "", "Amount of shares to unbond, either in decimal or keyword MAX (ex. 1.23456789, 99, MAX)")
	fsCandidate.String(FlagMoniker, "", "validator-candidate name")
	fsCandidate.String(FlagIdentity, "", "optional keybase signature")
	fsCandidate.String(FlagWebsite, "", "optional website")
	fsCandidate.String(FlagAddressCandidate, "", "hex address of the validator/candidate")
	fsDelegator.String(FlagAddressDelegator, "", "hex address of the delegator")
}

//TODO refactor to common functionality
func getNamePassword() (name, passphrase string, err error) {
	name = viper.GetString(client.FlagName)
	buf := client.BufferStdin()
	prompt := fmt.Sprintf("Password to sign with '%s':", name)
	passphrase, err = client.GetPassword(prompt, buf)
	return
}

//_________________________________________________________________________________________

// create declare candidacy command
func GetCmdDeclareCandidacy(cdc *wire.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "declare-candidacy",
		Short: "create new validator-candidate account and delegate some coins to it",
		RunE: func(cmd *cobra.Command, args []string) error {
			amount, err := sdk.ParseCoin(viper.GetString(FlagAmount))
			if err != nil {
				return err
			}
			candidateAddr, err := sdk.GetAddress(viper.GetString(FlagAddressCandidate))
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
			msg := stake.NewMsgDeclareCandidacy(candidateAddr, pk, amount, description)

			// build and sign the transaction, then broadcast to Tendermint
			ctx := context.NewCoreContextFromViper()
			res, err := ctx.SignBuildBroadcast(ctx.FromAddressName, msg, cdc)
			if err != nil {
				return err
			}

			fmt.Printf("Committed at block %d. Hash: %s\n", res.Height, res.Hash.String())
			return nil
		},
	}

	cmd.Flags().AddFlagSet(fsPk)
	cmd.Flags().AddFlagSet(fsAmount)
	cmd.Flags().AddFlagSet(fsCandidate)
	return cmd
}

// create edit candidacy command
func GetCmdEditCandidacy(cdc *wire.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "edit-candidacy",
		Short: "edit and existing validator-candidate account",
		RunE: func(cmd *cobra.Command, args []string) error {

			candidateAddr, err := sdk.GetAddress(viper.GetString(FlagAddressCandidate))
			if err != nil {
				return err
			}
			description := stake.Description{
				Moniker:  viper.GetString(FlagMoniker),
				Identity: viper.GetString(FlagIdentity),
				Website:  viper.GetString(FlagWebsite),
				Details:  viper.GetString(FlagDetails),
			}
			msg := stake.NewMsgEditCandidacy(candidateAddr, description)

			// build and sign the transaction, then broadcast to Tendermint
			ctx := context.NewCoreContextFromViper()
			res, err := ctx.SignBuildBroadcast(ctx.FromAddressName, msg, cdc)
			if err != nil {
				return err
			}

			fmt.Printf("Committed at block %d. Hash: %s\n", res.Height, res.Hash.String())
			return nil
		},
	}

	cmd.Flags().AddFlagSet(fsPk)
	cmd.Flags().AddFlagSet(fsCandidate)
	return cmd
}

// create edit candidacy command
func GetCmdDelegate(cdc *wire.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delegate",
		Short: "delegate coins to an existing validator/candidate",
		RunE: func(cmd *cobra.Command, args []string) error {
			amount, err := sdk.ParseCoin(viper.GetString(FlagAmount))
			if err != nil {
				return err
			}

			delegatorAddr, err := sdk.GetAddress(viper.GetString(FlagAddressDelegator))
			candidateAddr, err := sdk.GetAddress(viper.GetString(FlagAddressCandidate))
			if err != nil {
				return err
			}

			msg := stake.NewMsgDelegate(delegatorAddr, candidateAddr, amount)

			// build and sign the transaction, then broadcast to Tendermint
			ctx := context.NewCoreContextFromViper()
			res, err := ctx.SignBuildBroadcast(ctx.FromAddressName, msg, cdc)
			if err != nil {
				return err
			}

			fmt.Printf("Committed at block %d. Hash: %s\n", res.Height, res.Hash.String())
			return nil
		},
	}

	cmd.Flags().AddFlagSet(fsPk)
	cmd.Flags().AddFlagSet(fsAmount)
	cmd.Flags().AddFlagSet(fsDelegator)
	return cmd
}

// create edit candidacy command
func GetCmdUnbond(cdc *wire.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unbond",
		Short: "unbond coins from a validator/candidate",
		RunE: func(cmd *cobra.Command, args []string) error {

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

			delegatorAddr, err := sdk.GetAddress(viper.GetString(FlagAddressDelegator))
			candidateAddr, err := sdk.GetAddress(viper.GetString(FlagAddressCandidate))
			if err != nil {
				return err
			}

			msg := stake.NewMsgUnbond(delegatorAddr, candidateAddr, sharesStr)

			// build and sign the transaction, then broadcast to Tendermint
			ctx := context.NewCoreContextFromViper()
			res, err := ctx.SignBuildBroadcast(ctx.FromAddressName, msg, cdc)
			if err != nil {
				return err
			}

			fmt.Printf("Committed at block %d. Hash: %s\n", res.Height, res.Hash.String())
			return nil
		},
	}

	cmd.Flags().AddFlagSet(fsPk)
	cmd.Flags().AddFlagSet(fsShares)
	cmd.Flags().AddFlagSet(fsDelegator)
	return cmd
}

//______________________________________________________________________________________

// create the pubkey from a pubkey string
// TODO move to a better reusable place
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
