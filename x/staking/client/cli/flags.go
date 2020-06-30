package cli

import (
	flag "github.com/spf13/pflag"

	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

const (
	FlagAddressValidator    = "validator"
	FlagAddressValidatorSrc = "addr-validator-source"
	FlagAddressValidatorDst = "addr-validator-dest"
	FlagPubKey              = "pubkey"
	FlagAmount              = "amount"
	FlagSharesAmount        = "shares-amount"
	FlagSharesFraction      = "shares-fraction"

	FlagMoniker         = "moniker"
	FlagIdentity        = "identity"
	FlagWebsite         = "website"
	FlagSecurityContact = "security-contact"
	FlagDetails         = "details"

	FlagCommissionRate          = "commission-rate"
	FlagCommissionMaxRate       = "commission-max-rate"
	FlagCommissionMaxChangeRate = "commission-max-change-rate"

	FlagMinSelfDelegation = "min-self-delegation"

	FlagGenesisFormat = "genesis-format"
	FlagNodeID        = "node-id"
	FlagIP            = "ip"
)

// Define common flagsets to add to various functions

func FsPk() *flag.FlagSet {
	f := flag.NewFlagSet("", flag.ContinueOnError)
	f.String(FlagPubKey, "", "The Bech32 encoded PubKey of the validator")
	return f
}

func FsAmount() *flag.FlagSet {
	f := flag.NewFlagSet("", flag.ContinueOnError)
	f.String(FlagAmount, "", "Amount of coins to bond")
	return f
}

func fsShares() *flag.FlagSet {
	f := flag.NewFlagSet("", flag.ContinueOnError)
	f.String(FlagSharesAmount, "", "Amount of source-shares to either unbond or redelegate as a positive integer or decimal")
	f.String(FlagSharesFraction, "", "Fraction of source-shares to either unbond or redelegate as a positive integer or decimal >0 and <=1")
	return f
}

func fsDescriptionCreate() *flag.FlagSet {
	f := flag.NewFlagSet("", flag.ContinueOnError)
	f.String(FlagMoniker, "", "The validator's name")
	f.String(FlagIdentity, "", "The optional identity signature (ex. UPort or Keybase)")
	f.String(FlagWebsite, "", "The validator's (optional) website")
	f.String(FlagSecurityContact, "", "The validator's (optional) security contact email")
	f.String(FlagDetails, "", "The validator's (optional) details")
	return f
}

func fsCommissionUpdate() *flag.FlagSet {
	f := flag.NewFlagSet("", flag.ContinueOnError)
	f.String(FlagCommissionRate, "", "The new commission rate percentage")
	return f
}

func FsCommissionCreate() *flag.FlagSet {
	f := flag.NewFlagSet("", flag.ContinueOnError)
	f.String(FlagCommissionRate, "", "The initial commission rate percentage")
	f.String(FlagCommissionMaxRate, "", "The maximum commission rate percentage")
	f.String(FlagCommissionMaxChangeRate, "", "The maximum commission change rate percentage (per day)")
	return f
}

func FsMinSelfDelegation() *flag.FlagSet {
	f := flag.NewFlagSet("", flag.ContinueOnError)
	f.String(FlagMinSelfDelegation, "", "The minimum self delegation required on the validator")
	return f
}

func fsDescriptionEdit() *flag.FlagSet {
	f := flag.NewFlagSet("", flag.ContinueOnError)
	f.String(FlagMoniker, types.DoNotModifyDesc, "The validator's name")
	f.String(FlagIdentity, types.DoNotModifyDesc, "The (optional) identity signature (ex. UPort or Keybase)")
	f.String(FlagWebsite, types.DoNotModifyDesc, "The validator's (optional) website")
	f.String(FlagSecurityContact, types.DoNotModifyDesc, "The validator's (optional) security contact email")
	f.String(FlagDetails, types.DoNotModifyDesc, "The validator's (optional) details")
	return f
}

func fsValidator() *flag.FlagSet {
	f := flag.NewFlagSet("", flag.ContinueOnError)
	f.String(FlagAddressValidator, "", "The Bech32 address of the validator")
	return f
}

func fsRedelegation() *flag.FlagSet {
	f := flag.NewFlagSet("", flag.ContinueOnError)
	f.String(FlagAddressValidatorSrc, "", "The Bech32 address of the source validator")
	f.String(FlagAddressValidatorDst, "", "The Bech32 address of the destination validator")
	return f
}
