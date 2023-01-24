package cli

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/quarantine"
	"github.com/tendermint/tendermint/crypto"
)

// exampleAddr creates a consistent example address from the given name string.
func exampleAddr(name string) sdk.AccAddress {
	// The correct HRP may or may not be set yet.
	// So, in order for these example addresses to use the chain's HRP, this just creates the AccAddress,
	// allowing the .String() conversion to happen later, hopefully after the HRP is set.
	return sdk.AccAddress(crypto.AddressHash([]byte(name)))
}

var (
	// exampleAddr1 is a random (but static) address to use in examples.
	exampleAddr1 = exampleAddr("exampleAddr1")
	// exampleAddr2 is another random (but static) address to use in examples.
	exampleAddr2 = exampleAddr("exampleAddr2")
	// exampleAddr3 is third random (but static) address to use in examples.
	exampleAddr3 = exampleAddr("exampleAddr3")
)

// validateAddress checks to make sure the provided addr is a valid Bech32 address string.
// If it is invalid, "" is returned with an error that includes the argName.
// If it is valid, the addr is returned without an error.
//
// This validation is (hopefully) already done by the node, but it's more
// user-friendly to also do it here, before a request is actually sent.
func validateAddress(addr string, argName string) (string, error) {
	if _, err := sdk.AccAddressFromBech32(addr); err != nil {
		return "", sdkerrors.ErrInvalidAddress.Wrapf("invalid %s: %v", argName, err)
	}
	return addr, nil
}

// ParseAutoResponseUpdatesFromArgs parses the args to extract the desired AutoResponseUpdate entries.
// The args should be the entire args list. Parsing of the auto-response updates args will start at startIndex.
func ParseAutoResponseUpdatesFromArgs(args []string, startIndex int) ([]*quarantine.AutoResponseUpdate, error) {
	iLastArg := len(args) - 1 - startIndex // index of the last arg.
	arArgCount := 0                        // a count of arguments that have been auto-responses.
	arAddrCount := 0                       // a count of from_addresses provided for the most recent auto-response.
	var lastArArg string                   // the actual arg string of the last auto-response arg.
	var ar quarantine.AutoResponse         // the current auto-response.

	var rv []*quarantine.AutoResponseUpdate

	for i, arg := range args[startIndex:] {
		newAr, isAr := ParseAutoResponseArg(arg)
		// first arg must be an auto-response.
		if i == 0 && !isAr {
			return nil, fmt.Errorf("invalid arg %d: invalid auto-response: %q", i+startIndex+1, arg)
		}
		if isAr {
			// If not the first arg, there must be at least one address for the previous auto-response.
			if i > 0 && arAddrCount == 0 {
				return nil, fmt.Errorf("invalid arg %d: no from_address args provided after auto-response %d: %q", i+startIndex+1, arArgCount, lastArArg)
			}
			// The last argument cannot be an auto-response either.
			if i == iLastArg {
				// Slightly different message on purpose. Makes it easier to track down the source of an error.
				return nil, fmt.Errorf("invalid arg %d: last arg cannot be an auto-response, got: %q", i+startIndex+1, arg)
			}
			arArgCount += 1
			ar = newAr
			lastArArg = arg
			arAddrCount = 0
		} else {
			arAddrCount += 1
			fromAddr, err := validateAddress(arg, "from_address")
			if err != nil {
				return nil, fmt.Errorf("unknown arg %d %q: auto-response %d %q: from_address %d: %w", i+startIndex+1, arg, arArgCount, lastArArg, arAddrCount, err)
			}
			rv = append(rv, &quarantine.AutoResponseUpdate{
				FromAddress: fromAddr,
				Response:    ar,
			})
		}
	}

	return rv, nil
}

// ParseAutoResponseArg converts the provided arg to an AutoResponse enum entry.
// The bool return value is true if parsing was successful.
func ParseAutoResponseArg(arg string) (quarantine.AutoResponse, bool) {
	switch strings.ToLower(arg) {
	case "accept", "a", "1":
		return quarantine.AUTO_RESPONSE_ACCEPT, true
	case "decline", "d", "2":
		return quarantine.AUTO_RESPONSE_DECLINE, true
	case "unspecified", "u", "off", "o", "0":
		return quarantine.AUTO_RESPONSE_UNSPECIFIED, true
	default:
		return quarantine.AUTO_RESPONSE_UNSPECIFIED, false
	}
}
