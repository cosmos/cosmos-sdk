package cli

import (
	flag "github.com/spf13/pflag"
)

// nolint
const (
	FlagAddressProposer = "address-proposer"
	FlagDeposit         = "deposit"

	FlagTitle       = "title"
	FlagDescription = "description"
)

// common flagsets to add to various functions
var (
	fsDetails  = flag.NewFlagSet("", flag.ContinueOnError)
	fsAmount   = flag.NewFlagSet("", flag.ContinueOnError)
	fsProposer = flag.NewFlagSet("", flag.ContinueOnError)
)

// TODO flag for selecting a custom window of time to vote

func init() {
	fsDetails.String(FlagTitle, "", "Title of the proposal")
	fsDetails.String(FlagDescription, "", "Description of the proposal")
	fsAmount.String(FlagDeposit, "1steak", "Amount of coins to deposit on the proposal")
	fsProposer.String(FlagAddressProposer, "", "Address of the proposer")
}
