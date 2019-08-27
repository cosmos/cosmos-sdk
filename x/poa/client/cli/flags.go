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

	FlagMoniker         = "moniker"
	FlagIdentity        = "identity"
	FlagWebsite         = "website"
	FlagSecurityContact = "security-contact"
	FlagDetails         = "details"

	FlagGenesisFormat = "genesis-format"
	FlagNodeID        = "node-id"
	FlagIP            = "ip"
)

// common flagsets to add to various functions
var (
	FsPk                = flag.NewFlagSet("", flag.ContinueOnError)
	fsDescriptionCreate = flag.NewFlagSet("", flag.ContinueOnError)
	fsDescriptionEdit   = flag.NewFlagSet("", flag.ContinueOnError)
	fsValidator         = flag.NewFlagSet("", flag.ContinueOnError)
)

func init() {
	FsPk.String(FlagPubKey, "", "The Bech32 encoded PubKey of the validator")
	fsDescriptionCreate.String(FlagMoniker, "", "The validator's name")
	fsDescriptionCreate.String(FlagIdentity, "", "The optional identity signature (ex. UPort or Keybase)")
	fsDescriptionCreate.String(FlagWebsite, "", "The validator's (optional) website")
	fsDescriptionCreate.String(FlagSecurityContact, "", "The validator's (optional) security contact email")
	fsDescriptionCreate.String(FlagDetails, "", "The validator's (optional) details")
	fsDescriptionEdit.String(FlagMoniker, types.DoNotModifyDesc, "The validator's name")
	fsDescriptionEdit.String(FlagIdentity, types.DoNotModifyDesc, "The (optional) identity signature (ex. UPort or Keybase)")
	fsDescriptionEdit.String(FlagWebsite, types.DoNotModifyDesc, "The validator's (optional) website")
	fsDescriptionEdit.String(FlagSecurityContact, types.DoNotModifyDesc, "The validator's (optional) security contact email")
	fsDescriptionEdit.String(FlagDetails, types.DoNotModifyDesc, "The validator's (optional) details")
	fsValidator.String(FlagAddressValidator, "", "The Bech32 address of the validator")
}
