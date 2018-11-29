package bank

import (
	"fmt"
)

// TODO remove some of these prefixes once have working multistore
// Key for getting a the next available proposalID from the store
var (
	KeyDelimiter                = []byte(":")
	KeyNextProposalID           = []byte("newProposalID")
	PrefixActiveProposalQueue   = []byte("activeProposalQueue")
	PrefixInactiveProposalQueue = []byte("inactiveProposalQueue")
)

// Key for getting a denom's totalsupply from the denom metadata store
func KeySupply(denom string) []byte {
	return []byte(fmt.Sprintf("supply:%s", denom))
}

// Key for getting a denom's decimals from the denom metadata store
func KeyDecimals(denom string) []byte {
	return []byte(fmt.Sprintf("decimals:%s", denom))
}
