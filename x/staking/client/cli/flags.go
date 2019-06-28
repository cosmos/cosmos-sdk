package cli

import (
	flag "github.com/spf13/pflag"

	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

// nolint
const (
	FlagAddressValidator    = "validator"
	FlagAddressValidatorSrc = "addr-validator-source"
	FlagAddressValidatorDst = "addr-validator-dest"
	FlagPubKey              = "pubkey"
	FlagAmount              = "amount"
	FlagSharesAmount        = "shares-amount"
	FlagSharesFraction      = "shares-fraction"

	FlagMoniker  = "moniker"
	FlagIdentity = "identity"
	FlagWebsite  = "website"
	FlagDetails  = "details"

	FlagCommissionRate          = "commission-rate"
	FlagCommissionMaxRate       = "commission-max-rate"
	FlagCommissionMaxChangeRate = "commission-max-change-rate"

	FlagMinSelfDelegation = "min-self-delegation"

	FlagGenesisFormat = "genesis-format"
	FlagNodeID        = "node-id"
	FlagIP            = "ip"
)

// common flagsets to add to various functions
var (
	FsPk                = flag.NewFlagSet("", flag.ContinueOnError)
	FsAmount            = flag.NewFlagSet("", flag.ContinueOnError)
	fsShares            = flag.NewFlagSet("", flag.ContinueOnError)
	fsDescriptionCreate = flag.NewFlagSet("", flag.ContinueOnError)
	FsCommissionCreate  = flag.NewFlagSet("", flag.ContinueOnError)
	fsCommissionUpdate  = flag.NewFlagSet("", flag.ContinueOnError)
	FsMinSelfDelegation = flag.NewFlagSet("", flag.ContinueOnError)
	fsDescriptionEdit   = flag.NewFlagSet("", flag.ContinueOnError)
	fsValidator         = flag.NewFlagSet("", flag.ContinueOnError)
	fsRedelegation      = flag.NewFlagSet("", flag.ContinueOnError)
)

func init() {
	FsPk.String(FlagPubKey, "", "The Bech32 encoded PubKey of the validator")
	FsAmount.String(FlagAmount, "", "Amount of coins to bond")
	fsShares.String(FlagSharesAmount, "", "Amount of source-shares to either unbond or redelegate as a positive integer or decimal")
	fsShares.String(FlagSharesFraction, "", "Fraction of source-shares to either unbond or redelegate as a positive integer or decimal >0 and <=1")
	fsDescriptionCreate.String(FlagMoniker, "", "The validator's name")
	fsDescriptionCreate.String(FlagIdentity, "", "The optional identity signature (ex. UPort or Keybase)")
	fsDescriptionCreate.String(FlagWebsite, "", "The validator's (optional) website")
	fsDescriptionCreate.String(FlagDetails, "", "The validator's (optional) details")
	fsCommissionUpdate.String(FlagCommissionRate, "", "The new commission rate percentage")
	FsCommissionCreate.String(FlagCommissionRate, "", "The initial commission rate percentage")
	FsCommissionCreate.String(FlagCommissionMaxRate, "", "The maximum commission rate percentage")
	FsCommissionCreate.String(FlagCommissionMaxChangeRate, "", "The maximum commission change rate percentage (per day)")
	FsMinSelfDelegation.String(FlagMinSelfDelegation, "", "The minimum self delegation required on the validator")
	fsDescriptionEdit.String(FlagMoniker, types.DoNotModifyDesc, "The validator's name")
	fsDescriptionEdit.String(FlagIdentity, types.DoNotModifyDesc, "The (optional) identity signature (ex. UPort or Keybase)")
	fsDescriptionEdit.String(FlagWebsite, types.DoNotModifyDesc, "The validator's (optional) website")
	fsDescriptionEdit.String(FlagDetails, types.DoNotModifyDesc, "The validator's (optional) details")
	fsValidator.String(FlagAddressValidator, "", "The Bech32 address of the validator")
	fsRedelegation.String(FlagAddressValidatorSrc, "", "The Bech32 address of the source validator")
	fsRedelegation.String(FlagAddressValidatorDst, "", "The Bech32 address of the destination validator")
}
