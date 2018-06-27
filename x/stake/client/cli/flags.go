package cli

import (
	flag "github.com/spf13/pflag"
)

// nolint
const (
	FlagAddressDelegator    = "address-delegator"
	FlagAddressValidator    = "address-validator"
	FlagAddressValidatorSrc = "addr-validator-source"
	FlagAddressValidatorDst = "addr-validator-dest"
	FlagPubKey              = "pubkey"
	FlagAmount              = "amount"
	FlagSharesAmount        = "shares-amount"
	FlagSharesPercent       = "shares-percent"

	FlagMoniker  = "moniker"
	FlagIdentity = "keybase-sig"
	FlagWebsite  = "website"
	FlagDetails  = "details"
)

// common flagsets to add to various functions
var (
	fsPk           = flag.NewFlagSet("", flag.ContinueOnError)
	fsAmount       = flag.NewFlagSet("", flag.ContinueOnError)
	fsShares       = flag.NewFlagSet("", flag.ContinueOnError)
	fsDescription  = flag.NewFlagSet("", flag.ContinueOnError)
	fsValidator    = flag.NewFlagSet("", flag.ContinueOnError)
	fsDelegator    = flag.NewFlagSet("", flag.ContinueOnError)
	fsRedelegation = flag.NewFlagSet("", flag.ContinueOnError)
)

func init() {
	fsPk.String(FlagPubKey, "", "Go-Amino encoded hex PubKey of the validator. For Ed25519 the go-amino prepend hex is 1624de6220")
	fsAmount.String(FlagAmount, "1steak", "Amount of coins to bond")
	fsShares.String(FlagSharesAmount, "", "Amount of source-shares to either unbond or redelegate as a positive integer or decimal")
	fsShares.String(FlagSharesPercent, "", "Percent of source-shares to either unbond or redelegate as a positive integer or decimal >0 and <=1")
	fsDescription.String(FlagMoniker, "[do-not-modify]", "validator name")
	fsDescription.String(FlagIdentity, "[do-not-modify]", "optional keybase signature")
	fsDescription.String(FlagWebsite, "[do-not-modify]", "optional website")
	fsDescription.String(FlagDetails, "[do-not-modify]", "optional details")
	fsValidator.String(FlagAddressValidator, "", "hex address of the validator")
	fsDelegator.String(FlagAddressDelegator, "", "hex address of the delegator")
	fsRedelegation.String(FlagAddressValidatorSrc, "", "hex address of the source validator")
	fsRedelegation.String(FlagAddressValidatorDst, "", "hex address of the destination validator")
}
