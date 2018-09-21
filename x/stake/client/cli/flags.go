package cli

import (
	flag "github.com/spf13/pflag"

	"github.com/cosmos/cosmos-sdk/x/stake/types"
)

// nolint
const (
	FlagAddressDelegator    = "address-delegator"
	FlagAddressValidator    = "validator"
	FlagAddressValidatorSrc = "addr-validator-source"
	FlagAddressValidatorDst = "addr-validator-dest"
	FlagPubKey              = "pubkey"
	FlagAmount              = "amount"
	FlagSharesAmount        = "shares-amount"
	FlagSharesPercent       = "shares-percent"

	FlagMoniker  = "moniker"
	FlagIdentity = "identity"
	FlagWebsite  = "website"
	FlagDetails  = "details"

	FlagCommissionRate          = "commission-rate"
	FlagCommissionMaxRate       = "commission-max-rate"
	FlagCommissionMaxChangeRate = "commission-max-change-rate"
)

// common flagsets to add to various functions
var (
	fsPk                = flag.NewFlagSet("", flag.ContinueOnError)
	fsAmount            = flag.NewFlagSet("", flag.ContinueOnError)
	fsShares            = flag.NewFlagSet("", flag.ContinueOnError)
	fsDescriptionCreate = flag.NewFlagSet("", flag.ContinueOnError)
	fsCommissionCreate  = flag.NewFlagSet("", flag.ContinueOnError)
	fsCommissionUpdate  = flag.NewFlagSet("", flag.ContinueOnError)
	fsDescriptionEdit   = flag.NewFlagSet("", flag.ContinueOnError)
	fsValidator         = flag.NewFlagSet("", flag.ContinueOnError)
	fsDelegator         = flag.NewFlagSet("", flag.ContinueOnError)
	fsRedelegation      = flag.NewFlagSet("", flag.ContinueOnError)
)

func init() {
	fsPk.String(FlagPubKey, "", "Go-Amino encoded hex PubKey of the validator. For Ed25519 the go-amino prepend hex is 1624de6220")
	fsAmount.String(FlagAmount, "", "Amount of coins to bond")
	fsShares.String(FlagSharesAmount, "", "Amount of source-shares to either unbond or redelegate as a positive integer or decimal")
	fsShares.String(FlagSharesPercent, "", "Percent of source-shares to either unbond or redelegate as a positive integer or decimal >0 and <=1")
	fsDescriptionCreate.String(FlagMoniker, "", "validator name")
	fsDescriptionCreate.String(FlagIdentity, "", "optional identity signature (ex. UPort or Keybase)")
	fsDescriptionCreate.String(FlagWebsite, "", "optional website")
	fsDescriptionCreate.String(FlagDetails, "", "optional details")
	fsCommissionUpdate.String(FlagCommissionRate, "", "The new commission rate percentage")
	fsCommissionCreate.String(FlagCommissionRate, "", "The initial commission rate percentage")
	fsCommissionCreate.String(FlagCommissionMaxRate, "", "The maximum commission rate percentage")
	fsCommissionCreate.String(FlagCommissionMaxChangeRate, "", "The maximum commission change rate percentage (per day)")
	fsDescriptionEdit.String(FlagMoniker, types.DoNotModifyDesc, "validator name")
	fsDescriptionEdit.String(FlagIdentity, types.DoNotModifyDesc, "optional identity signature (ex. UPort or Keybase)")
	fsDescriptionEdit.String(FlagWebsite, types.DoNotModifyDesc, "optional website")
	fsDescriptionEdit.String(FlagDetails, types.DoNotModifyDesc, "optional details")
	fsValidator.String(FlagAddressValidator, "", "hex address of the validator")
	fsDelegator.String(FlagAddressDelegator, "", "hex address of the delegator")
	fsRedelegation.String(FlagAddressValidatorSrc, "", "hex address of the source validator")
	fsRedelegation.String(FlagAddressValidatorDst, "", "hex address of the destination validator")
}
