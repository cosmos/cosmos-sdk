package cli

import (
	flag "github.com/spf13/pflag"
)

// nolint
const (
	FlagAddressDelegator = "address-delegator"
	FlagAddressValidator = "address-validator"
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
	fsPk          = flag.NewFlagSet("", flag.ContinueOnError)
	fsAmount      = flag.NewFlagSet("", flag.ContinueOnError)
	fsShares      = flag.NewFlagSet("", flag.ContinueOnError)
	fsDescription = flag.NewFlagSet("", flag.ContinueOnError)
	fsValidator   = flag.NewFlagSet("", flag.ContinueOnError)
	fsDelegator   = flag.NewFlagSet("", flag.ContinueOnError)
)

func init() {
	fsPk.String(FlagPubKey, "", "Go-Amino encoded hex PubKey of the validator. For Ed25519 the go-amino prepend hex is 1624de6220")
	fsAmount.String(FlagAmount, "1steak", "Amount of coins to bond")
	fsShares.String(FlagShares, "", "Amount of shares to unbond, either in decimal or keyword MAX (ex. 1.23456789, 99, MAX)")
	fsDescription.String(FlagMoniker, "", "validator name")
	fsDescription.String(FlagIdentity, "", "optional keybase signature")
	fsDescription.String(FlagWebsite, "", "optional website")
	fsDescription.String(FlagDetails, "", "optional details")
	fsValidator.String(FlagAddressValidator, "", "hex address of the validator")
	fsDelegator.String(FlagAddressDelegator, "", "hex address of the delegator")
}
